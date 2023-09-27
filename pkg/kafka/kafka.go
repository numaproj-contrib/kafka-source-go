package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	sourcesdk "github.com/numaproj/numaflow-go/pkg/sourcer"
	"go.uber.org/zap"

	"github.com/numaproj-contrib/kafka-source-go/pkg/config"
	"github.com/numaproj-contrib/kafka-source-go/pkg/utils"
)

type kafkaSource struct {
	consumerGrpName string
	topic           string
	brokers         []string

	// sarama config for kafka consumer group
	config *sarama.Config

	// handler for a kafka consumer group
	handler *consumerHandler
	// size of the buffer that holds consumed but yet to be forwarded messages
	handlerBuffer int
	// client used to calculate pending messages
	adminClient sarama.ClusterAdmin
	// sarama client
	saramaClient sarama.Client

	volumeReader utils.VolumeReader

	// context cancel function
	cancelFn context.CancelFunc
	// lifecycle context
	lifecycleCtx context.Context

	// channel to indicate that we are done
	stopCh chan struct{}

	logger *zap.Logger
}

type Option func(*kafkaSource) error

// WithLogger is used to return logger information
func WithLogger(l *zap.Logger) Option {
	return func(k *kafkaSource) error {
		k.logger = l
		return nil
	}
}

// WithBufferSize is used to return size of message channel information
func WithBufferSize(s int) Option {
	return func(k *kafkaSource) error {
		k.handlerBuffer = s
		return nil
	}
}

func New(c *config.Config, opts ...Option) (*kafkaSource, error) {
	k := &kafkaSource{
		topic:           c.Topic,
		brokers:         c.Brokers,
		consumerGrpName: c.ConsumerGroupName,
		handlerBuffer:   100, // default buffer size for kafka reads
	}
	for _, o := range opts {
		if err := o(k); err != nil {
			return nil, err
		}
	}
	if k.logger == nil {
		var err error
		k.logger, err = zap.NewDevelopment()
		if err != nil {
			return nil, fmt.Errorf("failed to create logger, %w", err)
		}
	}
	k.volumeReader = utils.NewKafkaVolumeReader(utils.SecretVolumePath)

	sarama.NewConfig()
	kConfig, err := configFromOpts(c.Config)
	if err != nil {
		return nil, fmt.Errorf("error reading kafka source config, %w", err)
	}

	if t := c.TLS; t != nil {
		kConfig.Net.TLS.Enable = true
		if c, err := utils.GetTLSConfig(t, k.volumeReader); err != nil {
			return nil, err
		} else {
			kConfig.Net.TLS.Config = c
		}
	}
	if s := c.SASL; s != nil {
		if sasl, err := utils.GetSASL(s, k.volumeReader); err != nil {
			return nil, err
		} else {
			kConfig.Net.SASL = *sasl
		}
	}

	sarama.Logger = zap.NewStdLog(k.logger)
	// return errors from the underlying kafka client using the Errors channel
	kConfig.Consumer.Return.Errors = true
	k.config = kConfig

	ctx, cancel := context.WithCancel(context.Background())
	k.cancelFn = cancel
	k.lifecycleCtx = ctx

	k.stopCh = make(chan struct{})
	handler := newConsumerHandler(k.handlerBuffer)
	k.handler = handler

	k.logger.Info("Starting Kafka consumer...")
	go k.Start()
	k.logger.Info("Kafka consumer started.")
	return k, nil
}

func (k *kafkaSource) Start() {
	client, err := sarama.NewClient(k.brokers, k.config)
	if err != nil {
		k.logger.Panic("Failed to create sarama client", zap.Error(err))
	} else {
		k.saramaClient = client
	}
	// Does it require any special privileges to create a cluster admin client?
	adminClient, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		if !client.Closed() {
			_ = client.Close()
		}
		k.logger.Panic("Failed to create sarama cluster admin client", zap.Error(err))
	} else {
		k.adminClient = adminClient
	}

	go k.startConsumer()
	// wait for the consumer to setup.
	<-k.handler.ready
	k.logger.Info("Consumer ready. Starting kafka reader...")
	return
}

// Pending returns the number of pending records.
func (k *kafkaSource) Pending(_ context.Context) int64 {
	if k.adminClient == nil || k.saramaClient == nil {
		return -1
	}
	partitions, err := k.saramaClient.Partitions(k.topic)
	if err != nil {
		return -1
	}
	totalPending := int64(0)
	rep, err := k.adminClient.ListConsumerGroupOffsets(k.consumerGrpName, map[string][]int32{k.topic: partitions})
	if err != nil {
		err := k.refreshAdminClient()
		if err != nil {
			return -1
		}
		return -1
	}
	for _, partition := range partitions {
		block := rep.GetBlock(k.topic, partition)
		if block.Offset == -1 {
			// Note: if there is no offset associated with the partition under the consumer group, offset fetch sets the offset field to -1.
			// This is not an error and usually means that there has been no data published to this particular partition yet.
			// In this case, we can safely skip this partition from the pending calculation.
			continue
		}
		partitionOffset, err := k.saramaClient.GetOffset(k.topic, partition, sarama.OffsetNewest)
		if err != nil {
			return -1
		}
		totalPending += partitionOffset - block.Offset
	}
	return totalPending
}

func (k *kafkaSource) Read(_ context.Context, readRequest sourcesdk.ReadRequest, messageCh chan<- sourcesdk.Message) {
	// Handle the timeout specification in the read request.
	ctx, cancel := context.WithTimeout(context.Background(), readRequest.TimeOut())
	defer cancel()

	// Read the data from the source and send the data to the message channel.
	for i := 0; uint64(i) < readRequest.Count(); i++ {
		select {
		case <-ctx.Done():
			// If the context is done, the read request is timed out.
			return
		case m := <-k.handler.messages:
			// Otherwise, we read the data from the source and send the data to the message channel.
			messageCh <- toSDKMessage(m)
		}
	}
}

// Ack acknowledges the data from the source.
func (k *kafkaSource) Ack(_ context.Context, request sourcesdk.AckRequest) {
	// we want to block the handler from exiting if there are any inflight acks.
	k.handler.inflightacks = make(chan bool)
	defer close(k.handler.inflightacks)

	for _, offset := range request.Offsets() {
		kOffset, err := ToKafkaOffset(&offset)
		if err != nil {
			k.logger.Panic("Unable to convert offset to kafka offset", zap.Error(err))
		}
		topic := kOffset.Topic()

		// we need to mark the offset of the next message to read
		pOffset, err := kOffset.Sequence()
		if err != nil {
			k.logger.Error("Unable to extract partition offset of type int64 from the supplied offset. skipping and continuing", zap.String("supplied-offset", kOffset.String()), zap.Error(err))
			continue
		}
		k.handler.sess.MarkOffset(topic, kOffset.PartitionIdx(), pOffset, "")
	}
}

func (k *kafkaSource) Close() error {
	return nil
}

func configFromOpts(yamlConfig string) (*sarama.Config, error) {
	config, err := utils.GetSaramaConfigFromYAMLString(yamlConfig)
	if err != nil {
		return nil, err
	}
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	return config, nil
}

func (k *kafkaSource) startConsumer() {
	client, err := sarama.NewConsumerGroup(k.brokers, k.consumerGrpName, k.config)
	k.logger.Info("creating NewConsumerGroup", zap.String("topic", k.topic), zap.String("consumerGroupName", k.consumerGrpName), zap.Strings("brokers", k.brokers))
	if err != nil {
		k.logger.Panic("Problem initializing sarama client", zap.Error(err))
	}
	wg := new(sync.WaitGroup)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-k.lifecycleCtx.Done():
				return
			case cErr := <-client.Errors():
				k.logger.Error("Kafka consumer error", zap.Error(cErr))
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop; when a
			// server-side re-balance happens, the consumer session will need to be
			// recreated to get the new claims
			if conErr := client.Consume(k.lifecycleCtx, []string{k.topic}, k.handler); conErr != nil {
				// Panic on errors to let it crash and restart the process
				k.logger.Panic("Kafka consumer failed with error: ", zap.Error(conErr))
			}
			// check if context was cancelled, signaling that the consumer should stop
			if k.lifecycleCtx.Err() != nil {
				return
			}
		}
	}()
	wg.Wait()
	_ = client.Close()
	close(k.stopCh)
}

// refreshAdminClient refreshes the admin client
func (k *kafkaSource) refreshAdminClient() error {
	if _, err := k.saramaClient.RefreshController(); err != nil {
		return fmt.Errorf("failed to refresh controller, %w", err)
	}
	// we are not closing the old admin client because it will close the underlying sarama client as well
	// it is safe to not close the admin client,
	// since we are using the same sarama client, we will not leak any resources(tcp connections)
	admin, err := sarama.NewClusterAdminFromClient(k.saramaClient)
	if err != nil {
		return fmt.Errorf("failed to create new admin client, %w", err)
	}
	k.adminClient = admin
	return nil
}

func toSDKMessage(m *sarama.ConsumerMessage) sourcesdk.Message {
	return sourcesdk.NewMessage(
		m.Value,
		GenerateSourceOffset(m),
		time.Now())
}
