package mocks

import (
	context1 "context"
	container_registry2 "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/container_registry"
)

type ContainerRegistryServerCallerInfo struct {
	Args []any
}

type ContainerRegistryServer struct {
	Calls                       map[string][]ContainerRegistryServerCallerInfo
	MockCreateProjectForAccount func(ka context1.Context, kb *container_registry2.CreateProjectIn) (*container_registry2.CreateProjectOut, error)
	MockGetSvcCredentials       func(kc context1.Context, kd *container_registry2.GetSvcCredentialsIn) (*container_registry2.GetSvcCredentialsOut, error)
}

func (m *ContainerRegistryServer) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]ContainerRegistryServerCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], ContainerRegistryServerCallerInfo{Args: args})
}

func (c *ContainerRegistryServer) CreateProjectForAccount(ka context1.Context, kb *container_registry2.CreateProjectIn) (*container_registry2.CreateProjectOut, error) {
	if c.MockCreateProjectForAccount != nil {
		c.registerCall("CreateProjectForAccount", ka, kb)
		return c.MockCreateProjectForAccount(ka, kb)
	}
	panic("not implemented, yet")
}

func (c *ContainerRegistryServer) GetSvcCredentials(kc context1.Context, kd *container_registry2.GetSvcCredentialsIn) (*container_registry2.GetSvcCredentialsOut, error) {
	if c.MockGetSvcCredentials != nil {
		c.registerCall("GetSvcCredentials", kc, kd)
		return c.MockGetSvcCredentials(kc, kd)
	}
	panic("not implemented, yet")
}

func NewContainerRegistryServer() *ContainerRegistryServer {
	return &ContainerRegistryServer{}
}
