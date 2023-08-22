package mocks

import (
	context1 "context"
	grpc3 "google.golang.org/grpc"
	infra2 "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/infra"
)

type InfraClientCallerInfo struct {
	Args []any
}

type InfraClient struct {
	Calls                 map[string][]InfraClientCallerInfo
	MockGetResourceOutput func(ctx context1.Context, in *infra2.GetInput, opts ...grpc3.CallOption) (*infra2.Output, error)
}

func (m *InfraClient) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]InfraClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], InfraClientCallerInfo{Args: args})
}

func (i *InfraClient) GetResourceOutput(ctx context1.Context, in *infra2.GetInput, opts ...grpc3.CallOption) (*infra2.Output, error) {
	if i.MockGetResourceOutput != nil {
		i.registerCall("GetResourceOutput", ctx, in, opts)
		return i.MockGetResourceOutput(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func NewInfraClient() *InfraClient {
	return &InfraClient{}
}
