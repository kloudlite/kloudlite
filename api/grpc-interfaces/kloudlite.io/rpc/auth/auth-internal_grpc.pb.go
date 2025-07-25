// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.4
// source: auth-internal.proto

package auth

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
	AuthInternal_GetAccessToken_FullMethodName               = "/AuthInternal/GetAccessToken"
	AuthInternal_EnsureUserByEmail_FullMethodName            = "/AuthInternal/EnsureUserByEmail"
	AuthInternal_GetUser_FullMethodName                      = "/AuthInternal/GetUser"
	AuthInternal_GenerateMachineSession_FullMethodName       = "/AuthInternal/GenerateMachineSession"
	AuthInternal_ClearMachineSessionByMachine_FullMethodName = "/AuthInternal/ClearMachineSessionByMachine"
	AuthInternal_ClearMachineSessionByUser_FullMethodName    = "/AuthInternal/ClearMachineSessionByUser"
	AuthInternal_ClearMachineSessionByTeam_FullMethodName    = "/AuthInternal/ClearMachineSessionByTeam"
)

// AuthInternalClient is the client API for AuthInternal service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AuthInternalClient interface {
	GetAccessToken(ctx context.Context, in *GetAccessTokenRequest, opts ...grpc.CallOption) (*AccessTokenOut, error)
	EnsureUserByEmail(ctx context.Context, in *GetUserByEmailRequest, opts ...grpc.CallOption) (*GetUserByEmailOut, error)
	GetUser(ctx context.Context, in *GetUserIn, opts ...grpc.CallOption) (*GetUserOut, error)
	GenerateMachineSession(ctx context.Context, in *GenerateMachineSessionIn, opts ...grpc.CallOption) (*GenerateMachineSessionOut, error)
	ClearMachineSessionByMachine(ctx context.Context, in *ClearMachineSessionByMachineIn, opts ...grpc.CallOption) (*ClearMachineSessionByMachineOut, error)
	ClearMachineSessionByUser(ctx context.Context, in *ClearMachineSessionByUserIn, opts ...grpc.CallOption) (*ClearMachineSessionByUserOut, error)
	ClearMachineSessionByTeam(ctx context.Context, in *ClearMachineSessionByTeamIn, opts ...grpc.CallOption) (*ClearMachineSessionByTeamOut, error)
}

type authInternalClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthInternalClient(cc grpc.ClientConnInterface) AuthInternalClient {
	return &authInternalClient{cc}
}

func (c *authInternalClient) GetAccessToken(ctx context.Context, in *GetAccessTokenRequest, opts ...grpc.CallOption) (*AccessTokenOut, error) {
	out := new(AccessTokenOut)
	err := c.cc.Invoke(ctx, AuthInternal_GetAccessToken_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authInternalClient) EnsureUserByEmail(ctx context.Context, in *GetUserByEmailRequest, opts ...grpc.CallOption) (*GetUserByEmailOut, error) {
	out := new(GetUserByEmailOut)
	err := c.cc.Invoke(ctx, AuthInternal_EnsureUserByEmail_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authInternalClient) GetUser(ctx context.Context, in *GetUserIn, opts ...grpc.CallOption) (*GetUserOut, error) {
	out := new(GetUserOut)
	err := c.cc.Invoke(ctx, AuthInternal_GetUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authInternalClient) GenerateMachineSession(ctx context.Context, in *GenerateMachineSessionIn, opts ...grpc.CallOption) (*GenerateMachineSessionOut, error) {
	out := new(GenerateMachineSessionOut)
	err := c.cc.Invoke(ctx, AuthInternal_GenerateMachineSession_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authInternalClient) ClearMachineSessionByMachine(ctx context.Context, in *ClearMachineSessionByMachineIn, opts ...grpc.CallOption) (*ClearMachineSessionByMachineOut, error) {
	out := new(ClearMachineSessionByMachineOut)
	err := c.cc.Invoke(ctx, AuthInternal_ClearMachineSessionByMachine_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authInternalClient) ClearMachineSessionByUser(ctx context.Context, in *ClearMachineSessionByUserIn, opts ...grpc.CallOption) (*ClearMachineSessionByUserOut, error) {
	out := new(ClearMachineSessionByUserOut)
	err := c.cc.Invoke(ctx, AuthInternal_ClearMachineSessionByUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authInternalClient) ClearMachineSessionByTeam(ctx context.Context, in *ClearMachineSessionByTeamIn, opts ...grpc.CallOption) (*ClearMachineSessionByTeamOut, error) {
	out := new(ClearMachineSessionByTeamOut)
	err := c.cc.Invoke(ctx, AuthInternal_ClearMachineSessionByTeam_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthInternalServer is the server API for AuthInternal service.
// All implementations must embed UnimplementedAuthInternalServer
// for forward compatibility
type AuthInternalServer interface {
	GetAccessToken(context.Context, *GetAccessTokenRequest) (*AccessTokenOut, error)
	EnsureUserByEmail(context.Context, *GetUserByEmailRequest) (*GetUserByEmailOut, error)
	GetUser(context.Context, *GetUserIn) (*GetUserOut, error)
	GenerateMachineSession(context.Context, *GenerateMachineSessionIn) (*GenerateMachineSessionOut, error)
	ClearMachineSessionByMachine(context.Context, *ClearMachineSessionByMachineIn) (*ClearMachineSessionByMachineOut, error)
	ClearMachineSessionByUser(context.Context, *ClearMachineSessionByUserIn) (*ClearMachineSessionByUserOut, error)
	ClearMachineSessionByTeam(context.Context, *ClearMachineSessionByTeamIn) (*ClearMachineSessionByTeamOut, error)
	mustEmbedUnimplementedAuthInternalServer()
}

// UnimplementedAuthInternalServer must be embedded to have forward compatible implementations.
type UnimplementedAuthInternalServer struct {
}

func (UnimplementedAuthInternalServer) GetAccessToken(context.Context, *GetAccessTokenRequest) (*AccessTokenOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAccessToken not implemented")
}
func (UnimplementedAuthInternalServer) EnsureUserByEmail(context.Context, *GetUserByEmailRequest) (*GetUserByEmailOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EnsureUserByEmail not implemented")
}
func (UnimplementedAuthInternalServer) GetUser(context.Context, *GetUserIn) (*GetUserOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}
func (UnimplementedAuthInternalServer) GenerateMachineSession(context.Context, *GenerateMachineSessionIn) (*GenerateMachineSessionOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenerateMachineSession not implemented")
}
func (UnimplementedAuthInternalServer) ClearMachineSessionByMachine(context.Context, *ClearMachineSessionByMachineIn) (*ClearMachineSessionByMachineOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClearMachineSessionByMachine not implemented")
}
func (UnimplementedAuthInternalServer) ClearMachineSessionByUser(context.Context, *ClearMachineSessionByUserIn) (*ClearMachineSessionByUserOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClearMachineSessionByUser not implemented")
}
func (UnimplementedAuthInternalServer) ClearMachineSessionByTeam(context.Context, *ClearMachineSessionByTeamIn) (*ClearMachineSessionByTeamOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClearMachineSessionByTeam not implemented")
}
func (UnimplementedAuthInternalServer) mustEmbedUnimplementedAuthInternalServer() {}

// UnsafeAuthInternalServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuthInternalServer will
// result in compilation errors.
type UnsafeAuthInternalServer interface {
	mustEmbedUnimplementedAuthInternalServer()
}

func RegisterAuthInternalServer(s grpc.ServiceRegistrar, srv AuthInternalServer) {
	s.RegisterService(&AuthInternal_ServiceDesc, srv)
}

func _AuthInternal_GetAccessToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAccessTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthInternalServer).GetAccessToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthInternal_GetAccessToken_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthInternalServer).GetAccessToken(ctx, req.(*GetAccessTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthInternal_EnsureUserByEmail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserByEmailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthInternalServer).EnsureUserByEmail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthInternal_EnsureUserByEmail_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthInternalServer).EnsureUserByEmail(ctx, req.(*GetUserByEmailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthInternal_GetUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthInternalServer).GetUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthInternal_GetUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthInternalServer).GetUser(ctx, req.(*GetUserIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthInternal_GenerateMachineSession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GenerateMachineSessionIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthInternalServer).GenerateMachineSession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthInternal_GenerateMachineSession_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthInternalServer).GenerateMachineSession(ctx, req.(*GenerateMachineSessionIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthInternal_ClearMachineSessionByMachine_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClearMachineSessionByMachineIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthInternalServer).ClearMachineSessionByMachine(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthInternal_ClearMachineSessionByMachine_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthInternalServer).ClearMachineSessionByMachine(ctx, req.(*ClearMachineSessionByMachineIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthInternal_ClearMachineSessionByUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClearMachineSessionByUserIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthInternalServer).ClearMachineSessionByUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthInternal_ClearMachineSessionByUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthInternalServer).ClearMachineSessionByUser(ctx, req.(*ClearMachineSessionByUserIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthInternal_ClearMachineSessionByTeam_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClearMachineSessionByTeamIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthInternalServer).ClearMachineSessionByTeam(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthInternal_ClearMachineSessionByTeam_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthInternalServer).ClearMachineSessionByTeam(ctx, req.(*ClearMachineSessionByTeamIn))
	}
	return interceptor(ctx, in, info, handler)
}

// AuthInternal_ServiceDesc is the grpc.ServiceDesc for AuthInternal service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AuthInternal_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "AuthInternal",
	HandlerType: (*AuthInternalServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAccessToken",
			Handler:    _AuthInternal_GetAccessToken_Handler,
		},
		{
			MethodName: "EnsureUserByEmail",
			Handler:    _AuthInternal_EnsureUserByEmail_Handler,
		},
		{
			MethodName: "GetUser",
			Handler:    _AuthInternal_GetUser_Handler,
		},
		{
			MethodName: "GenerateMachineSession",
			Handler:    _AuthInternal_GenerateMachineSession_Handler,
		},
		{
			MethodName: "ClearMachineSessionByMachine",
			Handler:    _AuthInternal_ClearMachineSessionByMachine_Handler,
		},
		{
			MethodName: "ClearMachineSessionByUser",
			Handler:    _AuthInternal_ClearMachineSessionByUser_Handler,
		},
		{
			MethodName: "ClearMachineSessionByTeam",
			Handler:    _AuthInternal_ClearMachineSessionByTeam_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "auth-internal.proto",
}
