package types

import "time"

type NatsJetstreamProduceMsg struct {
	Subject string
	Payload []byte
}

type ProduceMsg struct {
	Subject string
	Payload []byte
}

type ProducerOutput struct{}

type ConsumeMsg struct {
	Subject   string
	Timestamp time.Time
	Payload   []byte
}

type ConsumerOutput struct{}

type ErrShouldRetry struct {
	Err error
}

func (e ErrShouldRetry) Error() string {
	return "error occurred"
}

type ConsumeOpts struct {
	/* OnError is called when an error occurs while consuming a message.
	   If OnError returns an error, then
	      the consumer will not commit that message, so that it will be queued again.
	    Otherwise,
	      the consumer will commit the message, so that it will not be consumed again
	*/
	OnError func(err error) error
}
