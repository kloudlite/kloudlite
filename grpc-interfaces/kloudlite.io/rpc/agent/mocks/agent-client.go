package mocks

import (
	context "context"
	agent "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/agent"
	grpc "google.golang.org/grpc"
)

type KubeAgentClientCallerInfo struct {
	Args []any
}

type KubeAgentClient struct {
	Calls         map[string][]KubeAgentClientCallerInfo
	MockKubeApply func(ctx context.Context, in *agent.PayloadIn, opts ...grpc.CallOption) (*agent.PayloadOut, error)
}

func (m *KubeAgentClient) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]KubeAgentClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], KubeAgentClientCallerInfo{Args: args})
}

func (kMock *KubeAgentClient) KubeApply(ctx context.Context, in *agent.PayloadIn, opts ...grpc.CallOption) (*agent.PayloadOut, error) {
	if kMock.MockKubeApply != nil {
		kMock.registerCall("KubeApply", ctx, in, opts)
		return kMock.MockKubeApply(ctx, in, opts...)
	}
	panic("method 'KubeApply' not implemented, yet")
}

func NewKubeAgentClient() *KubeAgentClient {
	return &KubeAgentClient{}
}
