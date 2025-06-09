package mocks

import (
	context "context"
	container_registry "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/container_registry"
	grpc "google.golang.org/grpc"
)

type ContainerRegistryClientCallerInfo struct {
	Args []any
}

type ContainerRegistryClient struct {
	Calls                       map[string][]ContainerRegistryClientCallerInfo
	MockCreateProjectForAccount func(ctx context.Context, in *container_registry.CreateProjectIn, opts ...grpc.CallOption) (*container_registry.CreateProjectOut, error)
	MockGetSvcCredentials       func(ctx context.Context, in *container_registry.GetSvcCredentialsIn, opts ...grpc.CallOption) (*container_registry.GetSvcCredentialsOut, error)
}

func (m *ContainerRegistryClient) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]ContainerRegistryClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], ContainerRegistryClientCallerInfo{Args: args})
}

func (cMock *ContainerRegistryClient) CreateProjectForAccount(ctx context.Context, in *container_registry.CreateProjectIn, opts ...grpc.CallOption) (*container_registry.CreateProjectOut, error) {
	if cMock.MockCreateProjectForAccount != nil {
		cMock.registerCall("CreateProjectForAccount", ctx, in, opts)
		return cMock.MockCreateProjectForAccount(ctx, in, opts...)
	}
	panic("method 'CreateProjectForAccount' not implemented, yet")
}

func (cMock *ContainerRegistryClient) GetSvcCredentials(ctx context.Context, in *container_registry.GetSvcCredentialsIn, opts ...grpc.CallOption) (*container_registry.GetSvcCredentialsOut, error) {
	if cMock.MockGetSvcCredentials != nil {
		cMock.registerCall("GetSvcCredentials", ctx, in, opts)
		return cMock.MockGetSvcCredentials(ctx, in, opts...)
	}
	panic("method 'GetSvcCredentials' not implemented, yet")
}

func NewContainerRegistryClient() *ContainerRegistryClient {
	return &ContainerRegistryClient{}
}
