package mocks

import (
	context "context"
	console "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/console"
	grpc "google.golang.org/grpc"
)

type ConsoleClientCallerInfo struct {
	Args []any
}

type ConsoleClient struct {
	Calls              map[string][]ConsoleClientCallerInfo
	MockGetApp         func(ctx context.Context, in *console.AppIn, opts ...grpc.CallOption) (*console.AppOut, error)
	MockGetManagedSvc  func(ctx context.Context, in *console.MSvcIn, opts ...grpc.CallOption) (*console.MSvcOut, error)
	MockGetProjectName func(ctx context.Context, in *console.ProjectIn, opts ...grpc.CallOption) (*console.ProjectOut, error)
	MockSetupAccount   func(ctx context.Context, in *console.AccountSetupIn, opts ...grpc.CallOption) (*console.AccountSetupVoid, error)
}

func (m *ConsoleClient) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]ConsoleClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], ConsoleClientCallerInfo{Args: args})
}

func (cMock *ConsoleClient) GetApp(ctx context.Context, in *console.AppIn, opts ...grpc.CallOption) (*console.AppOut, error) {
	if cMock.MockGetApp != nil {
		cMock.registerCall("GetApp", ctx, in, opts)
		return cMock.MockGetApp(ctx, in, opts...)
	}
	panic("method 'GetApp' not implemented, yet")
}

func (cMock *ConsoleClient) GetManagedSvc(ctx context.Context, in *console.MSvcIn, opts ...grpc.CallOption) (*console.MSvcOut, error) {
	if cMock.MockGetManagedSvc != nil {
		cMock.registerCall("GetManagedSvc", ctx, in, opts)
		return cMock.MockGetManagedSvc(ctx, in, opts...)
	}
	panic("method 'GetManagedSvc' not implemented, yet")
}

func (cMock *ConsoleClient) GetProjectName(ctx context.Context, in *console.ProjectIn, opts ...grpc.CallOption) (*console.ProjectOut, error) {
	if cMock.MockGetProjectName != nil {
		cMock.registerCall("GetProjectName", ctx, in, opts)
		return cMock.MockGetProjectName(ctx, in, opts...)
	}
	panic("method 'GetProjectName' not implemented, yet")
}

func (cMock *ConsoleClient) SetupAccount(ctx context.Context, in *console.AccountSetupIn, opts ...grpc.CallOption) (*console.AccountSetupVoid, error) {
	if cMock.MockSetupAccount != nil {
		cMock.registerCall("SetupAccount", ctx, in, opts)
		return cMock.MockSetupAccount(ctx, in, opts...)
	}
	panic("method 'SetupAccount' not implemented, yet")
}

func NewConsoleClient() *ConsoleClient {
	return &ConsoleClient{}
}
