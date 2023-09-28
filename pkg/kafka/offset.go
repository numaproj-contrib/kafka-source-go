package kafka

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/IBM/sarama"
	sourcesdk "github.com/numaproj/numaflow-go/pkg/sourcer"
)

// To translate a kafka offset to a source offset, we need to store the topic name, partition index and offset value
// in the source offset value. We use the following format:
// <topic>*<offset>
// For example, if the topic name is "test-topic", the partition index is 0 and the offset value is 123, the source offset value will be:
// test-topic*123
// We use "*" as the separator because it is not allowed in a topic name.
// The topic name can be up to 255 characters in length, and can include the following characters: a-z, A-Z, 0-9, . (dot), _ (underscore), and - (dash).
const offsetSeparator = "*"

type KafkaOffset struct {
	offset       int64
	partitionIdx int32
	topic        string
}

func (k *KafkaOffset) String() string {
	return fmt.Sprintf("%s:%d:%d", k.topic, k.offset, k.partitionIdx)
}

func (k *KafkaOffset) Sequence() (int64, error) {
	return k.offset, nil
}

func (k *KafkaOffset) PartitionIdx() int32 {
	return k.partitionIdx
}

func (k *KafkaOffset) Topic() string {
	return k.topic
}

// GenerateSourceSdkOffset generates a source offset from a kafka message
func GenerateSourceSdkOffset(m *sarama.ConsumerMessage) sourcesdk.Offset {
	k := &KafkaOffset{
		offset:       m.Offset,
		partitionIdx: m.Partition,
		topic:        m.Topic,
	}
	return k.ToSourceOffset()
}

func ToKafkaOffset(o *sourcesdk.Offset) (*KafkaOffset, error) {
	strVal := string(o.Value())
	strs := strings.Split(strVal, offsetSeparator)
	if len(strs) != 2 {
		return nil, fmt.Errorf("invalid offset value %s, the value cannot be divided to topic and offset", strVal)
	}
	var offset int
	var err error
	var partitionIdx int
	if offset, err = strconv.Atoi(strs[1]); err != nil {
		return nil, fmt.Errorf("invalid offset value %s", strs[1])
	}
	if partitionIdx, err = strconv.Atoi(o.PartitionId()); err != nil {
		return nil, fmt.Errorf("invalid partition id %s", o.PartitionId())
	}
	return &KafkaOffset{
		topic:        strs[0],
		offset:       int64(offset),
		partitionIdx: int32(partitionIdx),
	}, nil
}

func (k *KafkaOffset) ToSourceOffset() sourcesdk.Offset {
	value := []byte(fmt.Sprintf("%s%s%d", k.topic, offsetSeparator, k.offset))
	return sourcesdk.NewOffset(value, strconv.Itoa(int(k.partitionIdx)))
}
