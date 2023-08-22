package mocks

import (
	context1 "context"
	comms2 "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/comms"
)

type CommsServerCallerInfo struct {
	Args []any
}

type CommsServer struct {
	Calls                            map[string][]CommsServerCallerInfo
	MockSendAccountMemberInviteEmail func(ka context1.Context, kb *comms2.AccountMemberInviteEmailInput) (*comms2.Void, error)
	MockSendPasswordResetEmail       func(kc context1.Context, kd *comms2.PasswordResetEmailInput) (*comms2.Void, error)
	MockSendProjectMemberInviteEmail func(ke context1.Context, kf *comms2.ProjectMemberInviteEmailInput) (*comms2.Void, error)
	MockSendVerificationEmail        func(kg context1.Context, kh *comms2.VerificationEmailInput) (*comms2.Void, error)
	MockSendWelcomeEmail             func(ki context1.Context, kj *comms2.WelcomeEmailInput) (*comms2.Void, error)
}

func (m *CommsServer) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]CommsServerCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], CommsServerCallerInfo{Args: args})
}

func (c *CommsServer) SendAccountMemberInviteEmail(ka context1.Context, kb *comms2.AccountMemberInviteEmailInput) (*comms2.Void, error) {
	if c.MockSendAccountMemberInviteEmail != nil {
		c.registerCall("SendAccountMemberInviteEmail", ka, kb)
		return c.MockSendAccountMemberInviteEmail(ka, kb)
	}
	panic("not implemented, yet")
}

func (c *CommsServer) SendPasswordResetEmail(kc context1.Context, kd *comms2.PasswordResetEmailInput) (*comms2.Void, error) {
	if c.MockSendPasswordResetEmail != nil {
		c.registerCall("SendPasswordResetEmail", kc, kd)
		return c.MockSendPasswordResetEmail(kc, kd)
	}
	panic("not implemented, yet")
}

func (c *CommsServer) SendProjectMemberInviteEmail(ke context1.Context, kf *comms2.ProjectMemberInviteEmailInput) (*comms2.Void, error) {
	if c.MockSendProjectMemberInviteEmail != nil {
		c.registerCall("SendProjectMemberInviteEmail", ke, kf)
		return c.MockSendProjectMemberInviteEmail(ke, kf)
	}
	panic("not implemented, yet")
}

func (c *CommsServer) SendVerificationEmail(kg context1.Context, kh *comms2.VerificationEmailInput) (*comms2.Void, error) {
	if c.MockSendVerificationEmail != nil {
		c.registerCall("SendVerificationEmail", kg, kh)
		return c.MockSendVerificationEmail(kg, kh)
	}
	panic("not implemented, yet")
}

func (c *CommsServer) SendWelcomeEmail(ki context1.Context, kj *comms2.WelcomeEmailInput) (*comms2.Void, error) {
	if c.MockSendWelcomeEmail != nil {
		c.registerCall("SendWelcomeEmail", ki, kj)
		return c.MockSendWelcomeEmail(ki, kj)
	}
	panic("not implemented, yet")
}

func NewCommsServer() *CommsServer {
	return &CommsServer{}
}
