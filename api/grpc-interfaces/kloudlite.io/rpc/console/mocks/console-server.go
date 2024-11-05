package mocks

import (
	context "context"
	console "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/console"
)

type ConsoleServerCallerInfo struct {
	Args []any
}

type ConsoleServer struct {
	Calls              map[string][]ConsoleServerCallerInfo
	MockGetApp         func(ka context.Context, kb *console.AppIn) (*console.AppOut, error)
	MockGetManagedSvc  func(kc context.Context, kd *console.MSvcIn) (*console.MSvcOut, error)
	MockGetProjectName func(ke context.Context, kf *console.ProjectIn) (*console.ProjectOut, error)
	MockSetupAccount   func(kg context.Context, kh *console.AccountSetupIn) (*console.AccountSetupVoid, error)
}

func (m *ConsoleServer) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]ConsoleServerCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], ConsoleServerCallerInfo{Args: args})
}

func (cMock *ConsoleServer) GetApp(ka context.Context, kb *console.AppIn) (*console.AppOut, error) {
	if cMock.MockGetApp != nil {
		cMock.registerCall("GetApp", ka, kb)
		return cMock.MockGetApp(ka, kb)
	}
	panic("method 'GetApp' not implemented, yet")
}

func (cMock *ConsoleServer) GetManagedSvc(kc context.Context, kd *console.MSvcIn) (*console.MSvcOut, error) {
	if cMock.MockGetManagedSvc != nil {
		cMock.registerCall("GetManagedSvc", kc, kd)
		return cMock.MockGetManagedSvc(kc, kd)
	}
	panic("method 'GetManagedSvc' not implemented, yet")
}

func (cMock *ConsoleServer) GetProjectName(ke context.Context, kf *console.ProjectIn) (*console.ProjectOut, error) {
	if cMock.MockGetProjectName != nil {
		cMock.registerCall("GetProjectName", ke, kf)
		return cMock.MockGetProjectName(ke, kf)
	}
	panic("method 'GetProjectName' not implemented, yet")
}

func (cMock *ConsoleServer) SetupAccount(kg context.Context, kh *console.AccountSetupIn) (*console.AccountSetupVoid, error) {
	if cMock.MockSetupAccount != nil {
		cMock.registerCall("SetupAccount", kg, kh)
		return cMock.MockSetupAccount(kg, kh)
	}
	panic("method 'SetupAccount' not implemented, yet")
}

func NewConsoleServer() *ConsoleServer {
	return &ConsoleServer{}
}
