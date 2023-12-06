package mocks

import (
	context "context"
	kafka "kloudlite.io/pkg/kafka"
)

type ProducerCallerInfo struct {
	Args []any
}

type Producer struct {
	Calls                map[string][]ProducerCallerInfo
	MockClose            func()
	MockLifecycleOnStart func(ctx context.Context) error
	MockLifecycleOnStop  func(ctx context.Context) error
	MockPing             func(ctx context.Context) error
	MockProduce          func(ctx context.Context, topic string, value []byte, args kafka.MessageArgs) (*kafka.ProducerOutput, error)
}

func (m *Producer) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]ProducerCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], ProducerCallerInfo{Args: args})
}

func (pMock *Producer) Close() {
	if pMock.MockClose != nil {
		pMock.registerCall("Close")
		pMock.MockClose()
	}
	panic("Producer: method 'Close' not implemented, yet")
}

func (pMock *Producer) LifecycleOnStart(ctx context.Context) error {
	if pMock.MockLifecycleOnStart != nil {
		pMock.registerCall("LifecycleOnStart", ctx)
		return pMock.MockLifecycleOnStart(ctx)
	}
	panic("Producer: method 'LifecycleOnStart' not implemented, yet")
}

func (pMock *Producer) LifecycleOnStop(ctx context.Context) error {
	if pMock.MockLifecycleOnStop != nil {
		pMock.registerCall("LifecycleOnStop", ctx)
		return pMock.MockLifecycleOnStop(ctx)
	}
	panic("Producer: method 'LifecycleOnStop' not implemented, yet")
}

func (pMock *Producer) Ping(ctx context.Context) error {
	if pMock.MockPing != nil {
		pMock.registerCall("Ping", ctx)
		return pMock.MockPing(ctx)
	}
	panic("Producer: method 'Ping' not implemented, yet")
}

func (pMock *Producer) Produce(ctx context.Context, topic string, value []byte, args kafka.MessageArgs) (*kafka.ProducerOutput, error) {
	if pMock.MockProduce != nil {
		pMock.registerCall("Produce", ctx, topic, value, args)
		return pMock.MockProduce(ctx, topic, value, args)
	}
	panic("Producer: method 'Produce' not implemented, yet")
}

func NewProducer() *Producer {
	return &Producer{}
}
