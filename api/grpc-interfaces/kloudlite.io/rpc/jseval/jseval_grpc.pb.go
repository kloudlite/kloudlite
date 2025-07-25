// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.4
// source: jseval.proto

package jseval

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
	JSEval_Eval_FullMethodName = "/JSEval/Eval"
)

// JSEvalClient is the client API for JSEval service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type JSEvalClient interface {
	Eval(ctx context.Context, in *EvalIn, opts ...grpc.CallOption) (*EvalOut, error)
}

type jSEvalClient struct {
	cc grpc.ClientConnInterface
}

func NewJSEvalClient(cc grpc.ClientConnInterface) JSEvalClient {
	return &jSEvalClient{cc}
}

func (c *jSEvalClient) Eval(ctx context.Context, in *EvalIn, opts ...grpc.CallOption) (*EvalOut, error) {
	out := new(EvalOut)
	err := c.cc.Invoke(ctx, JSEval_Eval_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// JSEvalServer is the server API for JSEval service.
// All implementations must embed UnimplementedJSEvalServer
// for forward compatibility
type JSEvalServer interface {
	Eval(context.Context, *EvalIn) (*EvalOut, error)
	mustEmbedUnimplementedJSEvalServer()
}

// UnimplementedJSEvalServer must be embedded to have forward compatible implementations.
type UnimplementedJSEvalServer struct {
}

func (UnimplementedJSEvalServer) Eval(context.Context, *EvalIn) (*EvalOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Eval not implemented")
}
func (UnimplementedJSEvalServer) mustEmbedUnimplementedJSEvalServer() {}

// UnsafeJSEvalServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to JSEvalServer will
// result in compilation errors.
type UnsafeJSEvalServer interface {
	mustEmbedUnimplementedJSEvalServer()
}

func RegisterJSEvalServer(s grpc.ServiceRegistrar, srv JSEvalServer) {
	s.RegisterService(&JSEval_ServiceDesc, srv)
}

func _JSEval_Eval_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EvalIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JSEvalServer).Eval(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: JSEval_Eval_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JSEvalServer).Eval(ctx, req.(*EvalIn))
	}
	return interceptor(ctx, in, info, handler)
}

// JSEval_ServiceDesc is the grpc.ServiceDesc for JSEval service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var JSEval_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "JSEval",
	HandlerType: (*JSEvalServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Eval",
			Handler:    _JSEval_Eval_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "jseval.proto",
}
