package mocks

import (
	context "context"
	iam "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
)

type IAMServerCallerInfo struct {
	Args []any
}

type IAMServer struct {
	Calls                          map[string][]IAMServerCallerInfo
	MockAddMembership              func(ka context.Context, kb *iam.AddMembershipIn) (*iam.AddMembershipOut, error)
	MockCan                        func(kc context.Context, kd *iam.CanIn) (*iam.CanOut, error)
	MockGetMembership              func(ke context.Context, kf *iam.GetMembershipIn) (*iam.GetMembershipOut, error)
	MockListMembershipsForResource func(kg context.Context, kh *iam.MembershipsForResourceIn) (*iam.ListMembershipsOut, error)
	MockListMembershipsForUser     func(ki context.Context, kj *iam.MembershipsForUserIn) (*iam.ListMembershipsOut, error)
	MockPing                       func(kk context.Context, kl *iam.Message) (*iam.Message, error)
	MockRemoveMembership           func(km context.Context, kn *iam.RemoveMembershipIn) (*iam.RemoveMembershipOut, error)
	MockRemoveResource             func(ko context.Context, kp *iam.RemoveResourceIn) (*iam.RemoveResourceOut, error)
	MockUpdateMembership           func(kq context.Context, kr *iam.UpdateMembershipIn) (*iam.UpdateMembershipOut, error)
}

func (m *IAMServer) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]IAMServerCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], IAMServerCallerInfo{Args: args})
}

func (iMock *IAMServer) AddMembership(ka context.Context, kb *iam.AddMembershipIn) (*iam.AddMembershipOut, error) {
	if iMock.MockAddMembership != nil {
		iMock.registerCall("AddMembership", ka, kb)
		return iMock.MockAddMembership(ka, kb)
	}
	panic("method 'AddMembership' not implemented, yet")
}

func (iMock *IAMServer) Can(kc context.Context, kd *iam.CanIn) (*iam.CanOut, error) {
	if iMock.MockCan != nil {
		iMock.registerCall("Can", kc, kd)
		return iMock.MockCan(kc, kd)
	}
	panic("method 'Can' not implemented, yet")
}

func (iMock *IAMServer) GetMembership(ke context.Context, kf *iam.GetMembershipIn) (*iam.GetMembershipOut, error) {
	if iMock.MockGetMembership != nil {
		iMock.registerCall("GetMembership", ke, kf)
		return iMock.MockGetMembership(ke, kf)
	}
	panic("method 'GetMembership' not implemented, yet")
}

func (iMock *IAMServer) ListMembershipsForResource(kg context.Context, kh *iam.MembershipsForResourceIn) (*iam.ListMembershipsOut, error) {
	if iMock.MockListMembershipsForResource != nil {
		iMock.registerCall("ListMembershipsForResource", kg, kh)
		return iMock.MockListMembershipsForResource(kg, kh)
	}
	panic("method 'ListMembershipsForResource' not implemented, yet")
}

func (iMock *IAMServer) ListMembershipsForUser(ki context.Context, kj *iam.MembershipsForUserIn) (*iam.ListMembershipsOut, error) {
	if iMock.MockListMembershipsForUser != nil {
		iMock.registerCall("ListMembershipsForUser", ki, kj)
		return iMock.MockListMembershipsForUser(ki, kj)
	}
	panic("method 'ListMembershipsForUser' not implemented, yet")
}

func (iMock *IAMServer) Ping(kk context.Context, kl *iam.Message) (*iam.Message, error) {
	if iMock.MockPing != nil {
		iMock.registerCall("Ping", kk, kl)
		return iMock.MockPing(kk, kl)
	}
	panic("method 'Ping' not implemented, yet")
}

func (iMock *IAMServer) RemoveMembership(km context.Context, kn *iam.RemoveMembershipIn) (*iam.RemoveMembershipOut, error) {
	if iMock.MockRemoveMembership != nil {
		iMock.registerCall("RemoveMembership", km, kn)
		return iMock.MockRemoveMembership(km, kn)
	}
	panic("method 'RemoveMembership' not implemented, yet")
}

func (iMock *IAMServer) RemoveResource(ko context.Context, kp *iam.RemoveResourceIn) (*iam.RemoveResourceOut, error) {
	if iMock.MockRemoveResource != nil {
		iMock.registerCall("RemoveResource", ko, kp)
		return iMock.MockRemoveResource(ko, kp)
	}
	panic("method 'RemoveResource' not implemented, yet")
}

func (iMock *IAMServer) UpdateMembership(kq context.Context, kr *iam.UpdateMembershipIn) (*iam.UpdateMembershipOut, error) {
	if iMock.MockUpdateMembership != nil {
		iMock.registerCall("UpdateMembership", kq, kr)
		return iMock.MockUpdateMembership(kq, kr)
	}
	panic("method 'UpdateMembership' not implemented, yet")
}

func NewIAMServer() *IAMServer {
	return &IAMServer{}
}
