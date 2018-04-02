// Code generated by protoc-gen-go. DO NOT EDIT.
// source: plugin_interface.proto

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	plugin_interface.proto

It has these top-level messages:
	ErrorStatus
	DeployRequest
	Empty
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

type ErrorStatus struct {
	// Types that are valid to be assigned to OptionalErr:
	//	*ErrorStatus_Error
	OptionalErr isErrorStatus_OptionalErr `protobuf_oneof:"optional_err"`
}

func (m *ErrorStatus) Reset()                    { *m = ErrorStatus{} }
func (m *ErrorStatus) String() string            { return proto1.CompactTextString(m) }
func (*ErrorStatus) ProtoMessage()               {}
func (*ErrorStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type isErrorStatus_OptionalErr interface {
	isErrorStatus_OptionalErr()
}

type ErrorStatus_Error struct {
	Error string `protobuf:"bytes,2,opt,name=error,oneof"`
}

func (*ErrorStatus_Error) isErrorStatus_OptionalErr() {}

func (m *ErrorStatus) GetOptionalErr() isErrorStatus_OptionalErr {
	if m != nil {
		return m.OptionalErr
	}
	return nil
}

func (m *ErrorStatus) GetError() string {
	if x, ok := m.GetOptionalErr().(*ErrorStatus_Error); ok {
		return x.Error
	}
	return ""
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*ErrorStatus) XXX_OneofFuncs() (func(msg proto1.Message, b *proto1.Buffer) error, func(msg proto1.Message, tag, wire int, b *proto1.Buffer) (bool, error), func(msg proto1.Message) (n int), []interface{}) {
	return _ErrorStatus_OneofMarshaler, _ErrorStatus_OneofUnmarshaler, _ErrorStatus_OneofSizer, []interface{}{
		(*ErrorStatus_Error)(nil),
	}
}

func _ErrorStatus_OneofMarshaler(msg proto1.Message, b *proto1.Buffer) error {
	m := msg.(*ErrorStatus)
	// optional_err
	switch x := m.OptionalErr.(type) {
	case *ErrorStatus_Error:
		b.EncodeVarint(2<<3 | proto1.WireBytes)
		b.EncodeStringBytes(x.Error)
	case nil:
	default:
		return fmt.Errorf("ErrorStatus.OptionalErr has unexpected type %T", x)
	}
	return nil
}

func _ErrorStatus_OneofUnmarshaler(msg proto1.Message, tag, wire int, b *proto1.Buffer) (bool, error) {
	m := msg.(*ErrorStatus)
	switch tag {
	case 2: // optional_err.error
		if wire != proto1.WireBytes {
			return true, proto1.ErrInternalBadWireType
		}
		x, err := b.DecodeStringBytes()
		m.OptionalErr = &ErrorStatus_Error{x}
		return true, err
	default:
		return false, nil
	}
}

func _ErrorStatus_OneofSizer(msg proto1.Message) (n int) {
	m := msg.(*ErrorStatus)
	// optional_err
	switch x := m.OptionalErr.(type) {
	case *ErrorStatus_Error:
		n += proto1.SizeVarint(2<<3 | proto1.WireBytes)
		n += proto1.SizeVarint(uint64(len(x.Error)))
		n += len(x.Error)
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

type DeployRequest struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
}

func (m *DeployRequest) Reset()                    { *m = DeployRequest{} }
func (m *DeployRequest) String() string            { return proto1.CompactTextString(m) }
func (*DeployRequest) ProtoMessage()               {}
func (*DeployRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *DeployRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type Empty struct {
}

func (m *Empty) Reset()                    { *m = Empty{} }
func (m *Empty) String() string            { return proto1.CompactTextString(m) }
func (*Empty) ProtoMessage()               {}
func (*Empty) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func init() {
	proto1.RegisterType((*ErrorStatus)(nil), "proto.ErrorStatus")
	proto1.RegisterType((*DeployRequest)(nil), "proto.DeployRequest")
	proto1.RegisterType((*Empty)(nil), "proto.Empty")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for PluginInterface service

type PluginInterfaceClient interface {
	Initialization(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ErrorStatus, error)
	Deploy(ctx context.Context, in *DeployRequest, opts ...grpc.CallOption) (*ErrorStatus, error)
	UnInitialization(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ErrorStatus, error)
}

type pluginInterfaceClient struct {
	cc *grpc.ClientConn
}

func NewPluginInterfaceClient(cc *grpc.ClientConn) PluginInterfaceClient {
	return &pluginInterfaceClient{cc}
}

func (c *pluginInterfaceClient) Initialization(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ErrorStatus, error) {
	out := new(ErrorStatus)
	err := grpc.Invoke(ctx, "/proto.PluginInterface/Initialization", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginInterfaceClient) Deploy(ctx context.Context, in *DeployRequest, opts ...grpc.CallOption) (*ErrorStatus, error) {
	out := new(ErrorStatus)
	err := grpc.Invoke(ctx, "/proto.PluginInterface/Deploy", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginInterfaceClient) UnInitialization(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ErrorStatus, error) {
	out := new(ErrorStatus)
	err := grpc.Invoke(ctx, "/proto.PluginInterface/UnInitialization", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for PluginInterface service

type PluginInterfaceServer interface {
	Initialization(context.Context, *Empty) (*ErrorStatus, error)
	Deploy(context.Context, *DeployRequest) (*ErrorStatus, error)
	UnInitialization(context.Context, *Empty) (*ErrorStatus, error)
}

func RegisterPluginInterfaceServer(s *grpc.Server, srv PluginInterfaceServer) {
	s.RegisterService(&_PluginInterface_serviceDesc, srv)
}

func _PluginInterface_Initialization_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginInterfaceServer).Initialization(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.PluginInterface/Initialization",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginInterfaceServer).Initialization(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _PluginInterface_Deploy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeployRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginInterfaceServer).Deploy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.PluginInterface/Deploy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginInterfaceServer).Deploy(ctx, req.(*DeployRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PluginInterface_UnInitialization_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginInterfaceServer).UnInitialization(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.PluginInterface/UnInitialization",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginInterfaceServer).UnInitialization(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _PluginInterface_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.PluginInterface",
	HandlerType: (*PluginInterfaceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Initialization",
			Handler:    _PluginInterface_Initialization_Handler,
		},
		{
			MethodName: "Deploy",
			Handler:    _PluginInterface_Deploy_Handler,
		},
		{
			MethodName: "UnInitialization",
			Handler:    _PluginInterface_UnInitialization_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "plugin_interface.proto",
}

func init() { proto1.RegisterFile("plugin_interface.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 208 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2b, 0xc8, 0x29, 0x4d,
	0xcf, 0xcc, 0x8b, 0xcf, 0xcc, 0x2b, 0x49, 0x2d, 0x4a, 0x4b, 0x4c, 0x4e, 0xd5, 0x2b, 0x28, 0xca,
	0x2f, 0xc9, 0x17, 0x62, 0x05, 0x53, 0x4a, 0xa6, 0x5c, 0xdc, 0xae, 0x45, 0x45, 0xf9, 0x45, 0xc1,
	0x25, 0x89, 0x25, 0xa5, 0xc5, 0x42, 0x62, 0x5c, 0xac, 0xa9, 0x20, 0xae, 0x04, 0x93, 0x02, 0xa3,
	0x06, 0xa7, 0x07, 0x43, 0x10, 0x84, 0xeb, 0xc4, 0xc7, 0xc5, 0x93, 0x5f, 0x50, 0x92, 0x99, 0x9f,
	0x97, 0x98, 0x13, 0x9f, 0x5a, 0x54, 0xa4, 0xa4, 0xcc, 0xc5, 0xeb, 0x92, 0x5a, 0x90, 0x93, 0x5f,
	0x19, 0x94, 0x5a, 0x58, 0x9a, 0x5a, 0x5c, 0x22, 0x24, 0xc4, 0xc5, 0x92, 0x97, 0x98, 0x9b, 0x2a,
	0xc1, 0x08, 0xd2, 0x17, 0x04, 0x66, 0x2b, 0xb1, 0x73, 0xb1, 0xba, 0xe6, 0x16, 0x94, 0x54, 0x1a,
	0xad, 0x67, 0xe4, 0xe2, 0x0f, 0x00, 0x3b, 0xc3, 0x13, 0xe6, 0x0a, 0x21, 0x23, 0x2e, 0x3e, 0xcf,
	0xbc, 0xcc, 0x92, 0xcc, 0xc4, 0x9c, 0xcc, 0xaa, 0x44, 0x90, 0xc9, 0x42, 0x3c, 0x10, 0x97, 0xe9,
	0x81, 0xf5, 0x48, 0x09, 0xc1, 0x78, 0x48, 0xae, 0x33, 0xe2, 0x62, 0x83, 0xd8, 0x2a, 0x24, 0x02,
	0x95, 0x45, 0x71, 0x04, 0x56, 0x3d, 0x26, 0x5c, 0x02, 0xa1, 0x79, 0xa4, 0xda, 0x94, 0xc4, 0x06,
	0x16, 0x32, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0x13, 0x0c, 0xa4, 0xc4, 0x3e, 0x01, 0x00, 0x00,
}