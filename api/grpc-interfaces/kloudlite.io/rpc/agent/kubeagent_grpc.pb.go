// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.4
// source: kubeagent.proto

package agent

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
	KubeAgent_KubeApply_FullMethodName = "/KubeAgent/KubeApply"
)

// KubeAgentClient is the client API for KubeAgent service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type KubeAgentClient interface {
	KubeApply(ctx context.Context, in *PayloadIn, opts ...grpc.CallOption) (*PayloadOut, error)
}

type kubeAgentClient struct {
	cc grpc.ClientConnInterface
}

func NewKubeAgentClient(cc grpc.ClientConnInterface) KubeAgentClient {
	return &kubeAgentClient{cc}
}

func (c *kubeAgentClient) KubeApply(ctx context.Context, in *PayloadIn, opts ...grpc.CallOption) (*PayloadOut, error) {
	out := new(PayloadOut)
	err := c.cc.Invoke(ctx, KubeAgent_KubeApply_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KubeAgentServer is the server API for KubeAgent service.
// All implementations must embed UnimplementedKubeAgentServer
// for forward compatibility
type KubeAgentServer interface {
	KubeApply(context.Context, *PayloadIn) (*PayloadOut, error)
	mustEmbedUnimplementedKubeAgentServer()
}

// UnimplementedKubeAgentServer must be embedded to have forward compatible implementations.
type UnimplementedKubeAgentServer struct {
}

func (UnimplementedKubeAgentServer) KubeApply(context.Context, *PayloadIn) (*PayloadOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method KubeApply not implemented")
}
func (UnimplementedKubeAgentServer) mustEmbedUnimplementedKubeAgentServer() {}

// UnsafeKubeAgentServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to KubeAgentServer will
// result in compilation errors.
type UnsafeKubeAgentServer interface {
	mustEmbedUnimplementedKubeAgentServer()
}

func RegisterKubeAgentServer(s grpc.ServiceRegistrar, srv KubeAgentServer) {
	s.RegisterService(&KubeAgent_ServiceDesc, srv)
}

func _KubeAgent_KubeApply_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PayloadIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KubeAgentServer).KubeApply(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KubeAgent_KubeApply_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KubeAgentServer).KubeApply(ctx, req.(*PayloadIn))
	}
	return interceptor(ctx, in, info, handler)
}

// KubeAgent_ServiceDesc is the grpc.ServiceDesc for KubeAgent service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var KubeAgent_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "KubeAgent",
	HandlerType: (*KubeAgentServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "KubeApply",
			Handler:    _KubeAgent_KubeApply_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kubeagent.proto",
}
