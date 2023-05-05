// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.10
// source: pb/vote.proto

package pb

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

// VoteClient is the client API for Vote service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type VoteClient interface {
	VoteRequest(ctx context.Context, in *VoteRequestData, opts ...grpc.CallOption) (*VoteReplyData, error)
}

type voteClient struct {
	cc grpc.ClientConnInterface
}

func NewVoteClient(cc grpc.ClientConnInterface) VoteClient {
	return &voteClient{cc}
}

func (c *voteClient) VoteRequest(ctx context.Context, in *VoteRequestData, opts ...grpc.CallOption) (*VoteReplyData, error) {
	out := new(VoteReplyData)
	err := c.cc.Invoke(ctx, "/pb.Vote/VoteRequest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VoteServer is the server API for Vote service.
// All implementations must embed UnimplementedVoteServer
// for forward compatibility
type VoteServer interface {
	VoteRequest(context.Context, *VoteRequestData) (*VoteReplyData, error)
	mustEmbedUnimplementedVoteServer()
}

// UnimplementedVoteServer must be embedded to have forward compatible implementations.
type UnimplementedVoteServer struct {
}

func (UnimplementedVoteServer) VoteRequest(context.Context, *VoteRequestData) (*VoteReplyData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VoteRequest not implemented")
}
func (UnimplementedVoteServer) mustEmbedUnimplementedVoteServer() {}

// UnsafeVoteServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to VoteServer will
// result in compilation errors.
type UnsafeVoteServer interface {
	mustEmbedUnimplementedVoteServer()
}

func RegisterVoteServer(s grpc.ServiceRegistrar, srv VoteServer) {
	s.RegisterService(&Vote_ServiceDesc, srv)
}

func _Vote_VoteRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VoteRequestData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VoteServer).VoteRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Vote/VoteRequest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VoteServer).VoteRequest(ctx, req.(*VoteRequestData))
	}
	return interceptor(ctx, in, info, handler)
}

// Vote_ServiceDesc is the grpc.ServiceDesc for Vote service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Vote_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Vote",
	HandlerType: (*VoteServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "VoteRequest",
			Handler:    _Vote_VoteRequest_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pb/vote.proto",
}

// EntryClient is the client API for Entry service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EntryClient interface {
	ClientInit(ctx context.Context, in *EntryClientInitDataRequest, opts ...grpc.CallOption) (*EntryClientInitDataReply, error)
}

type entryClient struct {
	cc grpc.ClientConnInterface
}

func NewEntryClient(cc grpc.ClientConnInterface) EntryClient {
	return &entryClient{cc}
}

func (c *entryClient) ClientInit(ctx context.Context, in *EntryClientInitDataRequest, opts ...grpc.CallOption) (*EntryClientInitDataReply, error) {
	out := new(EntryClientInitDataReply)
	err := c.cc.Invoke(ctx, "/pb.Entry/ClientInit", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EntryServer is the server API for Entry service.
// All implementations must embed UnimplementedEntryServer
// for forward compatibility
type EntryServer interface {
	ClientInit(context.Context, *EntryClientInitDataRequest) (*EntryClientInitDataReply, error)
	mustEmbedUnimplementedEntryServer()
}

// UnimplementedEntryServer must be embedded to have forward compatible implementations.
type UnimplementedEntryServer struct {
}

func (UnimplementedEntryServer) ClientInit(context.Context, *EntryClientInitDataRequest) (*EntryClientInitDataReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClientInit not implemented")
}
func (UnimplementedEntryServer) mustEmbedUnimplementedEntryServer() {}

// UnsafeEntryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EntryServer will
// result in compilation errors.
type UnsafeEntryServer interface {
	mustEmbedUnimplementedEntryServer()
}

func RegisterEntryServer(s grpc.ServiceRegistrar, srv EntryServer) {
	s.RegisterService(&Entry_ServiceDesc, srv)
}

func _Entry_ClientInit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EntryClientInitDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EntryServer).ClientInit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Entry/ClientInit",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EntryServer).ClientInit(ctx, req.(*EntryClientInitDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Entry_ServiceDesc is the grpc.ServiceDesc for Entry service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Entry_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Entry",
	HandlerType: (*EntryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ClientInit",
			Handler:    _Entry_ClientInit_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pb/vote.proto",
}
