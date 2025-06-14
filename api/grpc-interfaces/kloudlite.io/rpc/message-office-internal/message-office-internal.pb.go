// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v3.21.12
// source: message-office-internal.proto

package message_office_internal

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
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

type GenerateClusterTokenIn struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	AccountName   string                 `protobuf:"bytes,1,opt,name=accountName,proto3" json:"accountName,omitempty"`
	ClusterName   string                 `protobuf:"bytes,2,opt,name=clusterName,proto3" json:"clusterName,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GenerateClusterTokenIn) Reset() {
	*x = GenerateClusterTokenIn{}
	mi := &file_message_office_internal_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GenerateClusterTokenIn) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenerateClusterTokenIn) ProtoMessage() {}

func (x *GenerateClusterTokenIn) ProtoReflect() protoreflect.Message {
	mi := &file_message_office_internal_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GenerateClusterTokenIn.ProtoReflect.Descriptor instead.
func (*GenerateClusterTokenIn) Descriptor() ([]byte, []int) {
	return file_message_office_internal_proto_rawDescGZIP(), []int{0}
}

func (x *GenerateClusterTokenIn) GetAccountName() string {
	if x != nil {
		return x.AccountName
	}
	return ""
}

func (x *GenerateClusterTokenIn) GetClusterName() string {
	if x != nil {
		return x.ClusterName
	}
	return ""
}

type GenerateClusterTokenOut struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ClusterToken  string                 `protobuf:"bytes,1,opt,name=clusterToken,proto3" json:"clusterToken,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GenerateClusterTokenOut) Reset() {
	*x = GenerateClusterTokenOut{}
	mi := &file_message_office_internal_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GenerateClusterTokenOut) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenerateClusterTokenOut) ProtoMessage() {}

func (x *GenerateClusterTokenOut) ProtoReflect() protoreflect.Message {
	mi := &file_message_office_internal_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GenerateClusterTokenOut.ProtoReflect.Descriptor instead.
func (*GenerateClusterTokenOut) Descriptor() ([]byte, []int) {
	return file_message_office_internal_proto_rawDescGZIP(), []int{1}
}

func (x *GenerateClusterTokenOut) GetClusterToken() string {
	if x != nil {
		return x.ClusterToken
	}
	return ""
}

type GetClusterTokenIn struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	AccountName   string                 `protobuf:"bytes,1,opt,name=accountName,proto3" json:"accountName,omitempty"`
	ClusterName   string                 `protobuf:"bytes,2,opt,name=clusterName,proto3" json:"clusterName,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetClusterTokenIn) Reset() {
	*x = GetClusterTokenIn{}
	mi := &file_message_office_internal_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetClusterTokenIn) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetClusterTokenIn) ProtoMessage() {}

func (x *GetClusterTokenIn) ProtoReflect() protoreflect.Message {
	mi := &file_message_office_internal_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetClusterTokenIn.ProtoReflect.Descriptor instead.
func (*GetClusterTokenIn) Descriptor() ([]byte, []int) {
	return file_message_office_internal_proto_rawDescGZIP(), []int{2}
}

func (x *GetClusterTokenIn) GetAccountName() string {
	if x != nil {
		return x.AccountName
	}
	return ""
}

func (x *GetClusterTokenIn) GetClusterName() string {
	if x != nil {
		return x.ClusterName
	}
	return ""
}

type GetClusterTokenOut struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ClusterToken  string                 `protobuf:"bytes,1,opt,name=clusterToken,proto3" json:"clusterToken,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetClusterTokenOut) Reset() {
	*x = GetClusterTokenOut{}
	mi := &file_message_office_internal_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetClusterTokenOut) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetClusterTokenOut) ProtoMessage() {}

func (x *GetClusterTokenOut) ProtoReflect() protoreflect.Message {
	mi := &file_message_office_internal_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetClusterTokenOut.ProtoReflect.Descriptor instead.
func (*GetClusterTokenOut) Descriptor() ([]byte, []int) {
	return file_message_office_internal_proto_rawDescGZIP(), []int{3}
}

func (x *GetClusterTokenOut) GetClusterToken() string {
	if x != nil {
		return x.ClusterToken
	}
	return ""
}

var File_message_office_internal_proto protoreflect.FileDescriptor

const file_message_office_internal_proto_rawDesc = "" +
	"\n" +
	"\x1dmessage-office-internal.proto\"\\\n" +
	"\x16GenerateClusterTokenIn\x12 \n" +
	"\vaccountName\x18\x01 \x01(\tR\vaccountName\x12 \n" +
	"\vclusterName\x18\x02 \x01(\tR\vclusterName\"=\n" +
	"\x17GenerateClusterTokenOut\x12\"\n" +
	"\fclusterToken\x18\x01 \x01(\tR\fclusterToken\"W\n" +
	"\x11GetClusterTokenIn\x12 \n" +
	"\vaccountName\x18\x01 \x01(\tR\vaccountName\x12 \n" +
	"\vclusterName\x18\x02 \x01(\tR\vclusterName\"8\n" +
	"\x12GetClusterTokenOut\x12\"\n" +
	"\fclusterToken\x18\x01 \x01(\tR\fclusterToken2\x9e\x01\n" +
	"\x15MessageOfficeInternal\x12I\n" +
	"\x14GenerateClusterToken\x12\x17.GenerateClusterTokenIn\x1a\x18.GenerateClusterTokenOut\x12:\n" +
	"\x0fGetClusterToken\x12\x12.GetClusterTokenIn\x1a\x13.GetClusterTokenOutB*Z(kloudlite.io/rpc/message-office-internalb\x06proto3"

var (
	file_message_office_internal_proto_rawDescOnce sync.Once
	file_message_office_internal_proto_rawDescData []byte
)

func file_message_office_internal_proto_rawDescGZIP() []byte {
	file_message_office_internal_proto_rawDescOnce.Do(func() {
		file_message_office_internal_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_message_office_internal_proto_rawDesc), len(file_message_office_internal_proto_rawDesc)))
	})
	return file_message_office_internal_proto_rawDescData
}

var file_message_office_internal_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_message_office_internal_proto_goTypes = []any{
	(*GenerateClusterTokenIn)(nil),  // 0: GenerateClusterTokenIn
	(*GenerateClusterTokenOut)(nil), // 1: GenerateClusterTokenOut
	(*GetClusterTokenIn)(nil),       // 2: GetClusterTokenIn
	(*GetClusterTokenOut)(nil),      // 3: GetClusterTokenOut
}
var file_message_office_internal_proto_depIdxs = []int32{
	0, // 0: MessageOfficeInternal.GenerateClusterToken:input_type -> GenerateClusterTokenIn
	2, // 1: MessageOfficeInternal.GetClusterToken:input_type -> GetClusterTokenIn
	1, // 2: MessageOfficeInternal.GenerateClusterToken:output_type -> GenerateClusterTokenOut
	3, // 3: MessageOfficeInternal.GetClusterToken:output_type -> GetClusterTokenOut
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_message_office_internal_proto_init() }
func file_message_office_internal_proto_init() {
	if File_message_office_internal_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_message_office_internal_proto_rawDesc), len(file_message_office_internal_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_message_office_internal_proto_goTypes,
		DependencyIndexes: file_message_office_internal_proto_depIdxs,
		MessageInfos:      file_message_office_internal_proto_msgTypes,
	}.Build()
	File_message_office_internal_proto = out.File
	file_message_office_internal_proto_goTypes = nil
	file_message_office_internal_proto_depIdxs = nil
}
