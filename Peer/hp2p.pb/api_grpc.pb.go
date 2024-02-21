// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.23.4
// source: api.proto

package hp2p_pb

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
	Hp2PApiProto_Ready_FullMethodName              = "/hp2pApiProto.Hp2pApiProto/Ready"
	Hp2PApiProto_Heartbeat_FullMethodName          = "/hp2pApiProto.Hp2pApiProto/Heartbeat"
	Hp2PApiProto_SessionChange_FullMethodName      = "/hp2pApiProto.Hp2pApiProto/SessionChange"
	Hp2PApiProto_SessionTermination_FullMethodName = "/hp2pApiProto.Hp2pApiProto/SessionTermination"
	Hp2PApiProto_PeerChange_FullMethodName         = "/hp2pApiProto.Hp2pApiProto/PeerChange"
	Hp2PApiProto_Expulsion_FullMethodName          = "/hp2pApiProto.Hp2pApiProto/Expulsion"
	Hp2PApiProto_Data_FullMethodName               = "/hp2pApiProto.Hp2pApiProto/Data"
	Hp2PApiProto_Homp_FullMethodName               = "/hp2pApiProto.Hp2pApiProto/Homp"
)

// Hp2PApiProtoClient is the client API for Hp2PApiProto service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type Hp2PApiProtoClient interface {
	Ready(ctx context.Context, in *ReadyRequest, opts ...grpc.CallOption) (*ReadyResponse, error)
	Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error)
	SessionChange(ctx context.Context, in *SessionChangeRequest, opts ...grpc.CallOption) (*SessionChangeResponse, error)
	SessionTermination(ctx context.Context, in *SessionTerminateRequest, opts ...grpc.CallOption) (*SessionTerminateResponse, error)
	PeerChange(ctx context.Context, in *PeerChangeRequest, opts ...grpc.CallOption) (*PeerChangeResponse, error)
	Expulsion(ctx context.Context, in *ExpulsionRequest, opts ...grpc.CallOption) (*ExpulsionResponse, error)
	Data(ctx context.Context, in *DataRequest, opts ...grpc.CallOption) (*DataResponse, error)
	Homp(ctx context.Context, opts ...grpc.CallOption) (Hp2PApiProto_HompClient, error)
}

type hp2PApiProtoClient struct {
	cc grpc.ClientConnInterface
}

func NewHp2PApiProtoClient(cc grpc.ClientConnInterface) Hp2PApiProtoClient {
	return &hp2PApiProtoClient{cc}
}

func (c *hp2PApiProtoClient) Ready(ctx context.Context, in *ReadyRequest, opts ...grpc.CallOption) (*ReadyResponse, error) {
	out := new(ReadyResponse)
	err := c.cc.Invoke(ctx, Hp2PApiProto_Ready_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hp2PApiProtoClient) Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error) {
	out := new(HeartbeatResponse)
	err := c.cc.Invoke(ctx, Hp2PApiProto_Heartbeat_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hp2PApiProtoClient) SessionChange(ctx context.Context, in *SessionChangeRequest, opts ...grpc.CallOption) (*SessionChangeResponse, error) {
	out := new(SessionChangeResponse)
	err := c.cc.Invoke(ctx, Hp2PApiProto_SessionChange_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hp2PApiProtoClient) SessionTermination(ctx context.Context, in *SessionTerminateRequest, opts ...grpc.CallOption) (*SessionTerminateResponse, error) {
	out := new(SessionTerminateResponse)
	err := c.cc.Invoke(ctx, Hp2PApiProto_SessionTermination_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hp2PApiProtoClient) PeerChange(ctx context.Context, in *PeerChangeRequest, opts ...grpc.CallOption) (*PeerChangeResponse, error) {
	out := new(PeerChangeResponse)
	err := c.cc.Invoke(ctx, Hp2PApiProto_PeerChange_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hp2PApiProtoClient) Expulsion(ctx context.Context, in *ExpulsionRequest, opts ...grpc.CallOption) (*ExpulsionResponse, error) {
	out := new(ExpulsionResponse)
	err := c.cc.Invoke(ctx, Hp2PApiProto_Expulsion_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hp2PApiProtoClient) Data(ctx context.Context, in *DataRequest, opts ...grpc.CallOption) (*DataResponse, error) {
	out := new(DataResponse)
	err := c.cc.Invoke(ctx, Hp2PApiProto_Data_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hp2PApiProtoClient) Homp(ctx context.Context, opts ...grpc.CallOption) (Hp2PApiProto_HompClient, error) {
	stream, err := c.cc.NewStream(ctx, &Hp2PApiProto_ServiceDesc.Streams[0], Hp2PApiProto_Homp_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &hp2PApiProtoHompClient{stream}
	return x, nil
}

type Hp2PApiProto_HompClient interface {
	Send(*Response) error
	Recv() (*Request, error)
	grpc.ClientStream
}

type hp2PApiProtoHompClient struct {
	grpc.ClientStream
}

func (x *hp2PApiProtoHompClient) Send(m *Response) error {
	return x.ClientStream.SendMsg(m)
}

func (x *hp2PApiProtoHompClient) Recv() (*Request, error) {
	m := new(Request)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Hp2PApiProtoServer is the server API for Hp2PApiProto service.
// All implementations must embed UnimplementedHp2PApiProtoServer
// for forward compatibility
type Hp2PApiProtoServer interface {
	Ready(context.Context, *ReadyRequest) (*ReadyResponse, error)
	Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
	SessionChange(context.Context, *SessionChangeRequest) (*SessionChangeResponse, error)
	SessionTermination(context.Context, *SessionTerminateRequest) (*SessionTerminateResponse, error)
	PeerChange(context.Context, *PeerChangeRequest) (*PeerChangeResponse, error)
	Expulsion(context.Context, *ExpulsionRequest) (*ExpulsionResponse, error)
	Data(context.Context, *DataRequest) (*DataResponse, error)
	Homp(Hp2PApiProto_HompServer) error
	mustEmbedUnimplementedHp2PApiProtoServer()
}

// UnimplementedHp2PApiProtoServer must be embedded to have forward compatible implementations.
type UnimplementedHp2PApiProtoServer struct {
}

func (UnimplementedHp2PApiProtoServer) Ready(context.Context, *ReadyRequest) (*ReadyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ready not implemented")
}
func (UnimplementedHp2PApiProtoServer) Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Heartbeat not implemented")
}
func (UnimplementedHp2PApiProtoServer) SessionChange(context.Context, *SessionChangeRequest) (*SessionChangeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SessionChange not implemented")
}
func (UnimplementedHp2PApiProtoServer) SessionTermination(context.Context, *SessionTerminateRequest) (*SessionTerminateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SessionTermination not implemented")
}
func (UnimplementedHp2PApiProtoServer) PeerChange(context.Context, *PeerChangeRequest) (*PeerChangeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PeerChange not implemented")
}
func (UnimplementedHp2PApiProtoServer) Expulsion(context.Context, *ExpulsionRequest) (*ExpulsionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Expulsion not implemented")
}
func (UnimplementedHp2PApiProtoServer) Data(context.Context, *DataRequest) (*DataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Data not implemented")
}
func (UnimplementedHp2PApiProtoServer) Homp(Hp2PApiProto_HompServer) error {
	return status.Errorf(codes.Unimplemented, "method Homp not implemented")
}
func (UnimplementedHp2PApiProtoServer) mustEmbedUnimplementedHp2PApiProtoServer() {}

// UnsafeHp2PApiProtoServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to Hp2PApiProtoServer will
// result in compilation errors.
type UnsafeHp2PApiProtoServer interface {
	mustEmbedUnimplementedHp2PApiProtoServer()
}

func RegisterHp2PApiProtoServer(s grpc.ServiceRegistrar, srv Hp2PApiProtoServer) {
	s.RegisterService(&Hp2PApiProto_ServiceDesc, srv)
}

func _Hp2PApiProto_Ready_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(Hp2PApiProtoServer).Ready(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Hp2PApiProto_Ready_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(Hp2PApiProtoServer).Ready(ctx, req.(*ReadyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hp2PApiProto_Heartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartbeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(Hp2PApiProtoServer).Heartbeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Hp2PApiProto_Heartbeat_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(Hp2PApiProtoServer).Heartbeat(ctx, req.(*HeartbeatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hp2PApiProto_SessionChange_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SessionChangeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(Hp2PApiProtoServer).SessionChange(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Hp2PApiProto_SessionChange_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(Hp2PApiProtoServer).SessionChange(ctx, req.(*SessionChangeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hp2PApiProto_SessionTermination_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SessionTerminateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(Hp2PApiProtoServer).SessionTermination(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Hp2PApiProto_SessionTermination_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(Hp2PApiProtoServer).SessionTermination(ctx, req.(*SessionTerminateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hp2PApiProto_PeerChange_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PeerChangeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(Hp2PApiProtoServer).PeerChange(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Hp2PApiProto_PeerChange_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(Hp2PApiProtoServer).PeerChange(ctx, req.(*PeerChangeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hp2PApiProto_Expulsion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExpulsionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(Hp2PApiProtoServer).Expulsion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Hp2PApiProto_Expulsion_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(Hp2PApiProtoServer).Expulsion(ctx, req.(*ExpulsionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hp2PApiProto_Data_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(Hp2PApiProtoServer).Data(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Hp2PApiProto_Data_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(Hp2PApiProtoServer).Data(ctx, req.(*DataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hp2PApiProto_Homp_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(Hp2PApiProtoServer).Homp(&hp2PApiProtoHompServer{stream})
}

type Hp2PApiProto_HompServer interface {
	Send(*Request) error
	Recv() (*Response, error)
	grpc.ServerStream
}

type hp2PApiProtoHompServer struct {
	grpc.ServerStream
}

func (x *hp2PApiProtoHompServer) Send(m *Request) error {
	return x.ServerStream.SendMsg(m)
}

func (x *hp2PApiProtoHompServer) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Hp2PApiProto_ServiceDesc is the grpc.ServiceDesc for Hp2PApiProto service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Hp2PApiProto_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "hp2pApiProto.Hp2pApiProto",
	HandlerType: (*Hp2PApiProtoServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ready",
			Handler:    _Hp2PApiProto_Ready_Handler,
		},
		{
			MethodName: "Heartbeat",
			Handler:    _Hp2PApiProto_Heartbeat_Handler,
		},
		{
			MethodName: "SessionChange",
			Handler:    _Hp2PApiProto_SessionChange_Handler,
		},
		{
			MethodName: "SessionTermination",
			Handler:    _Hp2PApiProto_SessionTermination_Handler,
		},
		{
			MethodName: "PeerChange",
			Handler:    _Hp2PApiProto_PeerChange_Handler,
		},
		{
			MethodName: "Expulsion",
			Handler:    _Hp2PApiProto_Expulsion_Handler,
		},
		{
			MethodName: "Data",
			Handler:    _Hp2PApiProto_Data_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Homp",
			Handler:       _Hp2PApiProto_Homp_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "api.proto",
}