// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.4
// source: iam.proto

package iam

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	IAM_Ping_FullMethodName                       = "/IAM/Ping"
	IAM_Can_FullMethodName                        = "/IAM/Can"
	IAM_ListMembershipsForResource_FullMethodName = "/IAM/ListMembershipsForResource"
	IAM_ListMembershipsForUser_FullMethodName     = "/IAM/ListMembershipsForUser"
	IAM_GetMembership_FullMethodName              = "/IAM/GetMembership"
	IAM_AddMembership_FullMethodName              = "/IAM/AddMembership"
	IAM_UpdateMembership_FullMethodName           = "/IAM/UpdateMembership"
	IAM_RemoveMembership_FullMethodName           = "/IAM/RemoveMembership"
	IAM_RemoveResource_FullMethodName             = "/IAM/RemoveResource"
)

// IAMClient is the client API for IAM service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IAMClient interface {
	// Query
	Ping(ctx context.Context, in *Message, opts ...grpc.CallOption) (*Message, error)
	Can(ctx context.Context, in *CanIn, opts ...grpc.CallOption) (*CanOut, error)
	ListMembershipsForResource(ctx context.Context, in *MembershipsForResourceIn, opts ...grpc.CallOption) (*ListMembershipsOut, error)
	ListMembershipsForUser(ctx context.Context, in *MembershipsForUserIn, opts ...grpc.CallOption) (*ListMembershipsOut, error)
	GetMembership(ctx context.Context, in *GetMembershipIn, opts ...grpc.CallOption) (*GetMembershipOut, error)
	// Mutation
	AddMembership(ctx context.Context, in *AddMembershipIn, opts ...grpc.CallOption) (*AddMembershipOut, error)
	UpdateMembership(ctx context.Context, in *UpdateMembershipIn, opts ...grpc.CallOption) (*UpdateMembershipOut, error)
	RemoveMembership(ctx context.Context, in *RemoveMembershipIn, opts ...grpc.CallOption) (*RemoveMembershipOut, error)
	RemoveResource(ctx context.Context, in *RemoveResourceIn, opts ...grpc.CallOption) (*RemoveResourceOut, error)
}

type iAMClient struct {
	cc grpc.ClientConnInterface
}

func NewIAMClient(cc grpc.ClientConnInterface) IAMClient {
	return &iAMClient{cc}
}

func (c *iAMClient) Ping(ctx context.Context, in *Message, opts ...grpc.CallOption) (*Message, error) {
	out := new(Message)
	err := c.cc.Invoke(ctx, IAM_Ping_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iAMClient) Can(ctx context.Context, in *CanIn, opts ...grpc.CallOption) (*CanOut, error) {
	out := new(CanOut)
	err := c.cc.Invoke(ctx, IAM_Can_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iAMClient) ListMembershipsForResource(ctx context.Context, in *MembershipsForResourceIn, opts ...grpc.CallOption) (*ListMembershipsOut, error) {
	out := new(ListMembershipsOut)
	err := c.cc.Invoke(ctx, IAM_ListMembershipsForResource_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iAMClient) ListMembershipsForUser(ctx context.Context, in *MembershipsForUserIn, opts ...grpc.CallOption) (*ListMembershipsOut, error) {
	out := new(ListMembershipsOut)
	err := c.cc.Invoke(ctx, IAM_ListMembershipsForUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iAMClient) GetMembership(ctx context.Context, in *GetMembershipIn, opts ...grpc.CallOption) (*GetMembershipOut, error) {
	out := new(GetMembershipOut)
	err := c.cc.Invoke(ctx, IAM_GetMembership_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iAMClient) AddMembership(ctx context.Context, in *AddMembershipIn, opts ...grpc.CallOption) (*AddMembershipOut, error) {
	out := new(AddMembershipOut)
	err := c.cc.Invoke(ctx, IAM_AddMembership_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iAMClient) UpdateMembership(ctx context.Context, in *UpdateMembershipIn, opts ...grpc.CallOption) (*UpdateMembershipOut, error) {
	out := new(UpdateMembershipOut)
	err := c.cc.Invoke(ctx, IAM_UpdateMembership_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iAMClient) RemoveMembership(ctx context.Context, in *RemoveMembershipIn, opts ...grpc.CallOption) (*RemoveMembershipOut, error) {
	out := new(RemoveMembershipOut)
	err := c.cc.Invoke(ctx, IAM_RemoveMembership_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iAMClient) RemoveResource(ctx context.Context, in *RemoveResourceIn, opts ...grpc.CallOption) (*RemoveResourceOut, error) {
	out := new(RemoveResourceOut)
	err := c.cc.Invoke(ctx, IAM_RemoveResource_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IAMServer is the server API for IAM service.
// All implementations must embed UnimplementedIAMServer
// for forward compatibility
type IAMServer interface {
	// Query
	Ping(context.Context, *Message) (*Message, error)
	Can(context.Context, *CanIn) (*CanOut, error)
	ListMembershipsForResource(context.Context, *MembershipsForResourceIn) (*ListMembershipsOut, error)
	ListMembershipsForUser(context.Context, *MembershipsForUserIn) (*ListMembershipsOut, error)
	GetMembership(context.Context, *GetMembershipIn) (*GetMembershipOut, error)
	// Mutation
	AddMembership(context.Context, *AddMembershipIn) (*AddMembershipOut, error)
	UpdateMembership(context.Context, *UpdateMembershipIn) (*UpdateMembershipOut, error)
	RemoveMembership(context.Context, *RemoveMembershipIn) (*RemoveMembershipOut, error)
	RemoveResource(context.Context, *RemoveResourceIn) (*RemoveResourceOut, error)
	mustEmbedUnimplementedIAMServer()
}

// UnimplementedIAMServer must be embedded to have forward compatible implementations.
type UnimplementedIAMServer struct {
}

func (UnimplementedIAMServer) Ping(context.Context, *Message) (*Message, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedIAMServer) Can(context.Context, *CanIn) (*CanOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Can not implemented")
}
func (UnimplementedIAMServer) ListMembershipsForResource(context.Context, *MembershipsForResourceIn) (*ListMembershipsOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListMembershipsForResource not implemented")
}
func (UnimplementedIAMServer) ListMembershipsForUser(context.Context, *MembershipsForUserIn) (*ListMembershipsOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListMembershipsForUser not implemented")
}
func (UnimplementedIAMServer) GetMembership(context.Context, *GetMembershipIn) (*GetMembershipOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMembership not implemented")
}
func (UnimplementedIAMServer) AddMembership(context.Context, *AddMembershipIn) (*AddMembershipOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddMembership not implemented")
}
func (UnimplementedIAMServer) UpdateMembership(context.Context, *UpdateMembershipIn) (*UpdateMembershipOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateMembership not implemented")
}
func (UnimplementedIAMServer) RemoveMembership(context.Context, *RemoveMembershipIn) (*RemoveMembershipOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveMembership not implemented")
}
func (UnimplementedIAMServer) RemoveResource(context.Context, *RemoveResourceIn) (*RemoveResourceOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveResource not implemented")
}
func (UnimplementedIAMServer) mustEmbedUnimplementedIAMServer() {}

// UnsafeIAMServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to IAMServer will
// result in compilation errors.
type UnsafeIAMServer interface {
	mustEmbedUnimplementedIAMServer()
}

func RegisterIAMServer(s grpc.ServiceRegistrar, srv IAMServer) {
	s.RegisterService(&IAM_ServiceDesc, srv)
}

func _IAM_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Message)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IAMServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IAM_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IAMServer).Ping(ctx, req.(*Message))
	}
	return interceptor(ctx, in, info, handler)
}

func _IAM_Can_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CanIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IAMServer).Can(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IAM_Can_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IAMServer).Can(ctx, req.(*CanIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _IAM_ListMembershipsForResource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MembershipsForResourceIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IAMServer).ListMembershipsForResource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IAM_ListMembershipsForResource_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IAMServer).ListMembershipsForResource(ctx, req.(*MembershipsForResourceIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _IAM_ListMembershipsForUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MembershipsForUserIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IAMServer).ListMembershipsForUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IAM_ListMembershipsForUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IAMServer).ListMembershipsForUser(ctx, req.(*MembershipsForUserIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _IAM_GetMembership_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMembershipIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IAMServer).GetMembership(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IAM_GetMembership_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IAMServer).GetMembership(ctx, req.(*GetMembershipIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _IAM_AddMembership_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddMembershipIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IAMServer).AddMembership(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IAM_AddMembership_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IAMServer).AddMembership(ctx, req.(*AddMembershipIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _IAM_UpdateMembership_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateMembershipIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IAMServer).UpdateMembership(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IAM_UpdateMembership_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IAMServer).UpdateMembership(ctx, req.(*UpdateMembershipIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _IAM_RemoveMembership_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveMembershipIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IAMServer).RemoveMembership(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IAM_RemoveMembership_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IAMServer).RemoveMembership(ctx, req.(*RemoveMembershipIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _IAM_RemoveResource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveResourceIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IAMServer).RemoveResource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IAM_RemoveResource_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IAMServer).RemoveResource(ctx, req.(*RemoveResourceIn))
	}
	return interceptor(ctx, in, info, handler)
}

// IAM_ServiceDesc is the grpc.ServiceDesc for IAM service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var IAM_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "IAM",
	HandlerType: (*IAMServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _IAM_Ping_Handler,
		},
		{
			MethodName: "Can",
			Handler:    _IAM_Can_Handler,
		},
		{
			MethodName: "ListMembershipsForResource",
			Handler:    _IAM_ListMembershipsForResource_Handler,
		},
		{
			MethodName: "ListMembershipsForUser",
			Handler:    _IAM_ListMembershipsForUser_Handler,
		},
		{
			MethodName: "GetMembership",
			Handler:    _IAM_GetMembership_Handler,
		},
		{
			MethodName: "AddMembership",
			Handler:    _IAM_AddMembership_Handler,
		},
		{
			MethodName: "UpdateMembership",
			Handler:    _IAM_UpdateMembership_Handler,
		},
		{
			MethodName: "RemoveMembership",
			Handler:    _IAM_RemoveMembership_Handler,
		},
		{
			MethodName: "RemoveResource",
			Handler:    _IAM_RemoveResource_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "iam.proto",
}
