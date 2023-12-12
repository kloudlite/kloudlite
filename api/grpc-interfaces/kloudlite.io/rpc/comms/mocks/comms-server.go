package mocks

import (
	context "context"
	comms "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"
)

type CommsServerCallerInfo struct {
	Args []any
}

type CommsServer struct {
	Calls                            map[string][]CommsServerCallerInfo
	MockSendAccountMemberInviteEmail func(ka context.Context, kb *comms.AccountMemberInviteEmailInput) (*comms.Void, error)
	MockSendPasswordResetEmail       func(kc context.Context, kd *comms.PasswordResetEmailInput) (*comms.Void, error)
	MockSendProjectMemberInviteEmail func(ke context.Context, kf *comms.ProjectMemberInviteEmailInput) (*comms.Void, error)
	MockSendVerificationEmail        func(kg context.Context, kh *comms.VerificationEmailInput) (*comms.Void, error)
	MockSendWelcomeEmail             func(ki context.Context, kj *comms.WelcomeEmailInput) (*comms.Void, error)
}

func (m *CommsServer) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]CommsServerCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], CommsServerCallerInfo{Args: args})
}

func (cMock *CommsServer) SendAccountMemberInviteEmail(ka context.Context, kb *comms.AccountMemberInviteEmailInput) (*comms.Void, error) {
	if cMock.MockSendAccountMemberInviteEmail != nil {
		cMock.registerCall("SendAccountMemberInviteEmail", ka, kb)
		return cMock.MockSendAccountMemberInviteEmail(ka, kb)
	}
	panic("method 'SendAccountMemberInviteEmail' not implemented, yet")
}

func (cMock *CommsServer) SendPasswordResetEmail(kc context.Context, kd *comms.PasswordResetEmailInput) (*comms.Void, error) {
	if cMock.MockSendPasswordResetEmail != nil {
		cMock.registerCall("SendPasswordResetEmail", kc, kd)
		return cMock.MockSendPasswordResetEmail(kc, kd)
	}
	panic("method 'SendPasswordResetEmail' not implemented, yet")
}

func (cMock *CommsServer) SendProjectMemberInviteEmail(ke context.Context, kf *comms.ProjectMemberInviteEmailInput) (*comms.Void, error) {
	if cMock.MockSendProjectMemberInviteEmail != nil {
		cMock.registerCall("SendProjectMemberInviteEmail", ke, kf)
		return cMock.MockSendProjectMemberInviteEmail(ke, kf)
	}
	panic("method 'SendProjectMemberInviteEmail' not implemented, yet")
}

func (cMock *CommsServer) SendVerificationEmail(kg context.Context, kh *comms.VerificationEmailInput) (*comms.Void, error) {
	if cMock.MockSendVerificationEmail != nil {
		cMock.registerCall("SendVerificationEmail", kg, kh)
		return cMock.MockSendVerificationEmail(kg, kh)
	}
	panic("method 'SendVerificationEmail' not implemented, yet")
}

func (cMock *CommsServer) SendWelcomeEmail(ki context.Context, kj *comms.WelcomeEmailInput) (*comms.Void, error) {
	if cMock.MockSendWelcomeEmail != nil {
		cMock.registerCall("SendWelcomeEmail", ki, kj)
		return cMock.MockSendWelcomeEmail(ki, kj)
	}
	panic("method 'SendWelcomeEmail' not implemented, yet")
}

func NewCommsServer() *CommsServer {
	return &CommsServer{}
}
