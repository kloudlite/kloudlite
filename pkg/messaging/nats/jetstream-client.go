package nats

import (
	"github.com/nats-io/nats.go/jetstream"
)

type JetstreamClient struct {
	js jetstream.JetStream
}

func NewJetstreamClient(nc *Client) (*JetstreamClient, error) {
	js, err := jetstream.New(nc.Conn)
	if err != nil {
		return nil, err
	}

	return &JetstreamClient{
		js: js,
	}, nil
}
