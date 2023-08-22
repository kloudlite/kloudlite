package mocks

import (
	context1 "context"
	auth2 "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
)

type AuthServerCallerInfo struct {
	Args []any
}

type AuthServer struct {
	Calls                 map[string][]AuthServerCallerInfo
	MockEnsureUserByEmail func(ka context1.Context, kb *auth2.GetUserByEmailRequest) (*auth2.GetUserByEmailOut, error)
	MockGetAccessToken    func(kc context1.Context, kd *auth2.GetAccessTokenRequest) (*auth2.AccessTokenOut, error)
}

func (m *AuthServer) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]AuthServerCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], AuthServerCallerInfo{Args: args})
}

func (a *AuthServer) EnsureUserByEmail(ka context1.Context, kb *auth2.GetUserByEmailRequest) (*auth2.GetUserByEmailOut, error) {
	if a.MockEnsureUserByEmail != nil {
		a.registerCall("EnsureUserByEmail", ka, kb)
		return a.MockEnsureUserByEmail(ka, kb)
	}
	panic("not implemented, yet")
}

func (a *AuthServer) GetAccessToken(kc context1.Context, kd *auth2.GetAccessTokenRequest) (*auth2.AccessTokenOut, error) {
	if a.MockGetAccessToken != nil {
		a.registerCall("GetAccessToken", kc, kd)
		return a.MockGetAccessToken(kc, kd)
	}
	panic("not implemented, yet")
}

func NewAuthServer() *AuthServer {
	return &AuthServer{}
}
