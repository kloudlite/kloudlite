package mocks

import (
	context1 "context"
	infra2 "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/infra"
)

type InfraServerCallerInfo struct {
	Args []any
}

type InfraServer struct {
	Calls                 map[string][]InfraServerCallerInfo
	MockGetResourceOutput func(ka context1.Context, kb *infra2.GetInput) (*infra2.Output, error)
}

func (m *InfraServer) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]InfraServerCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], InfraServerCallerInfo{Args: args})
}

func (i *InfraServer) GetResourceOutput(ka context1.Context, kb *infra2.GetInput) (*infra2.Output, error) {
	if i.MockGetResourceOutput != nil {
		i.registerCall("GetResourceOutput", ka, kb)
		return i.MockGetResourceOutput(ka, kb)
	}
	panic("not implemented, yet")
}

func NewInfraServer() *InfraServer {
	return &InfraServer{}
}
