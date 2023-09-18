package mocks

import (
	context "context"
	grpc "google.golang.org/grpc"
	iam "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
)

type IAMClientCallerInfo struct {
	Args []any
}

type IAMClient struct {
	Calls                          map[string][]IAMClientCallerInfo
	MockAddMembership              func(ctx context.Context, in *iam.AddMembershipIn, opts ...grpc.CallOption) (*iam.AddMembershipOut, error)
	MockCan                        func(ctx context.Context, in *iam.CanIn, opts ...grpc.CallOption) (*iam.CanOut, error)
	MockGetMembership              func(ctx context.Context, in *iam.GetMembershipIn, opts ...grpc.CallOption) (*iam.GetMembershipOut, error)
	MockListMembershipsForResource func(ctx context.Context, in *iam.MembershipsForResourceIn, opts ...grpc.CallOption) (*iam.ListMembershipsOut, error)
	MockListMembershipsForUser     func(ctx context.Context, in *iam.MembershipsForUserIn, opts ...grpc.CallOption) (*iam.ListMembershipsOut, error)
	MockPing                       func(ctx context.Context, in *iam.Message, opts ...grpc.CallOption) (*iam.Message, error)
	MockRemoveMembership           func(ctx context.Context, in *iam.RemoveMembershipIn, opts ...grpc.CallOption) (*iam.RemoveMembershipOut, error)
	MockRemoveResource             func(ctx context.Context, in *iam.RemoveResourceIn, opts ...grpc.CallOption) (*iam.RemoveResourceOut, error)
	MockUpdateMembership           func(ctx context.Context, in *iam.UpdateMembershipIn, opts ...grpc.CallOption) (*iam.UpdateMembershipOut, error)
}

func (m *IAMClient) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]IAMClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], IAMClientCallerInfo{Args: args})
}

func (iMock *IAMClient) AddMembership(ctx context.Context, in *iam.AddMembershipIn, opts ...grpc.CallOption) (*iam.AddMembershipOut, error) {
	if iMock.MockAddMembership != nil {
		iMock.registerCall("AddMembership", ctx, in, opts)
		return iMock.MockAddMembership(ctx, in, opts...)
	}
	panic("method 'AddMembership' not implemented, yet")
}

func (iMock *IAMClient) Can(ctx context.Context, in *iam.CanIn, opts ...grpc.CallOption) (*iam.CanOut, error) {
	if iMock.MockCan != nil {
		iMock.registerCall("Can", ctx, in, opts)
		return iMock.MockCan(ctx, in, opts...)
	}
	panic("method 'Can' not implemented, yet")
}

func (iMock *IAMClient) GetMembership(ctx context.Context, in *iam.GetMembershipIn, opts ...grpc.CallOption) (*iam.GetMembershipOut, error) {
	if iMock.MockGetMembership != nil {
		iMock.registerCall("GetMembership", ctx, in, opts)
		return iMock.MockGetMembership(ctx, in, opts...)
	}
	panic("method 'GetMembership' not implemented, yet")
}

func (iMock *IAMClient) ListMembershipsForResource(ctx context.Context, in *iam.MembershipsForResourceIn, opts ...grpc.CallOption) (*iam.ListMembershipsOut, error) {
	if iMock.MockListMembershipsForResource != nil {
		iMock.registerCall("ListMembershipsForResource", ctx, in, opts)
		return iMock.MockListMembershipsForResource(ctx, in, opts...)
	}
	panic("method 'ListMembershipsForResource' not implemented, yet")
}

func (iMock *IAMClient) ListMembershipsForUser(ctx context.Context, in *iam.MembershipsForUserIn, opts ...grpc.CallOption) (*iam.ListMembershipsOut, error) {
	if iMock.MockListMembershipsForUser != nil {
		iMock.registerCall("ListMembershipsForUser", ctx, in, opts)
		return iMock.MockListMembershipsForUser(ctx, in, opts...)
	}
	panic("method 'ListMembershipsForUser' not implemented, yet")
}

func (iMock *IAMClient) Ping(ctx context.Context, in *iam.Message, opts ...grpc.CallOption) (*iam.Message, error) {
	if iMock.MockPing != nil {
		iMock.registerCall("Ping", ctx, in, opts)
		return iMock.MockPing(ctx, in, opts...)
	}
	panic("method 'Ping' not implemented, yet")
}

func (iMock *IAMClient) RemoveMembership(ctx context.Context, in *iam.RemoveMembershipIn, opts ...grpc.CallOption) (*iam.RemoveMembershipOut, error) {
	if iMock.MockRemoveMembership != nil {
		iMock.registerCall("RemoveMembership", ctx, in, opts)
		return iMock.MockRemoveMembership(ctx, in, opts...)
	}
	panic("method 'RemoveMembership' not implemented, yet")
}

func (iMock *IAMClient) RemoveResource(ctx context.Context, in *iam.RemoveResourceIn, opts ...grpc.CallOption) (*iam.RemoveResourceOut, error) {
	if iMock.MockRemoveResource != nil {
		iMock.registerCall("RemoveResource", ctx, in, opts)
		return iMock.MockRemoveResource(ctx, in, opts...)
	}
	panic("method 'RemoveResource' not implemented, yet")
}

func (iMock *IAMClient) UpdateMembership(ctx context.Context, in *iam.UpdateMembershipIn, opts ...grpc.CallOption) (*iam.UpdateMembershipOut, error) {
	if iMock.MockUpdateMembership != nil {
		iMock.registerCall("UpdateMembership", ctx, in, opts)
		return iMock.MockUpdateMembership(ctx, in, opts...)
	}
	panic("method 'UpdateMembership' not implemented, yet")
}

func NewIAMClient() *IAMClient {
	return &IAMClient{}
}
