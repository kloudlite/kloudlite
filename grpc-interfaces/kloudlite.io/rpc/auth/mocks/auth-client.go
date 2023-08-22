package mocks

import (
	context1 "context"
	grpc3 "google.golang.org/grpc"
	auth2 "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
)

type AuthClientCallerInfo struct {
	Args []any
}

type AuthClient struct {
	Calls                 map[string][]AuthClientCallerInfo
	MockEnsureUserByEmail func(ctx context1.Context, in *auth2.GetUserByEmailRequest, opts ...grpc3.CallOption) (*auth2.GetUserByEmailOut, error)
	MockGetAccessToken    func(ctx context1.Context, in *auth2.GetAccessTokenRequest, opts ...grpc3.CallOption) (*auth2.AccessTokenOut, error)
}

func (m *AuthClient) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]AuthClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], AuthClientCallerInfo{Args: args})
}

func (a *AuthClient) EnsureUserByEmail(ctx context1.Context, in *auth2.GetUserByEmailRequest, opts ...grpc3.CallOption) (*auth2.GetUserByEmailOut, error) {
	if a.MockEnsureUserByEmail != nil {
		a.registerCall("EnsureUserByEmail", ctx, in, opts)
		return a.MockEnsureUserByEmail(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (a *AuthClient) GetAccessToken(ctx context1.Context, in *auth2.GetAccessTokenRequest, opts ...grpc3.CallOption) (*auth2.AccessTokenOut, error) {
	if a.MockGetAccessToken != nil {
		a.registerCall("GetAccessToken", ctx, in, opts)
		return a.MockGetAccessToken(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func NewAuthClient() *AuthClient {
	return &AuthClient{}
}
