package mocks

import (
	context "context"
	container_registry "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/container_registry"
)

type ContainerRegistryServerCallerInfo struct {
	Args []any
}

type ContainerRegistryServer struct {
	Calls                       map[string][]ContainerRegistryServerCallerInfo
	MockCreateProjectForAccount func(ka context.Context, kb *container_registry.CreateProjectIn) (*container_registry.CreateProjectOut, error)
	MockGetSvcCredentials       func(kc context.Context, kd *container_registry.GetSvcCredentialsIn) (*container_registry.GetSvcCredentialsOut, error)
}

func (m *ContainerRegistryServer) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]ContainerRegistryServerCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], ContainerRegistryServerCallerInfo{Args: args})
}

func (cMock *ContainerRegistryServer) CreateProjectForAccount(ka context.Context, kb *container_registry.CreateProjectIn) (*container_registry.CreateProjectOut, error) {
	if cMock.MockCreateProjectForAccount != nil {
		cMock.registerCall("CreateProjectForAccount", ka, kb)
		return cMock.MockCreateProjectForAccount(ka, kb)
	}
	panic("method 'CreateProjectForAccount' not implemented, yet")
}

func (cMock *ContainerRegistryServer) GetSvcCredentials(kc context.Context, kd *container_registry.GetSvcCredentialsIn) (*container_registry.GetSvcCredentialsOut, error) {
	if cMock.MockGetSvcCredentials != nil {
		cMock.registerCall("GetSvcCredentials", kc, kd)
		return cMock.MockGetSvcCredentials(kc, kd)
	}
	panic("method 'GetSvcCredentials' not implemented, yet")
}

func NewContainerRegistryServer() *ContainerRegistryServer {
	return &ContainerRegistryServer{}
}
