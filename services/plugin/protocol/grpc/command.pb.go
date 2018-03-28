// Code generated by protoc-gen-go. DO NOT EDIT.
// source: command.proto

/*
Package grpc is a generated protocol buffer package.

It is generated from these files:
	command.proto

It has these top-level messages:
	ErrorStatus
	Empty
	CommandRequest
	CommandHelprequest
	CommandHelpResp
*/
package grpc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/any"

import (
	context "golang.org/x/net/context"
	grpc1 "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type ErrorStatus struct {
	Message string                 `protobuf:"bytes,1,opt,name=message" json:"message,omitempty"`
	Details []*google_protobuf.Any `protobuf:"bytes,2,rep,name=details" json:"details,omitempty"`
}

func (m *ErrorStatus) Reset()                    { *m = ErrorStatus{} }
func (m *ErrorStatus) String() string            { return proto.CompactTextString(m) }
func (*ErrorStatus) ProtoMessage()               {}
func (*ErrorStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *ErrorStatus) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *ErrorStatus) GetDetails() []*google_protobuf.Any {
	if m != nil {
		return m.Details
	}
	return nil
}

type Empty struct {
}

func (m *Empty) Reset()                    { *m = Empty{} }
func (m *Empty) String() string            { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()               {}
func (*Empty) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type CommandRequest struct {
	Name   string                 `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Params []*google_protobuf.Any `protobuf:"bytes,2,rep,name=params" json:"params,omitempty"`
}

func (m *CommandRequest) Reset()                    { *m = CommandRequest{} }
func (m *CommandRequest) String() string            { return proto.CompactTextString(m) }
func (*CommandRequest) ProtoMessage()               {}
func (*CommandRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *CommandRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *CommandRequest) GetParams() []*google_protobuf.Any {
	if m != nil {
		return m.Params
	}
	return nil
}

type CommandHelprequest struct {
	Subcommand string `protobuf:"bytes,1,opt,name=subcommand" json:"subcommand,omitempty"`
}

func (m *CommandHelprequest) Reset()                    { *m = CommandHelprequest{} }
func (m *CommandHelprequest) String() string            { return proto.CompactTextString(m) }
func (*CommandHelprequest) ProtoMessage()               {}
func (*CommandHelprequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *CommandHelprequest) GetSubcommand() string {
	if m != nil {
		return m.Subcommand
	}
	return ""
}

type CommandHelpResp struct {
	Help string `protobuf:"bytes,1,opt,name=help" json:"help,omitempty"`
}

func (m *CommandHelpResp) Reset()                    { *m = CommandHelpResp{} }
func (m *CommandHelpResp) String() string            { return proto.CompactTextString(m) }
func (*CommandHelpResp) ProtoMessage()               {}
func (*CommandHelpResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *CommandHelpResp) GetHelp() string {
	if m != nil {
		return m.Help
	}
	return ""
}

func init() {
	proto.RegisterType((*ErrorStatus)(nil), "grpc.ErrorStatus")
	proto.RegisterType((*Empty)(nil), "grpc.Empty")
	proto.RegisterType((*CommandRequest)(nil), "grpc.CommandRequest")
	proto.RegisterType((*CommandHelprequest)(nil), "grpc.CommandHelprequest")
	proto.RegisterType((*CommandHelpResp)(nil), "grpc.CommandHelpResp")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc1.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc1.SupportPackageIsVersion4

// Client API for CommandInterface service

type CommandInterfaceClient interface {
	Command(ctx context.Context, in *CommandRequest, opts ...grpc1.CallOption) (*ErrorStatus, error)
	GetHelp(ctx context.Context, in *CommandHelprequest, opts ...grpc1.CallOption) (*CommandHelpResp, error)
}

type commandInterfaceClient struct {
	cc *grpc1.ClientConn
}

func NewCommandInterfaceClient(cc *grpc1.ClientConn) CommandInterfaceClient {
	return &commandInterfaceClient{cc}
}

func (c *commandInterfaceClient) Command(ctx context.Context, in *CommandRequest, opts ...grpc1.CallOption) (*ErrorStatus, error) {
	out := new(ErrorStatus)
	err := grpc1.Invoke(ctx, "/grpc.CommandInterface/Command", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commandInterfaceClient) GetHelp(ctx context.Context, in *CommandHelprequest, opts ...grpc1.CallOption) (*CommandHelpResp, error) {
	out := new(CommandHelpResp)
	err := grpc1.Invoke(ctx, "/grpc.CommandInterface/GetHelp", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for CommandInterface service

type CommandInterfaceServer interface {
	Command(context.Context, *CommandRequest) (*ErrorStatus, error)
	GetHelp(context.Context, *CommandHelprequest) (*CommandHelpResp, error)
}

func RegisterCommandInterfaceServer(s *grpc1.Server, srv CommandInterfaceServer) {
	s.RegisterService(&_CommandInterface_serviceDesc, srv)
}

func _CommandInterface_Command_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommandRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommandInterfaceServer).Command(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.CommandInterface/Command",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommandInterfaceServer).Command(ctx, req.(*CommandRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommandInterface_GetHelp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommandHelprequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommandInterfaceServer).GetHelp(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.CommandInterface/GetHelp",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommandInterfaceServer).GetHelp(ctx, req.(*CommandHelprequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _CommandInterface_serviceDesc = grpc1.ServiceDesc{
	ServiceName: "grpc.CommandInterface",
	HandlerType: (*CommandInterfaceServer)(nil),
	Methods: []grpc1.MethodDesc{
		{
			MethodName: "Command",
			Handler:    _CommandInterface_Command_Handler,
		},
		{
			MethodName: "GetHelp",
			Handler:    _CommandInterface_GetHelp_Handler,
		},
	},
	Streams:  []grpc1.StreamDesc{},
	Metadata: "command.proto",
}

func init() { proto.RegisterFile("command.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 278 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x90, 0xcf, 0x4b, 0xfc, 0x30,
	0x10, 0xc5, 0xe9, 0xf7, 0xbb, 0x6e, 0x71, 0x16, 0x7f, 0x0d, 0x2b, 0xd4, 0x1e, 0x64, 0x29, 0x08,
	0x7b, 0x90, 0x2c, 0x54, 0x4f, 0xde, 0x44, 0x16, 0xf5, 0x5a, 0x0f, 0x9e, 0xd3, 0xee, 0x6c, 0x15,
	0x9a, 0x26, 0x26, 0xe9, 0xa1, 0x57, 0xff, 0x72, 0x49, 0x9b, 0x40, 0x17, 0xc1, 0xdb, 0xe4, 0xe5,
	0xcd, 0x67, 0xde, 0x0c, 0x9c, 0x54, 0x52, 0x08, 0xde, 0xee, 0x98, 0xd2, 0xd2, 0x4a, 0x9c, 0xd5,
	0x5a, 0x55, 0xe9, 0x55, 0x2d, 0x65, 0xdd, 0xd0, 0x66, 0xd0, 0xca, 0x6e, 0xbf, 0xe1, 0x6d, 0x3f,
	0x1a, 0xb2, 0x77, 0x58, 0x6c, 0xb5, 0x96, 0xfa, 0xcd, 0x72, 0xdb, 0x19, 0x4c, 0x20, 0x16, 0x64,
	0x0c, 0xaf, 0x29, 0x89, 0x56, 0xd1, 0xfa, 0xb8, 0x08, 0x4f, 0x64, 0x10, 0xef, 0xc8, 0xf2, 0xcf,
	0xc6, 0x24, 0xff, 0x56, 0xff, 0xd7, 0x8b, 0x7c, 0xc9, 0x46, 0x2a, 0x0b, 0x54, 0xf6, 0xd8, 0xf6,
	0x45, 0x30, 0x65, 0x31, 0x1c, 0x6d, 0x85, 0xb2, 0x7d, 0x56, 0xc0, 0xe9, 0xd3, 0x98, 0xa9, 0xa0,
	0xaf, 0x8e, 0x8c, 0x45, 0x84, 0x59, 0xcb, 0x45, 0x98, 0x30, 0xd4, 0x78, 0x0b, 0x73, 0xc5, 0x35,
	0x17, 0x7f, 0xd3, 0xbd, 0x27, 0xbb, 0x07, 0xf4, 0xcc, 0x17, 0x6a, 0x94, 0xf6, 0xdc, 0x6b, 0x00,
	0xd3, 0x95, 0xfe, 0x00, 0x9e, 0x3e, 0x51, 0xb2, 0x1b, 0x38, 0x9b, 0x74, 0x15, 0x64, 0x94, 0x8b,
	0xf2, 0x41, 0x8d, 0x0a, 0x51, 0x5c, 0x9d, 0x7f, 0x47, 0x70, 0xee, 0x7d, 0xaf, 0xad, 0x25, 0xbd,
	0xe7, 0x15, 0x61, 0x0e, 0xb1, 0xd7, 0x70, 0xc9, 0xdc, 0x51, 0xd9, 0xe1, 0x52, 0xe9, 0xc5, 0xa8,
	0x4e, 0x8f, 0xf9, 0x00, 0xf1, 0x33, 0x59, 0x37, 0x0b, 0x93, 0x83, 0x9e, 0x49, 0xe8, 0xf4, 0xf2,
	0xd7, 0x8f, 0x0b, 0x56, 0xce, 0x87, 0xbd, 0xef, 0x7e, 0x02, 0x00, 0x00, 0xff, 0xff, 0xc9, 0x40,
	0x07, 0xbf, 0xd0, 0x01, 0x00, 0x00,
}