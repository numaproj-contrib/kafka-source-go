package kafka

import (
	"testing"

	sourcesdk "github.com/numaproj/numaflow-go/pkg/sourcer"
	"github.com/stretchr/testify/assert"
)

func TestTransformation(t *testing.T) {
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
