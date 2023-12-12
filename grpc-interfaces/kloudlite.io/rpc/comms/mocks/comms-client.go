package mocks

import (
	context "context"
	comms "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"
	grpc "google.golang.org/grpc"
)

type CommsClientCallerInfo struct {
	Args []any
}

type CommsClient struct {
	Calls                            map[string][]CommsClientCallerInfo
	MockSendAccountMemberInviteEmail func(ctx context.Context, in *comms.AccountMemberInviteEmailInput, opts ...grpc.CallOption) (*comms.Void, error)
	MockSendPasswordResetEmail       func(ctx context.Context, in *comms.PasswordResetEmailInput, opts ...grpc.CallOption) (*comms.Void, error)
	MockSendProjectMemberInviteEmail func(ctx context.Context, in *comms.ProjectMemberInviteEmailInput, opts ...grpc.CallOption) (*comms.Void, error)
	MockSendVerificationEmail        func(ctx context.Context, in *comms.VerificationEmailInput, opts ...grpc.CallOption) (*comms.Void, error)
	MockSendWelcomeEmail             func(ctx context.Context, in *comms.WelcomeEmailInput, opts ...grpc.CallOption) (*comms.Void, error)
}

func (m *CommsClient) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]CommsClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], CommsClientCallerInfo{Args: args})
}

func (cMock *CommsClient) SendAccountMemberInviteEmail(ctx context.Context, in *comms.AccountMemberInviteEmailInput, opts ...grpc.CallOption) (*comms.Void, error) {
	if cMock.MockSendAccountMemberInviteEmail != nil {
		cMock.registerCall("SendAccountMemberInviteEmail", ctx, in, opts)
		return cMock.MockSendAccountMemberInviteEmail(ctx, in, opts...)
	}
	panic("method 'SendAccountMemberInviteEmail' not implemented, yet")
}

func (cMock *CommsClient) SendPasswordResetEmail(ctx context.Context, in *comms.PasswordResetEmailInput, opts ...grpc.CallOption) (*comms.Void, error) {
	if cMock.MockSendPasswordResetEmail != nil {
		cMock.registerCall("SendPasswordResetEmail", ctx, in, opts)
		return cMock.MockSendPasswordResetEmail(ctx, in, opts...)
	}
	panic("method 'SendPasswordResetEmail' not implemented, yet")
}

func (cMock *CommsClient) SendProjectMemberInviteEmail(ctx context.Context, in *comms.ProjectMemberInviteEmailInput, opts ...grpc.CallOption) (*comms.Void, error) {
	if cMock.MockSendProjectMemberInviteEmail != nil {
		cMock.registerCall("SendProjectMemberInviteEmail", ctx, in, opts)
		return cMock.MockSendProjectMemberInviteEmail(ctx, in, opts...)
	}
	panic("method 'SendProjectMemberInviteEmail' not implemented, yet")
}

func (cMock *CommsClient) SendVerificationEmail(ctx context.Context, in *comms.VerificationEmailInput, opts ...grpc.CallOption) (*comms.Void, error) {
	if cMock.MockSendVerificationEmail != nil {
		cMock.registerCall("SendVerificationEmail", ctx, in, opts)
		return cMock.MockSendVerificationEmail(ctx, in, opts...)
	}
	panic("method 'SendVerificationEmail' not implemented, yet")
}

func (cMock *CommsClient) SendWelcomeEmail(ctx context.Context, in *comms.WelcomeEmailInput, opts ...grpc.CallOption) (*comms.Void, error) {
	if cMock.MockSendWelcomeEmail != nil {
		cMock.registerCall("SendWelcomeEmail", ctx, in, opts)
		return cMock.MockSendWelcomeEmail(ctx, in, opts...)
	}
	panic("method 'SendWelcomeEmail' not implemented, yet")
}

func NewCommsClient() *CommsClient {
	return &CommsClient{}
}
