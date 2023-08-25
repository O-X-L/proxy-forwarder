// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package proto

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

// BypassClient is the client API for Bypass service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BypassClient interface {
	Bypass(ctx context.Context, in *BypassRequest, opts ...grpc.CallOption) (*BypassReply, error)
}

type bypassClient struct {
	cc grpc.ClientConnInterface
}

func NewBypassClient(cc grpc.ClientConnInterface) BypassClient {
	return &bypassClient{cc}
}

func (c *bypassClient) Bypass(ctx context.Context, in *BypassRequest, opts ...grpc.CallOption) (*BypassReply, error) {
	out := new(BypassReply)
	err := c.cc.Invoke(ctx, "/proto.Bypass/Bypass", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BypassServer is the server API for Bypass service.
// All implementations must embed UnimplementedBypassServer
// for forward compatibility
type BypassServer interface {
	Bypass(context.Context, *BypassRequest) (*BypassReply, error)
	mustEmbedUnimplementedBypassServer()
}

// UnimplementedBypassServer must be embedded to have forward compatible implementations.
type UnimplementedBypassServer struct {
}

func (UnimplementedBypassServer) Bypass(context.Context, *BypassRequest) (*BypassReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Bypass not implemented")
}
func (UnimplementedBypassServer) mustEmbedUnimplementedBypassServer() {}

// UnsafeBypassServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BypassServer will
// result in compilation errors.
type UnsafeBypassServer interface {
	mustEmbedUnimplementedBypassServer()
}

func RegisterBypassServer(s grpc.ServiceRegistrar, srv BypassServer) {
	s.RegisterService(&Bypass_ServiceDesc, srv)
}

func _Bypass_Bypass_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BypassRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BypassServer).Bypass(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Bypass/Bypass",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BypassServer).Bypass(ctx, req.(*BypassRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Bypass_ServiceDesc is the grpc.ServiceDesc for Bypass service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Bypass_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Bypass",
	HandlerType: (*BypassServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Bypass",
			Handler:    _Bypass_Bypass_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "bypass.proto",
}