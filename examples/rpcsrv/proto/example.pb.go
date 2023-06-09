// Code generated by protoc-gen-go. DO NOT EDIT.
// source: example.proto

package proto

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Request struct {
	Value                string   `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()    {}
func (*Request) Descriptor() ([]byte, []int) {
	return fileDescriptor_15a1dc8d40dadaa6, []int{0}
}

func (m *Request) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Request.Unmarshal(m, b)
}
func (m *Request) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Request.Marshal(b, m, deterministic)
}
func (m *Request) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Request.Merge(m, src)
}
func (m *Request) XXX_Size() int {
	return xxx_messageInfo_Request.Size(m)
}
func (m *Request) XXX_DiscardUnknown() {
	xxx_messageInfo_Request.DiscardUnknown(m)
}

var xxx_messageInfo_Request proto.InternalMessageInfo

func (m *Request) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type Response struct {
	Msg                  string   `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Response) Reset()         { *m = Response{} }
func (m *Response) String() string { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()    {}
func (*Response) Descriptor() ([]byte, []int) {
	return fileDescriptor_15a1dc8d40dadaa6, []int{1}
}

func (m *Response) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Response.Unmarshal(m, b)
}
func (m *Response) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Response.Marshal(b, m, deterministic)
}
func (m *Response) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Response.Merge(m, src)
}
func (m *Response) XXX_Size() int {
	return xxx_messageInfo_Response.Size(m)
}
func (m *Response) XXX_DiscardUnknown() {
	xxx_messageInfo_Response.DiscardUnknown(m)
}

var xxx_messageInfo_Response proto.InternalMessageInfo

func (m *Response) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

func init() {
	proto.RegisterType((*Request)(nil), "examples.proto.Request")
	proto.RegisterType((*Response)(nil), "examples.proto.Response")
}

func init() { proto.RegisterFile("example.proto", fileDescriptor_15a1dc8d40dadaa6) }

var fileDescriptor_15a1dc8d40dadaa6 = []byte{
	// 142 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4d, 0xad, 0x48, 0xcc,
	0x2d, 0xc8, 0x49, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x83, 0x72, 0x8b, 0x21, 0x7c,
	0x25, 0x79, 0x2e, 0xf6, 0xa0, 0xd4, 0xc2, 0xd2, 0xd4, 0xe2, 0x12, 0x21, 0x11, 0x2e, 0xd6, 0xb2,
	0xc4, 0x9c, 0xd2, 0x54, 0x09, 0x46, 0x05, 0x46, 0x0d, 0xce, 0x20, 0x08, 0x47, 0x49, 0x86, 0x8b,
	0x23, 0x28, 0xb5, 0xb8, 0x20, 0x3f, 0xaf, 0x38, 0x55, 0x48, 0x80, 0x8b, 0x39, 0xb7, 0x38, 0x1d,
	0x2a, 0x0f, 0x62, 0x1a, 0xb9, 0x71, 0xb1, 0xbb, 0x42, 0x0c, 0x14, 0xb2, 0xe6, 0x62, 0x71, 0x4e,
	0xcc, 0xc9, 0x11, 0x12, 0xd7, 0x43, 0xb5, 0x42, 0x0f, 0x6a, 0xbe, 0x94, 0x04, 0xa6, 0x04, 0xc4,
	0x5c, 0x25, 0x06, 0x27, 0xae, 0x28, 0x0e, 0x3d, 0x3d, 0x7d, 0xb0, 0x70, 0x12, 0x1b, 0x98, 0x32,
	0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0xfc, 0x72, 0xe1, 0x64, 0xba, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ExampleClient is the client API for Example service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ExampleClient interface {
	// gw: GET "/server-rpc/example/call"
	Call(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
}

type exampleClient struct {
	cc *grpc.ClientConn
}

func NewExampleClient(cc *grpc.ClientConn) ExampleClient {
	return &exampleClient{cc}
}

func (c *exampleClient) Call(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/examples.proto.Example/Call", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ExampleServer is the server API for Example service.
type ExampleServer interface {
	// gw: GET "/server-rpc/example/call"
	Call(context.Context, *Request) (*Response, error)
}

// UnimplementedExampleServer can be embedded to have forward compatible implementations.
type UnimplementedExampleServer struct {
}

func (*UnimplementedExampleServer) Call(ctx context.Context, req *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Call not implemented")
}

func RegisterExampleServer(s *grpc.Server, srv ExampleServer) {
	s.RegisterService(&_Example_serviceDesc, srv)
}

func _Example_Call_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExampleServer).Call(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/examples.proto.Example/Call",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExampleServer).Call(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

var _Example_serviceDesc = grpc.ServiceDesc{
	ServiceName: "examples.proto.Example",
	HandlerType: (*ExampleServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Call",
			Handler:    _Example_Call_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "example.proto",
}
