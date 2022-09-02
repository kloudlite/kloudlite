package redpanda

import "time"

type Message struct {
	Key        []byte    `json:"key"`
	Value      []byte    `json:"value"`
	Timestamp  time.Time `json:"timestamp"`
	Topic      string    `json:"topic"`
	Partition  int32     `json:"partition"`
	ProducerId int64     `json:"producerId"`
	Offset     int64     `json:"offset"`
}
