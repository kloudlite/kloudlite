package mocks

import (
	context1 "context"
	agent2 "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/agent"
)

type KubeAgentServerCallerInfo struct {
	Args []any
}

type KubeAgentServer struct {
	Calls         map[string][]KubeAgentServerCallerInfo
	MockKubeApply func(ka context1.Context, kb *agent2.PayloadIn) (*agent2.PayloadOut, error)
}

func (m *KubeAgentServer) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]KubeAgentServerCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], KubeAgentServerCallerInfo{Args: args})
}

func (k *KubeAgentServer) KubeApply(ka context1.Context, kb *agent2.PayloadIn) (*agent2.PayloadOut, error) {
	if k.MockKubeApply != nil {
		k.registerCall("KubeApply", ka, kb)
		return k.MockKubeApply(ka, kb)
	}
	panic("not implemented, yet")
}

func NewKubeAgentServer() *KubeAgentServer {
	return &KubeAgentServer{}
}
