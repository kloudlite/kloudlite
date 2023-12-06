package mocks

import (
	context "context"
	kafka "kloudlite.io/pkg/kafka"
)

type ConsumerCallerInfo struct {
	Args []any
}

type Consumer struct {
	Calls                map[string][]ConsumerCallerInfo
	MockClose            func() error
	MockLifecycleOnStart func(ctx context.Context) error
	MockLifecycleOnStop  func(ctx context.Context) error
	MockPing             func(ctx context.Context) error
	MockStartConsuming   func(readerMsg kafka.ReaderFunc)
	MockStopConsuming    func()
}

func (m *Consumer) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]ConsumerCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], ConsumerCallerInfo{Args: args})
}

func (cMock *Consumer) Close() error {
	if cMock.MockClose != nil {
		cMock.registerCall("Close")
		return cMock.MockClose()
	}
	panic("Consumer: method 'Close' not implemented, yet")
}

func (cMock *Consumer) LifecycleOnStart(ctx context.Context) error {
	if cMock.MockLifecycleOnStart != nil {
		cMock.registerCall("LifecycleOnStart", ctx)
		return cMock.MockLifecycleOnStart(ctx)
	}
	panic("Consumer: method 'LifecycleOnStart' not implemented, yet")
}

func (cMock *Consumer) LifecycleOnStop(ctx context.Context) error {
	if cMock.MockLifecycleOnStop != nil {
		cMock.registerCall("LifecycleOnStop", ctx)
		return cMock.MockLifecycleOnStop(ctx)
	}
	panic("Consumer: method 'LifecycleOnStop' not implemented, yet")
}

func (cMock *Consumer) Ping(ctx context.Context) error {
	if cMock.MockPing != nil {
		cMock.registerCall("Ping", ctx)
		return cMock.MockPing(ctx)
	}
	panic("Consumer: method 'Ping' not implemented, yet")
}

func (cMock *Consumer) StartConsuming(readerMsg kafka.ReaderFunc) {
	if cMock.MockStartConsuming != nil {
		cMock.registerCall("StartConsuming", readerMsg)
		cMock.MockStartConsuming(readerMsg)
	}
	panic("Consumer: method 'StartConsuming' not implemented, yet")
}

func (cMock *Consumer) StopConsuming() {
	if cMock.MockStopConsuming != nil {
		cMock.registerCall("StopConsuming")
		cMock.MockStopConsuming()
	}
	panic("Consumer: method 'StopConsuming' not implemented, yet")
}

func NewConsumer() *Consumer {
	return &Consumer{}
}
