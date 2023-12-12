package mocks

import (
	context "context"
	auth "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	grpc "google.golang.org/grpc"
)

type AuthClientCallerInfo struct {
	Args []any
}

type AuthClient struct {
	Calls                 map[string][]AuthClientCallerInfo
	MockEnsureUserByEmail func(ctx context.Context, in *auth.GetUserByEmailRequest, opts ...grpc.CallOption) (*auth.GetUserByEmailOut, error)
	MockGetAccessToken    func(ctx context.Context, in *auth.GetAccessTokenRequest, opts ...grpc.CallOption) (*auth.AccessTokenOut, error)
	MockGetUser           func(ctx context.Context, in *auth.GetUserIn, opts ...grpc.CallOption) (*auth.GetUserOut, error)
}

func (m *AuthClient) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]AuthClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], AuthClientCallerInfo{Args: args})
}

func (aMock *AuthClient) EnsureUserByEmail(ctx context.Context, in *auth.GetUserByEmailRequest, opts ...grpc.CallOption) (*auth.GetUserByEmailOut, error) {
	if aMock.MockEnsureUserByEmail != nil {
		aMock.registerCall("EnsureUserByEmail", ctx, in, opts)
		return aMock.MockEnsureUserByEmail(ctx, in, opts...)
	}
	panic("method 'EnsureUserByEmail' not implemented, yet")
}

func (aMock *AuthClient) GetAccessToken(ctx context.Context, in *auth.GetAccessTokenRequest, opts ...grpc.CallOption) (*auth.AccessTokenOut, error) {
	if aMock.MockGetAccessToken != nil {
		aMock.registerCall("GetAccessToken", ctx, in, opts)
		return aMock.MockGetAccessToken(ctx, in, opts...)
	}
	panic("method 'GetAccessToken' not implemented, yet")
}

func (aMock *AuthClient) GetUser(ctx context.Context, in *auth.GetUserIn, opts ...grpc.CallOption) (*auth.GetUserOut, error) {
	if aMock.MockGetUser != nil {
		aMock.registerCall("GetUser", ctx, in, opts)
		return aMock.MockGetUser(ctx, in, opts...)
	}
	panic("method 'GetUser' not implemented, yet")
}

func NewAuthClient() *AuthClient {
	return &AuthClient{}
}
