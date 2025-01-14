// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.2
// 	protoc        v5.29.2
// source: proto/registry.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetRegistryIDRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetRegistryIDRequest) Reset() {
	*x = GetRegistryIDRequest{}
	mi := &file_proto_registry_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetRegistryIDRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRegistryIDRequest) ProtoMessage() {}

func (x *GetRegistryIDRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_registry_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRegistryIDRequest.ProtoReflect.Descriptor instead.
func (*GetRegistryIDRequest) Descriptor() ([]byte, []int) {
	return file_proto_registry_proto_rawDescGZIP(), []int{0}
}

type GetRegistryIDResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	RegistryId    string                 `protobuf:"bytes,1,opt,name=registry_id,json=registryId,proto3" json:"registry_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetRegistryIDResponse) Reset() {
	*x = GetRegistryIDResponse{}
	mi := &file_proto_registry_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetRegistryIDResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRegistryIDResponse) ProtoMessage() {}

func (x *GetRegistryIDResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_registry_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRegistryIDResponse.ProtoReflect.Descriptor instead.
func (*GetRegistryIDResponse) Descriptor() ([]byte, []int) {
	return file_proto_registry_proto_rawDescGZIP(), []int{1}
}

func (x *GetRegistryIDResponse) GetRegistryId() string {
	if x != nil {
		return x.RegistryId
	}
	return ""
}

type JoinRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Address       string                 `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`                         // 操作节点的以太坊地址
	Message       string                 `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`                         // 加入请求消息
	Signature     string                 `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`                     // 签名
	SigningKey    string                 `protobuf:"bytes,4,opt,name=signing_key,json=signingKey,proto3" json:"signing_key,omitempty"` // 操作节点的签名公钥
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *JoinRequest) Reset() {
	*x = JoinRequest{}
	mi := &file_proto_registry_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *JoinRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JoinRequest) ProtoMessage() {}

func (x *JoinRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_registry_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JoinRequest.ProtoReflect.Descriptor instead.
func (*JoinRequest) Descriptor() ([]byte, []int) {
	return file_proto_registry_proto_rawDescGZIP(), []int{2}
}

func (x *JoinRequest) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *JoinRequest) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *JoinRequest) GetSignature() string {
	if x != nil {
		return x.Signature
	}
	return ""
}

func (x *JoinRequest) GetSigningKey() string {
	if x != nil {
		return x.SigningKey
	}
	return ""
}

type JoinResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Error         string                 `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"` // 如果失败，返回错误信息
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *JoinResponse) Reset() {
	*x = JoinResponse{}
	mi := &file_proto_registry_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *JoinResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JoinResponse) ProtoMessage() {}

func (x *JoinResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_registry_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JoinResponse.ProtoReflect.Descriptor instead.
func (*JoinResponse) Descriptor() ([]byte, []int) {
	return file_proto_registry_proto_rawDescGZIP(), []int{3}
}

func (x *JoinResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *JoinResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var File_proto_registry_proto protoreflect.FileDescriptor

var file_proto_registry_proto_rawDesc = []byte{
	0x0a, 0x14, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79,
	0x22, 0x16, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x49,
	0x44, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x38, 0x0a, 0x15, 0x47, 0x65, 0x74, 0x52,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x49, 0x44, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79,
	0x49, 0x64, 0x22, 0x80, 0x01, 0x0a, 0x0b, 0x4a, 0x6f, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x18, 0x0a, 0x07,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74,
	0x75, 0x72, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x5f,
	0x6b, 0x65, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x69, 0x67, 0x6e, 0x69,
	0x6e, 0x67, 0x4b, 0x65, 0x79, 0x22, 0x3e, 0x0a, 0x0c, 0x4a, 0x6f, 0x69, 0x6e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12,
	0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x32, 0x97, 0x01, 0x0a, 0x08, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74,
	0x72, 0x79, 0x12, 0x52, 0x0a, 0x0d, 0x47, 0x65, 0x74, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x79, 0x49, 0x44, 0x12, 0x1e, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x47,
	0x65, 0x74, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x49, 0x44, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x47,
	0x65, 0x74, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x49, 0x44, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x37, 0x0a, 0x04, 0x4a, 0x6f, 0x69, 0x6e, 0x12, 0x15,
	0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x4a, 0x6f, 0x69, 0x6e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79,
	0x2e, 0x4a, 0x6f, 0x69, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42,
	0x28, 0x5a, 0x26, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x61,
	0x6c, 0x78, 0x65, 0x2f, 0x73, 0x70, 0x6f, 0x74, 0x74, 0x65, 0x64, 0x2d, 0x6e, 0x65, 0x74, 0x77,
	0x6f, 0x72, 0x6b, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_proto_registry_proto_rawDescOnce sync.Once
	file_proto_registry_proto_rawDescData = file_proto_registry_proto_rawDesc
)

func file_proto_registry_proto_rawDescGZIP() []byte {
	file_proto_registry_proto_rawDescOnce.Do(func() {
		file_proto_registry_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_registry_proto_rawDescData)
	})
	return file_proto_registry_proto_rawDescData
}

var file_proto_registry_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_proto_registry_proto_goTypes = []any{
	(*GetRegistryIDRequest)(nil),  // 0: registry.GetRegistryIDRequest
	(*GetRegistryIDResponse)(nil), // 1: registry.GetRegistryIDResponse
	(*JoinRequest)(nil),           // 2: registry.JoinRequest
	(*JoinResponse)(nil),          // 3: registry.JoinResponse
}
var file_proto_registry_proto_depIdxs = []int32{
	0, // 0: registry.Registry.GetRegistryID:input_type -> registry.GetRegistryIDRequest
	2, // 1: registry.Registry.Join:input_type -> registry.JoinRequest
	1, // 2: registry.Registry.GetRegistryID:output_type -> registry.GetRegistryIDResponse
	3, // 3: registry.Registry.Join:output_type -> registry.JoinResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_registry_proto_init() }
func file_proto_registry_proto_init() {
	if File_proto_registry_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_registry_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_registry_proto_goTypes,
		DependencyIndexes: file_proto_registry_proto_depIdxs,
		MessageInfos:      file_proto_registry_proto_msgTypes,
	}.Build()
	File_proto_registry_proto = out.File
	file_proto_registry_proto_rawDesc = nil
	file_proto_registry_proto_goTypes = nil
	file_proto_registry_proto_depIdxs = nil
}
