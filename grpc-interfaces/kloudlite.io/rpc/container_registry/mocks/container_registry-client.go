package mocks

import (
	context1 "context"
	grpc3 "google.golang.org/grpc"
	container_registry2 "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/container_registry"
)

type ContainerRegistryClientCallerInfo struct {
	Args []any
}

type ContainerRegistryClient struct {
	Calls                       map[string][]ContainerRegistryClientCallerInfo
	MockCreateProjectForAccount func(ctx context1.Context, in *container_registry2.CreateProjectIn, opts ...grpc3.CallOption) (*container_registry2.CreateProjectOut, error)
	MockGetSvcCredentials       func(ctx context1.Context, in *container_registry2.GetSvcCredentialsIn, opts ...grpc3.CallOption) (*container_registry2.GetSvcCredentialsOut, error)
}

func (m *ContainerRegistryClient) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]ContainerRegistryClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], ContainerRegistryClientCallerInfo{Args: args})
}

func (c *ContainerRegistryClient) CreateProjectForAccount(ctx context1.Context, in *container_registry2.CreateProjectIn, opts ...grpc3.CallOption) (*container_registry2.CreateProjectOut, error) {
	if c.MockCreateProjectForAccount != nil {
		c.registerCall("CreateProjectForAccount", ctx, in, opts)
		return c.MockCreateProjectForAccount(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (c *ContainerRegistryClient) GetSvcCredentials(ctx context1.Context, in *container_registry2.GetSvcCredentialsIn, opts ...grpc3.CallOption) (*container_registry2.GetSvcCredentialsOut, error) {
	if c.MockGetSvcCredentials != nil {
		c.registerCall("GetSvcCredentials", ctx, in, opts)
		return c.MockGetSvcCredentials(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func NewContainerRegistryClient() *ContainerRegistryClient {
	return &ContainerRegistryClient{}
}
