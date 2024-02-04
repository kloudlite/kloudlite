package mocks

import (
	context "context"
	types "k8s.io/apimachinery/pkg/types"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type ClientCallerInfo struct {
	Args []any
}

type Client struct {
	Calls              map[string][]ClientCallerInfo
	MockApplyYAML      func(ctx context.Context, yamls ...[]byte) error
	MockDeleteYAML     func(ctx context.Context, yamls ...[]byte) error
	MockGet            func(ctx context.Context, nn types.NamespacedName, obj client.Object) error
	MockValidateObject func(ctx context.Context, obj client.Object) error
}

func (m *Client) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]ClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], ClientCallerInfo{Args: args})
}

func (cMock *Client) ApplyYAML(ctx context.Context, yamls ...[]byte) error {
	if cMock.MockApplyYAML != nil {
		cMock.registerCall("ApplyYAML", ctx, yamls)
		return cMock.MockApplyYAML(ctx, yamls...)
	}
	panic("Client: method 'ApplyYAML' not implemented, yet")
}

func (cMock *Client) DeleteYAML(ctx context.Context, yamls ...[]byte) error {
	if cMock.MockDeleteYAML != nil {
		cMock.registerCall("DeleteYAML", ctx, yamls)
		return cMock.MockDeleteYAML(ctx, yamls...)
	}
	panic("Client: method 'DeleteYAML' not implemented, yet")
}

func (cMock *Client) Get(ctx context.Context, nn types.NamespacedName, obj client.Object) error {
	if cMock.MockGet != nil {
		cMock.registerCall("Get", ctx, nn, obj)
		return cMock.MockGet(ctx, nn, obj)
	}
	panic("Client: method 'Get' not implemented, yet")
}

func (cMock *Client) ValidateObject(ctx context.Context, obj client.Object) error {
	if cMock.MockValidateObject != nil {
		cMock.registerCall("ValidateObject", ctx, obj)
		return cMock.MockValidateObject(ctx, obj)
	}
	panic("Client: method 'ValidateObject' not implemented, yet")
}

func NewClient() *Client {
	return &Client{}
}
