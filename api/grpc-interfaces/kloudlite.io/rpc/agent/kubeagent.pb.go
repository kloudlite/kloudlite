// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v3.21.12
// source: kubeagent.proto

package agent

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PayloadIn struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Action        string                 `protobuf:"bytes,1,opt,name=Action,proto3" json:"Action,omitempty"`
	Payload       map[string]*anypb.Any  `protobuf:"bytes,2,rep,name=payload,proto3" json:"payload,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	AccountId     string                 `protobuf:"bytes,3,opt,name=accountId,proto3" json:"accountId,omitempty"`
	ResourceRef   string                 `protobuf:"bytes,4,opt,name=ResourceRef,proto3" json:"ResourceRef,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PayloadIn) Reset() {
	*x = PayloadIn{}
	mi := &file_kubeagent_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PayloadIn) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PayloadIn) ProtoMessage() {}

func (x *PayloadIn) ProtoReflect() protoreflect.Message {
	mi := &file_kubeagent_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PayloadIn.ProtoReflect.Descriptor instead.
func (*PayloadIn) Descriptor() ([]byte, []int) {
	return file_kubeagent_proto_rawDescGZIP(), []int{0}
}

func (x *PayloadIn) GetAction() string {
	if x != nil {
		return x.Action
	}
	return ""
}

func (x *PayloadIn) GetPayload() map[string]*anypb.Any {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *PayloadIn) GetAccountId() string {
	if x != nil {
		return x.AccountId
	}
	return ""
}

func (x *PayloadIn) GetResourceRef() string {
	if x != nil {
		return x.ResourceRef
	}
	return ""
}

type PayloadOut struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Stdout        string                 `protobuf:"bytes,2,opt,name=stdout,proto3" json:"stdout,omitempty"`
	Stderr        string                 `protobuf:"bytes,3,opt,name=stderr,proto3" json:"stderr,omitempty"`
	ExecErr       string                 `protobuf:"bytes,4,opt,name=execErr,proto3" json:"execErr,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PayloadOut) Reset() {
	*x = PayloadOut{}
	mi := &file_kubeagent_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PayloadOut) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PayloadOut) ProtoMessage() {}

func (x *PayloadOut) ProtoReflect() protoreflect.Message {
	mi := &file_kubeagent_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PayloadOut.ProtoReflect.Descriptor instead.
func (*PayloadOut) Descriptor() ([]byte, []int) {
	return file_kubeagent_proto_rawDescGZIP(), []int{1}
}

func (x *PayloadOut) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *PayloadOut) GetStdout() string {
	if x != nil {
		return x.Stdout
	}
	return ""
}

func (x *PayloadOut) GetStderr() string {
	if x != nil {
		return x.Stderr
	}
	return ""
}

func (x *PayloadOut) GetExecErr() string {
	if x != nil {
		return x.ExecErr
	}
	return ""
}

var File_kubeagent_proto protoreflect.FileDescriptor

const file_kubeagent_proto_rawDesc = "" +
	"\n" +
	"\x0fkubeagent.proto\x1a\x19google/protobuf/any.proto\"\xe8\x01\n" +
	"\tPayloadIn\x12\x16\n" +
	"\x06Action\x18\x01 \x01(\tR\x06Action\x121\n" +
	"\apayload\x18\x02 \x03(\v2\x17.PayloadIn.PayloadEntryR\apayload\x12\x1c\n" +
	"\taccountId\x18\x03 \x01(\tR\taccountId\x12 \n" +
	"\vResourceRef\x18\x04 \x01(\tR\vResourceRef\x1aP\n" +
	"\fPayloadEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12*\n" +
	"\x05value\x18\x02 \x01(\v2\x14.google.protobuf.AnyR\x05value:\x028\x01\"p\n" +
	"\n" +
	"PayloadOut\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\x12\x16\n" +
	"\x06stdout\x18\x02 \x01(\tR\x06stdout\x12\x16\n" +
	"\x06stderr\x18\x03 \x01(\tR\x06stderr\x12\x18\n" +
	"\aexecErr\x18\x04 \x01(\tR\aexecErr21\n" +
	"\tKubeAgent\x12$\n" +
	"\tKubeApply\x12\n" +
	".PayloadIn\x1a\v.PayloadOutB\x18Z\x16kloudlite.io/rpc/agentb\x06proto3"

var (
	file_kubeagent_proto_rawDescOnce sync.Once
	file_kubeagent_proto_rawDescData []byte
)

func file_kubeagent_proto_rawDescGZIP() []byte {
	file_kubeagent_proto_rawDescOnce.Do(func() {
		file_kubeagent_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_kubeagent_proto_rawDesc), len(file_kubeagent_proto_rawDesc)))
	})
	return file_kubeagent_proto_rawDescData
}

var file_kubeagent_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_kubeagent_proto_goTypes = []any{
	(*PayloadIn)(nil),  // 0: PayloadIn
	(*PayloadOut)(nil), // 1: PayloadOut
	nil,                // 2: PayloadIn.PayloadEntry
	(*anypb.Any)(nil),  // 3: google.protobuf.Any
}
var file_kubeagent_proto_depIdxs = []int32{
	2, // 0: PayloadIn.payload:type_name -> PayloadIn.PayloadEntry
	3, // 1: PayloadIn.PayloadEntry.value:type_name -> google.protobuf.Any
	0, // 2: KubeAgent.KubeApply:input_type -> PayloadIn
	1, // 3: KubeAgent.KubeApply:output_type -> PayloadOut
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_kubeagent_proto_init() }
func file_kubeagent_proto_init() {
	if File_kubeagent_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_kubeagent_proto_rawDesc), len(file_kubeagent_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_kubeagent_proto_goTypes,
		DependencyIndexes: file_kubeagent_proto_depIdxs,
		MessageInfos:      file_kubeagent_proto_msgTypes,
	}.Build()
	File_kubeagent_proto = out.File
	file_kubeagent_proto_goTypes = nil
	file_kubeagent_proto_depIdxs = nil
}
