package mocks

import (
	context "context"
	grpc "google.golang.org/grpc"
	infra "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/infra"
)

type InfraClientCallerInfo struct {
	Args []any
}

type InfraClient struct {
	Calls                 map[string][]InfraClientCallerInfo
	MockGetResourceOutput func(ctx context.Context, in *infra.GetInput, opts ...grpc.CallOption) (*infra.Output, error)
}

func (m *InfraClient) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]InfraClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], InfraClientCallerInfo{Args: args})
}

func (iMock *InfraClient) GetResourceOutput(ctx context.Context, in *infra.GetInput, opts ...grpc.CallOption) (*infra.Output, error) {
	if iMock.MockGetResourceOutput != nil {
		iMock.registerCall("GetResourceOutput", ctx, in, opts)
		return iMock.MockGetResourceOutput(ctx, in, opts...)
	}
	panic("method 'GetResourceOutput' not implemented, yet")
}

func NewInfraClient() *InfraClient {
	return &InfraClient{}
}
