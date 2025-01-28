// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.2
// 	protoc        v5.29.2
// source: task.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// TaskResponseMessage represents a task response message
type TaskResponseMessage struct {
	state           protoimpl.MessageState `protogen:"open.v1"`
	TaskId          string                 `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	OperatorAddress string                 `protobuf:"bytes,2,opt,name=operator_address,json=operatorAddress,proto3" json:"operator_address,omitempty"`
	SigningKey      string                 `protobuf:"bytes,3,opt,name=signing_key,json=signingKey,proto3" json:"signing_key,omitempty"`
	Signature       []byte                 `protobuf:"bytes,4,opt,name=signature,proto3" json:"signature,omitempty"`
	Epoch           uint32                 `protobuf:"varint,5,opt,name=epoch,proto3" json:"epoch,omitempty"`
	ChainId         uint32                 `protobuf:"varint,6,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	TargetAddress   string                 `protobuf:"bytes,7,opt,name=target_address,json=targetAddress,proto3" json:"target_address,omitempty"`
	Key             string                 `protobuf:"bytes,8,opt,name=key,proto3" json:"key,omitempty"`
	Value           string                 `protobuf:"bytes,9,opt,name=value,proto3" json:"value,omitempty"`
	BlockNumber     uint64                 `protobuf:"varint,10,opt,name=block_number,json=blockNumber,proto3" json:"block_number,omitempty"`
	Timestamp       uint64                 `protobuf:"varint,11,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *TaskResponseMessage) Reset() {
	*x = TaskResponseMessage{}
	mi := &file_task_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TaskResponseMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskResponseMessage) ProtoMessage() {}

func (x *TaskResponseMessage) ProtoReflect() protoreflect.Message {
	mi := &file_task_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskResponseMessage.ProtoReflect.Descriptor instead.
func (*TaskResponseMessage) Descriptor() ([]byte, []int) {
	return file_task_proto_rawDescGZIP(), []int{0}
}

func (x *TaskResponseMessage) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

func (x *TaskResponseMessage) GetOperatorAddress() string {
	if x != nil {
		return x.OperatorAddress
	}
	return ""
}

func (x *TaskResponseMessage) GetSigningKey() string {
	if x != nil {
		return x.SigningKey
	}
	return ""
}

func (x *TaskResponseMessage) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *TaskResponseMessage) GetEpoch() uint32 {
	if x != nil {
		return x.Epoch
	}
	return 0
}

func (x *TaskResponseMessage) GetChainId() uint32 {
	if x != nil {
		return x.ChainId
	}
	return 0
}

func (x *TaskResponseMessage) GetTargetAddress() string {
	if x != nil {
		return x.TargetAddress
	}
	return ""
}

func (x *TaskResponseMessage) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *TaskResponseMessage) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *TaskResponseMessage) GetBlockNumber() uint64 {
	if x != nil {
		return x.BlockNumber
	}
	return 0
}

func (x *TaskResponseMessage) GetTimestamp() uint64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

// TaskMessage represents a task message
type TaskMessage struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TaskId        string                 `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	ChainId       uint32                 `protobuf:"varint,2,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	TargetAddress string                 `protobuf:"bytes,3,opt,name=target_address,json=targetAddress,proto3" json:"target_address,omitempty"`
	Key           string                 `protobuf:"bytes,4,opt,name=key,proto3" json:"key,omitempty"`
	BlockNumber   uint64                 `protobuf:"varint,5,opt,name=block_number,json=blockNumber,proto3" json:"block_number,omitempty"`
	Timestamp     uint64                 `protobuf:"varint,6,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Epoch         uint32                 `protobuf:"varint,7,opt,name=epoch,proto3" json:"epoch,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TaskMessage) Reset() {
	*x = TaskMessage{}
	mi := &file_task_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TaskMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskMessage) ProtoMessage() {}

func (x *TaskMessage) ProtoReflect() protoreflect.Message {
	mi := &file_task_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskMessage.ProtoReflect.Descriptor instead.
func (*TaskMessage) Descriptor() ([]byte, []int) {
	return file_task_proto_rawDescGZIP(), []int{1}
}

func (x *TaskMessage) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

func (x *TaskMessage) GetChainId() uint32 {
	if x != nil {
		return x.ChainId
	}
	return 0
}

func (x *TaskMessage) GetTargetAddress() string {
	if x != nil {
		return x.TargetAddress
	}
	return ""
}

func (x *TaskMessage) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *TaskMessage) GetBlockNumber() uint64 {
	if x != nil {
		return x.BlockNumber
	}
	return 0
}

func (x *TaskMessage) GetTimestamp() uint64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *TaskMessage) GetEpoch() uint32 {
	if x != nil {
		return x.Epoch
	}
	return 0
}

// ConsensusReachedMessage represents a consensus reached message
type ConsensusReachedMessage struct {
	state                protoimpl.MessageState `protogen:"open.v1"`
	TaskId               string                 `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	Epoch                uint32                 `protobuf:"varint,2,opt,name=epoch,proto3" json:"epoch,omitempty"`
	OperatorSignatures   map[string][]byte      `protobuf:"bytes,3,rep,name=operator_signatures,json=operatorSignatures,proto3" json:"operator_signatures,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"` // operator_addr -> signature
	AggregatedSignatures []byte                 `protobuf:"bytes,4,opt,name=aggregated_signatures,json=aggregatedSignatures,proto3" json:"aggregated_signatures,omitempty"`
	ConsensusReachedAt   *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=consensus_reached_at,json=consensusReachedAt,proto3" json:"consensus_reached_at,omitempty"`
	unknownFields        protoimpl.UnknownFields
	sizeCache            protoimpl.SizeCache
}

func (x *ConsensusReachedMessage) Reset() {
	*x = ConsensusReachedMessage{}
	mi := &file_task_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConsensusReachedMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConsensusReachedMessage) ProtoMessage() {}

func (x *ConsensusReachedMessage) ProtoReflect() protoreflect.Message {
	mi := &file_task_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConsensusReachedMessage.ProtoReflect.Descriptor instead.
func (*ConsensusReachedMessage) Descriptor() ([]byte, []int) {
	return file_task_proto_rawDescGZIP(), []int{2}
}

func (x *ConsensusReachedMessage) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

func (x *ConsensusReachedMessage) GetEpoch() uint32 {
	if x != nil {
		return x.Epoch
	}
	return 0
}

func (x *ConsensusReachedMessage) GetOperatorSignatures() map[string][]byte {
	if x != nil {
		return x.OperatorSignatures
	}
	return nil
}

func (x *ConsensusReachedMessage) GetAggregatedSignatures() []byte {
	if x != nil {
		return x.AggregatedSignatures
	}
	return nil
}

func (x *ConsensusReachedMessage) GetConsensusReachedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.ConsensusReachedAt
	}
	return nil
}

var File_task_proto protoreflect.FileDescriptor

var file_task_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x74, 0x61, 0x73, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd9, 0x02, 0x0a, 0x13, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x17, 0x0a, 0x07,
	0x74, 0x61, 0x73, 0x6b, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74,
	0x61, 0x73, 0x6b, 0x49, 0x64, 0x12, 0x29, 0x0a, 0x10, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f,
	0x72, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0f, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x5f, 0x6b, 0x65, 0x79, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x4b, 0x65,
	0x79, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12,
	0x14, 0x0a, 0x05, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05,
	0x65, 0x70, 0x6f, 0x63, 0x68, 0x12, 0x19, 0x0a, 0x08, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x69,
	0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x64,
	0x12, 0x25, 0x0a, 0x0e, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74,
	0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x08,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12,
	0x21, 0x0a, 0x0c, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18,
	0x0a, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62,
	0x65, 0x72, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18,
	0x0b, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x22, 0xd1, 0x01, 0x0a, 0x0b, 0x54, 0x61, 0x73, 0x6b, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x17, 0x0a, 0x07, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x74, 0x61, 0x73, 0x6b, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x63, 0x68, 0x61,
	0x69, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x63, 0x68, 0x61,
	0x69, 0x6e, 0x49, 0x64, 0x12, 0x25, 0x0a, 0x0e, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x61,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x74, 0x61,
	0x72, 0x67, 0x65, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x6b,
	0x65, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x21, 0x0a,
	0x0c, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72,
	0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x14,
	0x0a, 0x05, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x65,
	0x70, 0x6f, 0x63, 0x68, 0x22, 0xfb, 0x02, 0x0a, 0x17, 0x43, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73,
	0x75, 0x73, 0x52, 0x65, 0x61, 0x63, 0x68, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x17, 0x0a, 0x07, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x74, 0x61, 0x73, 0x6b, 0x49, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x70, 0x6f,
	0x63, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x12,
	0x67, 0x0a, 0x13, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x73, 0x69, 0x67, 0x6e,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x36, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73, 0x75, 0x73, 0x52, 0x65,
	0x61, 0x63, 0x68, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x4f, 0x70, 0x65,
	0x72, 0x61, 0x74, 0x6f, 0x72, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x12, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x53, 0x69,
	0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x12, 0x33, 0x0a, 0x15, 0x61, 0x67, 0x67, 0x72,
	0x65, 0x67, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x14, 0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61,
	0x74, 0x65, 0x64, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x12, 0x4c, 0x0a,
	0x14, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73, 0x75, 0x73, 0x5f, 0x72, 0x65, 0x61, 0x63, 0x68,
	0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x12, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73,
	0x75, 0x73, 0x52, 0x65, 0x61, 0x63, 0x68, 0x65, 0x64, 0x41, 0x74, 0x1a, 0x45, 0x0a, 0x17, 0x4f,
	0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x42, 0x28, 0x5a, 0x26, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x67, 0x61, 0x6c, 0x78, 0x65, 0x2f, 0x73, 0x70, 0x6f, 0x74, 0x74, 0x65, 0x64, 0x2d, 0x6e,
	0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_task_proto_rawDescOnce sync.Once
	file_task_proto_rawDescData = file_task_proto_rawDesc
)

func file_task_proto_rawDescGZIP() []byte {
	file_task_proto_rawDescOnce.Do(func() {
		file_task_proto_rawDescData = protoimpl.X.CompressGZIP(file_task_proto_rawDescData)
	})
	return file_task_proto_rawDescData
}

var file_task_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_task_proto_goTypes = []any{
	(*TaskResponseMessage)(nil),     // 0: proto.TaskResponseMessage
	(*TaskMessage)(nil),             // 1: proto.TaskMessage
	(*ConsensusReachedMessage)(nil), // 2: proto.ConsensusReachedMessage
	nil,                             // 3: proto.ConsensusReachedMessage.OperatorSignaturesEntry
	(*timestamppb.Timestamp)(nil),   // 4: google.protobuf.Timestamp
}
var file_task_proto_depIdxs = []int32{
	3, // 0: proto.ConsensusReachedMessage.operator_signatures:type_name -> proto.ConsensusReachedMessage.OperatorSignaturesEntry
	4, // 1: proto.ConsensusReachedMessage.consensus_reached_at:type_name -> google.protobuf.Timestamp
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_task_proto_init() }
func file_task_proto_init() {
	if File_task_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_task_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_task_proto_goTypes,
		DependencyIndexes: file_task_proto_depIdxs,
		MessageInfos:      file_task_proto_msgTypes,
	}.Build()
	File_task_proto = out.File
	file_task_proto_rawDesc = nil
	file_task_proto_goTypes = nil
	file_task_proto_depIdxs = nil
}
