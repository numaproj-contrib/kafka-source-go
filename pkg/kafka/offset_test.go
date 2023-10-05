package kafka

import (
	"testing"

	"github.com/IBM/sarama"
	sourcesdk "github.com/numaproj/numaflow-go/pkg/sourcer"
	"github.com/stretchr/testify/assert"
)

func TestSdkAndKafkaTransformation(t *testing.T) {
	originalSrcOffset := sourcesdk.NewOffset([]byte("test-topic*100"), "0")
	kafkaOffset, err := ToKafkaOffset(&originalSrcOffset)
	assert.NoError(t, err)
	convertBackToSrc := kafkaOffset.ToSourceOffset()
	assert.Equal(t, originalSrcOffset, convertBackToSrc)

	originalKafkaOffset := &KafkaOffset{
		offset:       100,
		partitionIdx: 0,
		topic:        "test-topic",
	}
	srcOffset := originalKafkaOffset.ToSourceOffset()
	convertBackToKafka, err := ToKafkaOffset(&srcOffset)
	assert.NoError(t, err)
	assert.Equal(t, originalKafkaOffset, convertBackToKafka)
}

func TestGenerateSourceSdkOffset(t *testing.T) {
	m := &sarama.ConsumerMessage{
		Offset:    100,
		Partition: 0,
		Topic:     "test-topic",
	}
	generatedOffset := GenerateSourceSdkOffset(m)
	kafkaOffset, err := ToKafkaOffset(&generatedOffset)
	assert.NoError(t, err)
	assert.Equal(t, int64(100), kafkaOffset.offset)
	assert.Equal(t, int32(0), kafkaOffset.partitionIdx)
	assert.Equal(t, "test-topic", kafkaOffset.topic)
}
