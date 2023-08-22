package mocks

import (
	context1 "context"
	iam2 "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
)

type IAMServerCallerInfo struct {
	Args []any
}

type IAMServer struct {
	Calls                         map[string][]IAMServerCallerInfo
	MockAddMembership             func(ka context1.Context, kb *iam2.AddMembershipIn) (*iam2.AddMembershipOut, error)
	MockCan                       func(kc context1.Context, kd *iam2.CanIn) (*iam2.CanOut, error)
	MockConfirmMembership         func(ke context1.Context, kf *iam2.ConfirmMembershipIn) (*iam2.ConfirmMembershipOut, error)
	MockGetMembership             func(kg context1.Context, kh *iam2.GetMembershipIn) (*iam2.GetMembershipOut, error)
	MockInviteMembership          func(ki context1.Context, kj *iam2.AddMembershipIn) (*iam2.AddMembershipOut, error)
	MockListMembershipsByResource func(kk context1.Context, kl *iam2.MembershipsByResourceIn) (*iam2.ListMembershipsOut, error)
	MockListMembershipsForUser    func(km context1.Context, kn *iam2.MembershipsForUserIn) (*iam2.ListMembershipsOut, error)
	MockListResourceMemberships   func(ko context1.Context, kp *iam2.ResourceMembershipsIn) (*iam2.ListMembershipsOut, error)
	MockListUserMemberships       func(kq context1.Context, kr *iam2.UserMembershipsIn) (*iam2.ListMembershipsOut, error)
	MockPing                      func(ks context1.Context, kt *iam2.Message) (*iam2.Message, error)
	MockRemoveMembership          func(ku context1.Context, kv *iam2.RemoveMembershipIn) (*iam2.RemoveMembershipOut, error)
	MockRemoveResource            func(kw context1.Context, kx *iam2.RemoveResourceIn) (*iam2.RemoveResourceOut, error)
}

func (m *IAMServer) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]IAMServerCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], IAMServerCallerInfo{Args: args})
}

func (i *IAMServer) AddMembership(ka context1.Context, kb *iam2.AddMembershipIn) (*iam2.AddMembershipOut, error) {
	if i.MockAddMembership != nil {
		i.registerCall("AddMembership", ka, kb)
		return i.MockAddMembership(ka, kb)
	}
	panic("not implemented, yet")
}

func (i *IAMServer) Can(kc context1.Context, kd *iam2.CanIn) (*iam2.CanOut, error) {
	if i.MockCan != nil {
		i.registerCall("Can", kc, kd)
		return i.MockCan(kc, kd)
	}
	panic("not implemented, yet")
}

func (i *IAMServer) ConfirmMembership(ke context1.Context, kf *iam2.ConfirmMembershipIn) (*iam2.ConfirmMembershipOut, error) {
	if i.MockConfirmMembership != nil {
		i.registerCall("ConfirmMembership", ke, kf)
		return i.MockConfirmMembership(ke, kf)
	}
	panic("not implemented, yet")
}

func (i *IAMServer) GetMembership(kg context1.Context, kh *iam2.GetMembershipIn) (*iam2.GetMembershipOut, error) {
	if i.MockGetMembership != nil {
		i.registerCall("GetMembership", kg, kh)
		return i.MockGetMembership(kg, kh)
	}
	panic("not implemented, yet")
}

func (i *IAMServer) InviteMembership(ki context1.Context, kj *iam2.AddMembershipIn) (*iam2.AddMembershipOut, error) {
	if i.MockInviteMembership != nil {
		i.registerCall("InviteMembership", ki, kj)
		return i.MockInviteMembership(ki, kj)
	}
	panic("not implemented, yet")
}

func (i *IAMServer) ListMembershipsByResource(kk context1.Context, kl *iam2.MembershipsByResourceIn) (*iam2.ListMembershipsOut, error) {
	if i.MockListMembershipsByResource != nil {
		i.registerCall("ListMembershipsByResource", kk, kl)
		return i.MockListMembershipsByResource(kk, kl)
	}
	panic("not implemented, yet")
}

func (i *IAMServer) ListMembershipsForUser(km context1.Context, kn *iam2.MembershipsForUserIn) (*iam2.ListMembershipsOut, error) {
	if i.MockListMembershipsForUser != nil {
		i.registerCall("ListMembershipsForUser", km, kn)
		return i.MockListMembershipsForUser(km, kn)
	}
	panic("not implemented, yet")
}

func (i *IAMServer) ListResourceMemberships(ko context1.Context, kp *iam2.ResourceMembershipsIn) (*iam2.ListMembershipsOut, error) {
	if i.MockListResourceMemberships != nil {
		i.registerCall("ListResourceMemberships", ko, kp)
		return i.MockListResourceMemberships(ko, kp)
	}
	panic("not implemented, yet")
}

func (i *IAMServer) ListUserMemberships(kq context1.Context, kr *iam2.UserMembershipsIn) (*iam2.ListMembershipsOut, error) {
	if i.MockListUserMemberships != nil {
		i.registerCall("ListUserMemberships", kq, kr)
		return i.MockListUserMemberships(kq, kr)
	}
	panic("not implemented, yet")
}

func (i *IAMServer) Ping(ks context1.Context, kt *iam2.Message) (*iam2.Message, error) {
	if i.MockPing != nil {
		i.registerCall("Ping", ks, kt)
		return i.MockPing(ks, kt)
	}
	panic("not implemented, yet")
}

func (i *IAMServer) RemoveMembership(ku context1.Context, kv *iam2.RemoveMembershipIn) (*iam2.RemoveMembershipOut, error) {
	if i.MockRemoveMembership != nil {
		i.registerCall("RemoveMembership", ku, kv)
		return i.MockRemoveMembership(ku, kv)
	}
	panic("not implemented, yet")
}

func (i *IAMServer) RemoveResource(kw context1.Context, kx *iam2.RemoveResourceIn) (*iam2.RemoveResourceOut, error) {
	if i.MockRemoveResource != nil {
		i.registerCall("RemoveResource", kw, kx)
		return i.MockRemoveResource(kw, kx)
	}
	panic("not implemented, yet")
}

func NewIAMServer() *IAMServer {
	return &IAMServer{}
}
