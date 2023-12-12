package mocks

import (
	context "context"
	auth "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
)

type AuthServerCallerInfo struct {
	Args []any
}

type AuthServer struct {
	Calls                 map[string][]AuthServerCallerInfo
	MockEnsureUserByEmail func(ka context.Context, kb *auth.GetUserByEmailRequest) (*auth.GetUserByEmailOut, error)
	MockGetAccessToken    func(kc context.Context, kd *auth.GetAccessTokenRequest) (*auth.AccessTokenOut, error)
	MockGetUser           func(ke context.Context, kf *auth.GetUserIn) (*auth.GetUserOut, error)
}

func (m *AuthServer) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]AuthServerCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], AuthServerCallerInfo{Args: args})
}

func (aMock *AuthServer) EnsureUserByEmail(ka context.Context, kb *auth.GetUserByEmailRequest) (*auth.GetUserByEmailOut, error) {
	if aMock.MockEnsureUserByEmail != nil {
		aMock.registerCall("EnsureUserByEmail", ka, kb)
		return aMock.MockEnsureUserByEmail(ka, kb)
	}
	panic("method 'EnsureUserByEmail' not implemented, yet")
}

func (aMock *AuthServer) GetAccessToken(kc context.Context, kd *auth.GetAccessTokenRequest) (*auth.AccessTokenOut, error) {
	if aMock.MockGetAccessToken != nil {
		aMock.registerCall("GetAccessToken", kc, kd)
		return aMock.MockGetAccessToken(kc, kd)
	}
	panic("method 'GetAccessToken' not implemented, yet")
}

func (aMock *AuthServer) GetUser(ke context.Context, kf *auth.GetUserIn) (*auth.GetUserOut, error) {
	if aMock.MockGetUser != nil {
		aMock.registerCall("GetUser", ke, kf)
		return aMock.MockGetUser(ke, kf)
	}
	panic("method 'GetUser' not implemented, yet")
}

func NewAuthServer() *AuthServer {
	return &AuthServer{}
}
