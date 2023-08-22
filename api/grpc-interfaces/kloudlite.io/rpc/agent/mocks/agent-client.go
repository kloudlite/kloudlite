package mocks

import (
	context1 "context"
	grpc3 "google.golang.org/grpc"
	agent2 "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/agent"
)

type KubeAgentClientCallerInfo struct {
	Args []any
}

type KubeAgentClient struct {
	Calls         map[string][]KubeAgentClientCallerInfo
	MockKubeApply func(ctx context1.Context, in *agent2.PayloadIn, opts ...grpc3.CallOption) (*agent2.PayloadOut, error)
}

func (m *KubeAgentClient) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]KubeAgentClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], KubeAgentClientCallerInfo{Args: args})
}

func (k *KubeAgentClient) KubeApply(ctx context1.Context, in *agent2.PayloadIn, opts ...grpc3.CallOption) (*agent2.PayloadOut, error) {
	if k.MockKubeApply != nil {
		k.registerCall("KubeApply", ctx, in, opts)
		return k.MockKubeApply(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func NewKubeAgentClient() *KubeAgentClient {
	return &KubeAgentClient{}
}
