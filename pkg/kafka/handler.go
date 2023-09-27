package kafka

import (
	"sync"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"github.com/numaproj-contrib/kafka-source-go/pkg/utils"
)

// consumerHandler struct
type consumerHandler struct {
	inflightacks chan bool
	ready        chan bool
	readycloser  sync.Once
	messages     chan *sarama.ConsumerMessage
	sess         sarama.ConsumerGroupSession
	logger       *zap.SugaredLogger
}

// new handler initializes the channel for passing messages
func newConsumerHandler(readChanSize int) *consumerHandler {
	return &consumerHandler{
		ready:    make(chan bool),
		messages: make(chan *sarama.ConsumerMessage, readChanSize),
		logger:   utils.NewLogger(),
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *consumerHandler) Setup(sess sarama.ConsumerGroupSession) error {
	consumer.sess = sess
	consumer.readycloser.Do(func() {
		close(consumer.ready)
	})
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *consumerHandler) Cleanup(sess sarama.ConsumerGroupSession) error {
	// wait for inflight acks to be completed.
	<-consumer.inflightacks
	sess.Commit()
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/IBM/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			consumer.messages <- msg
		case <-session.Context().Done():
			consumer.logger.Info("context was canceled, stopping consumer claim")
			return nil
		}
	}
}
