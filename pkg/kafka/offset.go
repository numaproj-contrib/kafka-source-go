package kafka

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/IBM/sarama"
	sourcesdk "github.com/numaproj/numaflow-go/pkg/sourcer"
)

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

// GenerateSourceOffset generates a source offset from a kafka message
func GenerateSourceOffset(m *sarama.ConsumerMessage) sourcesdk.Offset {
	k := &KafkaOffset{
		offset:       m.Offset,
		partitionIdx: m.Partition,
		topic:        m.Topic,
	}
	return k.ToSourceOffset()
}

// TODO - unit test it
func ToKafkaOffset(o *sourcesdk.Offset) (*KafkaOffset, error) {
	strVal := string(o.Value())
	strs := strings.Split(strVal, "*")
	if len(strs) != 2 {
		return nil, fmt.Errorf("invalid offset value %s", strVal)
	}
	var offset int
	var err error
	var partitionIdx int
	if offset, err = strconv.Atoi(strs[1]); err != nil {
		return nil, fmt.Errorf("invalid offset value %s", strVal)
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
	value := []byte(fmt.Sprintf("%s*%d", k.topic, k.offset))
	return sourcesdk.NewOffset(value, strconv.Itoa(int(k.partitionIdx)))
}
