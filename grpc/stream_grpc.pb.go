// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.11
// source: stream.proto

package grpc

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

// FreedomClient is the client API for Freedom service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FreedomClient interface {
	Pipe(ctx context.Context, opts ...grpc.CallOption) (Freedom_PipeClient, error)
}

type freedomClient struct {
	cc grpc.ClientConnInterface
}

func NewFreedomClient(cc grpc.ClientConnInterface) FreedomClient {
	return &freedomClient{cc}
}

func (c *freedomClient) Pipe(ctx context.Context, opts ...grpc.CallOption) (Freedom_PipeClient, error) {
	stream, err := c.cc.NewStream(ctx, &Freedom_ServiceDesc.Streams[0], "/freedomGo.grpc.Freedom/Pipe", opts...)
	if err != nil {
		return nil, err
	}
	x := &freedomPipeClient{stream}
	return x, nil
}

type Freedom_PipeClient interface {
	Send(*FreedomRequest) error
	Recv() (*FreedomResponse, error)
	grpc.ClientStream
}

type freedomPipeClient struct {
	grpc.ClientStream
}

func (x *freedomPipeClient) Send(m *FreedomRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *freedomPipeClient) Recv() (*FreedomResponse, error) {
	m := new(FreedomResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// FreedomServer is the server API for Freedom service.
// All implementations must embed UnimplementedFreedomServer
// for forward compatibility
type FreedomServer interface {
	Pipe(Freedom_PipeServer) error
	mustEmbedUnimplementedFreedomServer()
}

// UnimplementedFreedomServer must be embedded to have forward compatible implementations.
type UnimplementedFreedomServer struct {
}

func (UnimplementedFreedomServer) Pipe(Freedom_PipeServer) error {
	return status.Errorf(codes.Unimplemented, "method Pipe not implemented")
}
func (UnimplementedFreedomServer) mustEmbedUnimplementedFreedomServer() {}

// UnsafeFreedomServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FreedomServer will
// result in compilation errors.
type UnsafeFreedomServer interface {
	mustEmbedUnimplementedFreedomServer()
}

func RegisterFreedomServer(s grpc.ServiceRegistrar, srv FreedomServer) {
	s.RegisterService(&Freedom_ServiceDesc, srv)
}

func _Freedom_Pipe_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(FreedomServer).Pipe(&freedomPipeServer{stream})
}

type Freedom_PipeServer interface {
	Send(*FreedomResponse) error
	Recv() (*FreedomRequest, error)
	grpc.ServerStream
}

type freedomPipeServer struct {
	grpc.ServerStream
}

func (x *freedomPipeServer) Send(m *FreedomResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *freedomPipeServer) Recv() (*FreedomRequest, error) {
	m := new(FreedomRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Freedom_ServiceDesc is the grpc.ServiceDesc for Freedom service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Freedom_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "freedomGo.grpc.Freedom",
	HandlerType: (*FreedomServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Pipe",
			Handler:       _Freedom_Pipe_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "stream.proto",
}