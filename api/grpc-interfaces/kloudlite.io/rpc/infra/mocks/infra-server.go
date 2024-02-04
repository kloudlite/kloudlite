package mocks

import (
	context "context"
	infra "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/infra"
)

type InfraServerCallerInfo struct {
	Args []any
}

type InfraServer struct {
	Calls                 map[string][]InfraServerCallerInfo
	MockGetResourceOutput func(ka context.Context, kb *infra.GetInput) (*infra.Output, error)
}

func (m *InfraServer) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]InfraServerCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], InfraServerCallerInfo{Args: args})
}

func (iMock *InfraServer) GetResourceOutput(ka context.Context, kb *infra.GetInput) (*infra.Output, error) {
	if iMock.MockGetResourceOutput != nil {
		iMock.registerCall("GetResourceOutput", ka, kb)
		return iMock.MockGetResourceOutput(ka, kb)
	}
	panic("method 'GetResourceOutput' not implemented, yet")
}

func NewInfraServer() *InfraServer {
	return &InfraServer{}
}
