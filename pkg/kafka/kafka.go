package kafka

import (
	"context"
	"fmt"
	"time"

	sourcesdk "github.com/numaproj/numaflow-go/pkg/sourcer"
	"go.uber.org/zap"

	"github.com/numaproj-contrib/kafka-source-go/pkg/config"
)

type kafkaSource struct {
	logger *zap.Logger
}

type Option func(*kafkaSource) error

// WithLogger is used to return logger information
func WithLogger(l *zap.Logger) Option {
	return func(o *kafkaSource) error {
		o.logger = l
		return nil
	}
}

func New(c *config.Config, opts ...Option) (*kafkaSource, error) {
	n := &kafkaSource{}
	for _, o := range opts {
		if err := o(n); err != nil {
			return nil, err
		}
	}
	if n.logger == nil {
		var err error
		n.logger, err = zap.NewDevelopment()
		if err != nil {
			return nil, fmt.Errorf("failed to create logger, %w", err)
		}
	}

	n.logger.Info("Connecting to Kafka Service...")
	n.logger.Info("Kafka source server started")
	return n, nil
}

// Pending returns the number of pending records.
func (n *kafkaSource) Pending(_ context.Context) int64 {
	// Pending is not supported for NATS for now, returning -1 to indicate pending is not available.
	return -1
}

func (n *kafkaSource) Read(_ context.Context, readRequest sourcesdk.ReadRequest, messageCh chan<- sourcesdk.Message) {
	// Handle the timeout specification in the read request.
	ctx, cancel := context.WithTimeout(context.Background(), readRequest.TimeOut())
	defer cancel()

	// Read the data from the source and send the data to the message channel.
	for i := 0; uint64(i) < readRequest.Count(); i++ {
		select {
		case <-ctx.Done():
			// If the context is done, the read request is timed out.
			return
		default:
			// Otherwise, we read the data from the source and send the data to the message channel.
			messageCh <- sourcesdk.NewMessage(
				[]byte("test-payload"),
				sourcesdk.NewOffset([]byte("test-offset"), "0"),
				time.Now())
		}
	}
}

// Ack acknowledges the data from the source.
func (n *kafkaSource) Ack(_ context.Context, request sourcesdk.AckRequest) {
	// Ack is a no-op for the NATS source.
}

func (n *kafkaSource) Close() error {
	return nil
}
