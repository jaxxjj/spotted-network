// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.2
// source: sync.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	StateSync_GetFullState_FullMethodName     = "/sync.StateSync/GetFullState"
	StateSync_SubscribeUpdates_FullMethodName = "/sync.StateSync/SubscribeUpdates"
)

// StateSyncClient is the client API for StateSync service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// State sync service
type StateSyncClient interface {
	// Get full operator state
	GetFullState(ctx context.Context, in *GetFullStateRequest, opts ...grpc.CallOption) (*GetFullStateResponse, error)
	// Subscribe to state updates
	SubscribeUpdates(ctx context.Context, in *SubscribeRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[OperatorStateUpdate], error)
}

type stateSyncClient struct {
	cc grpc.ClientConnInterface
}

func NewStateSyncClient(cc grpc.ClientConnInterface) StateSyncClient {
	return &stateSyncClient{cc}
}

func (c *stateSyncClient) GetFullState(ctx context.Context, in *GetFullStateRequest, opts ...grpc.CallOption) (*GetFullStateResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetFullStateResponse)
	err := c.cc.Invoke(ctx, StateSync_GetFullState_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stateSyncClient) SubscribeUpdates(ctx context.Context, in *SubscribeRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[OperatorStateUpdate], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &StateSync_ServiceDesc.Streams[0], StateSync_SubscribeUpdates_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[SubscribeRequest, OperatorStateUpdate]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StateSync_SubscribeUpdatesClient = grpc.ServerStreamingClient[OperatorStateUpdate]

// StateSyncServer is the server API for StateSync service.
// All implementations must embed UnimplementedStateSyncServer
// for forward compatibility.
//
// State sync service
type StateSyncServer interface {
	// Get full operator state
	GetFullState(context.Context, *GetFullStateRequest) (*GetFullStateResponse, error)
	// Subscribe to state updates
	SubscribeUpdates(*SubscribeRequest, grpc.ServerStreamingServer[OperatorStateUpdate]) error
	mustEmbedUnimplementedStateSyncServer()
}

// UnimplementedStateSyncServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedStateSyncServer struct{}

func (UnimplementedStateSyncServer) GetFullState(context.Context, *GetFullStateRequest) (*GetFullStateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFullState not implemented")
}
func (UnimplementedStateSyncServer) SubscribeUpdates(*SubscribeRequest, grpc.ServerStreamingServer[OperatorStateUpdate]) error {
	return status.Errorf(codes.Unimplemented, "method SubscribeUpdates not implemented")
}
func (UnimplementedStateSyncServer) mustEmbedUnimplementedStateSyncServer() {}
func (UnimplementedStateSyncServer) testEmbeddedByValue()                   {}

// UnsafeStateSyncServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StateSyncServer will
// result in compilation errors.
type UnsafeStateSyncServer interface {
	mustEmbedUnimplementedStateSyncServer()
}

func RegisterStateSyncServer(s grpc.ServiceRegistrar, srv StateSyncServer) {
	// If the following call pancis, it indicates UnimplementedStateSyncServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&StateSync_ServiceDesc, srv)
}

func _StateSync_GetFullState_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFullStateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StateSyncServer).GetFullState(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: StateSync_GetFullState_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StateSyncServer).GetFullState(ctx, req.(*GetFullStateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StateSync_SubscribeUpdates_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SubscribeRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(StateSyncServer).SubscribeUpdates(m, &grpc.GenericServerStream[SubscribeRequest, OperatorStateUpdate]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StateSync_SubscribeUpdatesServer = grpc.ServerStreamingServer[OperatorStateUpdate]

// StateSync_ServiceDesc is the grpc.ServiceDesc for StateSync service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var StateSync_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "sync.StateSync",
	HandlerType: (*StateSyncServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetFullState",
			Handler:    _StateSync_GetFullState_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SubscribeUpdates",
			Handler:       _StateSync_SubscribeUpdates_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "sync.proto",
}
