package mocks

import (
	context1 "context"
	grpc3 "google.golang.org/grpc"
	comms2 "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/comms"
)

type CommsClientCallerInfo struct {
	Args []any
}

type CommsClient struct {
	Calls                            map[string][]CommsClientCallerInfo
	MockSendAccountMemberInviteEmail func(ctx context1.Context, in *comms2.AccountMemberInviteEmailInput, opts ...grpc3.CallOption) (*comms2.Void, error)
	MockSendPasswordResetEmail       func(ctx context1.Context, in *comms2.PasswordResetEmailInput, opts ...grpc3.CallOption) (*comms2.Void, error)
	MockSendProjectMemberInviteEmail func(ctx context1.Context, in *comms2.ProjectMemberInviteEmailInput, opts ...grpc3.CallOption) (*comms2.Void, error)
	MockSendVerificationEmail        func(ctx context1.Context, in *comms2.VerificationEmailInput, opts ...grpc3.CallOption) (*comms2.Void, error)
	MockSendWelcomeEmail             func(ctx context1.Context, in *comms2.WelcomeEmailInput, opts ...grpc3.CallOption) (*comms2.Void, error)
}

func (m *CommsClient) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]CommsClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], CommsClientCallerInfo{Args: args})
}

func (c *CommsClient) SendAccountMemberInviteEmail(ctx context1.Context, in *comms2.AccountMemberInviteEmailInput, opts ...grpc3.CallOption) (*comms2.Void, error) {
	if c.MockSendAccountMemberInviteEmail != nil {
		c.registerCall("SendAccountMemberInviteEmail", ctx, in, opts)
		return c.MockSendAccountMemberInviteEmail(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (c *CommsClient) SendPasswordResetEmail(ctx context1.Context, in *comms2.PasswordResetEmailInput, opts ...grpc3.CallOption) (*comms2.Void, error) {
	if c.MockSendPasswordResetEmail != nil {
		c.registerCall("SendPasswordResetEmail", ctx, in, opts)
		return c.MockSendPasswordResetEmail(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (c *CommsClient) SendProjectMemberInviteEmail(ctx context1.Context, in *comms2.ProjectMemberInviteEmailInput, opts ...grpc3.CallOption) (*comms2.Void, error) {
	if c.MockSendProjectMemberInviteEmail != nil {
		c.registerCall("SendProjectMemberInviteEmail", ctx, in, opts)
		return c.MockSendProjectMemberInviteEmail(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (c *CommsClient) SendVerificationEmail(ctx context1.Context, in *comms2.VerificationEmailInput, opts ...grpc3.CallOption) (*comms2.Void, error) {
	if c.MockSendVerificationEmail != nil {
		c.registerCall("SendVerificationEmail", ctx, in, opts)
		return c.MockSendVerificationEmail(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (c *CommsClient) SendWelcomeEmail(ctx context1.Context, in *comms2.WelcomeEmailInput, opts ...grpc3.CallOption) (*comms2.Void, error) {
	if c.MockSendWelcomeEmail != nil {
		c.registerCall("SendWelcomeEmail", ctx, in, opts)
		return c.MockSendWelcomeEmail(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func NewCommsClient() *CommsClient {
	return &CommsClient{}
}
