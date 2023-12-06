package types

import "fmt"

type NatsJetstreamProduceMsg struct {
	Subject string
	Payload []byte
}

type ProduceMsg struct {
	NatsJetstreamMsg *NatsJetstreamProduceMsg
}

type NatsJetstreamProducerOutput struct{}

type ProducerOutput struct {
	NatsJetstream *NatsJetstreamProducerOutput
}

type NatsJetstreamConsumeMsg struct {
	Payload []byte
}

type ConsumeMsg struct {
	NatsJetstreamMsg *NatsJetstreamConsumeMsg
}

type ConsumerOutput struct{}

type ErrShouldRetry struct {
	Err error
}

func (e ErrShouldRetry) Error() string {
	return fmt.Sprintf("error occurred")
}
