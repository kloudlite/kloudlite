package redpanda

import (
	"time"
)

type ReaderFunc func(msg *KafkaMessage) error

type KafkaMessage struct {
	Key        []byte
	Value      []byte
	Timestamp  time.Time
	Topic      string
	Partition  int32
	ProducerId int64
	Offset     int64
}
