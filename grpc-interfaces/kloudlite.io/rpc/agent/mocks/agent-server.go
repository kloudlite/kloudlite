package mocks

import (
	context "context"
	agent "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/agent"
)

type KubeAgentServerCallerInfo struct {
	Args []any
}

type KubeAgentServer struct {
	Calls         map[string][]KubeAgentServerCallerInfo
	MockKubeApply func(ka context.Context, kb *agent.PayloadIn) (*agent.PayloadOut, error)
}

func (m *KubeAgentServer) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]KubeAgentServerCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], KubeAgentServerCallerInfo{Args: args})
}

func (kMock *KubeAgentServer) KubeApply(ka context.Context, kb *agent.PayloadIn) (*agent.PayloadOut, error) {
	if kMock.MockKubeApply != nil {
		kMock.registerCall("KubeApply", ka, kb)
		return kMock.MockKubeApply(ka, kb)
	}
	panic("method 'KubeApply' not implemented, yet")
}

func NewKubeAgentServer() *KubeAgentServer {
	return &KubeAgentServer{}
}
