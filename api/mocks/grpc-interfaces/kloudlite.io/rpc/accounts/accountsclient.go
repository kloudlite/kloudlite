package mocks

import (
	context "context"
	grpc "google.golang.org/grpc"
	accounts "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/accounts"
)

type AccountsClientCallerInfo struct {
	Args []any
}

type AccountsClient struct {
	Calls          map[string][]AccountsClientCallerInfo
	MockGetAccount func(ctx context.Context, in *accounts.GetAccountIn, opts ...grpc.CallOption) (*accounts.GetAccountOut, error)
}

func (m *AccountsClient) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]AccountsClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], AccountsClientCallerInfo{Args: args})
}

func (aMock *AccountsClient) GetAccount(ctx context.Context, in *accounts.GetAccountIn, opts ...grpc.CallOption) (*accounts.GetAccountOut, error) {
	if aMock.MockGetAccount != nil {
		aMock.registerCall("GetAccount", ctx, in, opts)
		return aMock.MockGetAccount(ctx, in, opts...)
	}
	panic("AccountsClient: method 'GetAccount' not implemented, yet")
}

func NewAccountsClient() *AccountsClient {
	return &AccountsClient{}
}
