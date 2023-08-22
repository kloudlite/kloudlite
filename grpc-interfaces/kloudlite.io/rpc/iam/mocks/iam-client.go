package mocks

import (
	context1 "context"
	grpc3 "google.golang.org/grpc"
	iam2 "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
)

type IAMClientCallerInfo struct {
	Args []any
}

type IAMClient struct {
	Calls                         map[string][]IAMClientCallerInfo
	MockAddMembership             func(ctx context1.Context, in *iam2.AddMembershipIn, opts ...grpc3.CallOption) (*iam2.AddMembershipOut, error)
	MockCan                       func(ctx context1.Context, in *iam2.CanIn, opts ...grpc3.CallOption) (*iam2.CanOut, error)
	MockConfirmMembership         func(ctx context1.Context, in *iam2.ConfirmMembershipIn, opts ...grpc3.CallOption) (*iam2.ConfirmMembershipOut, error)
	MockGetMembership             func(ctx context1.Context, in *iam2.GetMembershipIn, opts ...grpc3.CallOption) (*iam2.GetMembershipOut, error)
	MockInviteMembership          func(ctx context1.Context, in *iam2.AddMembershipIn, opts ...grpc3.CallOption) (*iam2.AddMembershipOut, error)
	MockListMembershipsByResource func(ctx context1.Context, in *iam2.MembershipsByResourceIn, opts ...grpc3.CallOption) (*iam2.ListMembershipsOut, error)
	MockListMembershipsForUser    func(ctx context1.Context, in *iam2.MembershipsForUserIn, opts ...grpc3.CallOption) (*iam2.ListMembershipsOut, error)
	MockListResourceMemberships   func(ctx context1.Context, in *iam2.ResourceMembershipsIn, opts ...grpc3.CallOption) (*iam2.ListMembershipsOut, error)
	MockListUserMemberships       func(ctx context1.Context, in *iam2.UserMembershipsIn, opts ...grpc3.CallOption) (*iam2.ListMembershipsOut, error)
	MockPing                      func(ctx context1.Context, in *iam2.Message, opts ...grpc3.CallOption) (*iam2.Message, error)
	MockRemoveMembership          func(ctx context1.Context, in *iam2.RemoveMembershipIn, opts ...grpc3.CallOption) (*iam2.RemoveMembershipOut, error)
	MockRemoveResource            func(ctx context1.Context, in *iam2.RemoveResourceIn, opts ...grpc3.CallOption) (*iam2.RemoveResourceOut, error)
}

func (m *IAMClient) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]IAMClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], IAMClientCallerInfo{Args: args})
}

func (i *IAMClient) AddMembership(ctx context1.Context, in *iam2.AddMembershipIn, opts ...grpc3.CallOption) (*iam2.AddMembershipOut, error) {
	if i.MockAddMembership != nil {
		i.registerCall("AddMembership", ctx, in, opts)
		return i.MockAddMembership(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (i *IAMClient) Can(ctx context1.Context, in *iam2.CanIn, opts ...grpc3.CallOption) (*iam2.CanOut, error) {
	if i.MockCan != nil {
		i.registerCall("Can", ctx, in, opts)
		return i.MockCan(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (i *IAMClient) ConfirmMembership(ctx context1.Context, in *iam2.ConfirmMembershipIn, opts ...grpc3.CallOption) (*iam2.ConfirmMembershipOut, error) {
	if i.MockConfirmMembership != nil {
		i.registerCall("ConfirmMembership", ctx, in, opts)
		return i.MockConfirmMembership(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (i *IAMClient) GetMembership(ctx context1.Context, in *iam2.GetMembershipIn, opts ...grpc3.CallOption) (*iam2.GetMembershipOut, error) {
	if i.MockGetMembership != nil {
		i.registerCall("GetMembership", ctx, in, opts)
		return i.MockGetMembership(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (i *IAMClient) InviteMembership(ctx context1.Context, in *iam2.AddMembershipIn, opts ...grpc3.CallOption) (*iam2.AddMembershipOut, error) {
	if i.MockInviteMembership != nil {
		i.registerCall("InviteMembership", ctx, in, opts)
		return i.MockInviteMembership(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (i *IAMClient) ListMembershipsByResource(ctx context1.Context, in *iam2.MembershipsByResourceIn, opts ...grpc3.CallOption) (*iam2.ListMembershipsOut, error) {
	if i.MockListMembershipsByResource != nil {
		i.registerCall("ListMembershipsByResource", ctx, in, opts)
		return i.MockListMembershipsByResource(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (i *IAMClient) ListMembershipsForUser(ctx context1.Context, in *iam2.MembershipsForUserIn, opts ...grpc3.CallOption) (*iam2.ListMembershipsOut, error) {
	if i.MockListMembershipsForUser != nil {
		i.registerCall("ListMembershipsForUser", ctx, in, opts)
		return i.MockListMembershipsForUser(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (i *IAMClient) ListResourceMemberships(ctx context1.Context, in *iam2.ResourceMembershipsIn, opts ...grpc3.CallOption) (*iam2.ListMembershipsOut, error) {
	if i.MockListResourceMemberships != nil {
		i.registerCall("ListResourceMemberships", ctx, in, opts)
		return i.MockListResourceMemberships(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (i *IAMClient) ListUserMemberships(ctx context1.Context, in *iam2.UserMembershipsIn, opts ...grpc3.CallOption) (*iam2.ListMembershipsOut, error) {
	if i.MockListUserMemberships != nil {
		i.registerCall("ListUserMemberships", ctx, in, opts)
		return i.MockListUserMemberships(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (i *IAMClient) Ping(ctx context1.Context, in *iam2.Message, opts ...grpc3.CallOption) (*iam2.Message, error) {
	if i.MockPing != nil {
		i.registerCall("Ping", ctx, in, opts)
		return i.MockPing(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (i *IAMClient) RemoveMembership(ctx context1.Context, in *iam2.RemoveMembershipIn, opts ...grpc3.CallOption) (*iam2.RemoveMembershipOut, error) {
	if i.MockRemoveMembership != nil {
		i.registerCall("RemoveMembership", ctx, in, opts)
		return i.MockRemoveMembership(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func (i *IAMClient) RemoveResource(ctx context1.Context, in *iam2.RemoveResourceIn, opts ...grpc3.CallOption) (*iam2.RemoveResourceOut, error) {
	if i.MockRemoveResource != nil {
		i.registerCall("RemoveResource", ctx, in, opts)
		return i.MockRemoveResource(ctx, in, opts...)
	}
	panic("not implemented, yet")
}

func NewIAMClient() *IAMClient {
	return &IAMClient{}
}
