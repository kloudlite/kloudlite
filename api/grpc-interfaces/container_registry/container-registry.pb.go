// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v3.21.12
// source: container-registry.proto

package container_registry

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

type CreateReadOnlyCredentialIn struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	AccountName      string                 `protobuf:"bytes,1,opt,name=accountName,proto3" json:"accountName,omitempty"`
	UserId           string                 `protobuf:"bytes,2,opt,name=userId,proto3" json:"userId,omitempty"`
	CredentialName   string                 `protobuf:"bytes,3,opt,name=credentialName,proto3" json:"credentialName,omitempty"`
	RegistryUsername string                 `protobuf:"bytes,4,opt,name=registryUsername,proto3" json:"registryUsername,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *CreateReadOnlyCredentialIn) Reset() {
	*x = CreateReadOnlyCredentialIn{}
	mi := &file_container_registry_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateReadOnlyCredentialIn) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateReadOnlyCredentialIn) ProtoMessage() {}

func (x *CreateReadOnlyCredentialIn) ProtoReflect() protoreflect.Message {
	mi := &file_container_registry_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateReadOnlyCredentialIn.ProtoReflect.Descriptor instead.
func (*CreateReadOnlyCredentialIn) Descriptor() ([]byte, []int) {
	return file_container_registry_proto_rawDescGZIP(), []int{0}
}

func (x *CreateReadOnlyCredentialIn) GetAccountName() string {
	if x != nil {
		return x.AccountName
	}
	return ""
}

func (x *CreateReadOnlyCredentialIn) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *CreateReadOnlyCredentialIn) GetCredentialName() string {
	if x != nil {
		return x.CredentialName
	}
	return ""
}

func (x *CreateReadOnlyCredentialIn) GetRegistryUsername() string {
	if x != nil {
		return x.RegistryUsername
	}
	return ""
}

type CreateReadOnlyCredentialOut struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// dcokerconfigjson is as per format: https://kubernetes.io/docs/concepts/configuration/secret/#docker-config-secrets
	DockerConfigJson []byte `protobuf:"bytes,1,opt,name=dockerConfigJson,proto3" json:"dockerConfigJson,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *CreateReadOnlyCredentialOut) Reset() {
	*x = CreateReadOnlyCredentialOut{}
	mi := &file_container_registry_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateReadOnlyCredentialOut) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateReadOnlyCredentialOut) ProtoMessage() {}

func (x *CreateReadOnlyCredentialOut) ProtoReflect() protoreflect.Message {
	mi := &file_container_registry_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateReadOnlyCredentialOut.ProtoReflect.Descriptor instead.
func (*CreateReadOnlyCredentialOut) Descriptor() ([]byte, []int) {
	return file_container_registry_proto_rawDescGZIP(), []int{1}
}

func (x *CreateReadOnlyCredentialOut) GetDockerConfigJson() []byte {
	if x != nil {
		return x.DockerConfigJson
	}
	return nil
}

var File_container_registry_proto protoreflect.FileDescriptor

const file_container_registry_proto_rawDesc = "" +
	"\n" +
	"\x18container-registry.proto\"\xaa\x01\n" +
	"\x1aCreateReadOnlyCredentialIn\x12 \n" +
	"\vaccountName\x18\x01 \x01(\tR\vaccountName\x12\x16\n" +
	"\x06userId\x18\x02 \x01(\tR\x06userId\x12&\n" +
	"\x0ecredentialName\x18\x03 \x01(\tR\x0ecredentialName\x12*\n" +
	"\x10registryUsername\x18\x04 \x01(\tR\x10registryUsername\"I\n" +
	"\x1bCreateReadOnlyCredentialOut\x12*\n" +
	"\x10dockerConfigJson\x18\x01 \x01(\fR\x10dockerConfigJson2j\n" +
	"\x11ContainerRegistry\x12U\n" +
	"\x18CreateReadOnlyCredential\x12\x1b.CreateReadOnlyCredentialIn\x1a\x1c.CreateReadOnlyCredentialOutB\x16Z\x14./container_registryb\x06proto3"

var (
	file_container_registry_proto_rawDescOnce sync.Once
	file_container_registry_proto_rawDescData []byte
)

func file_container_registry_proto_rawDescGZIP() []byte {
	file_container_registry_proto_rawDescOnce.Do(func() {
		file_container_registry_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_container_registry_proto_rawDesc), len(file_container_registry_proto_rawDesc)))
	})
	return file_container_registry_proto_rawDescData
}

var file_container_registry_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_container_registry_proto_goTypes = []any{
	(*CreateReadOnlyCredentialIn)(nil),  // 0: CreateReadOnlyCredentialIn
	(*CreateReadOnlyCredentialOut)(nil), // 1: CreateReadOnlyCredentialOut
}
var file_container_registry_proto_depIdxs = []int32{
	0, // 0: ContainerRegistry.CreateReadOnlyCredential:input_type -> CreateReadOnlyCredentialIn
	1, // 1: ContainerRegistry.CreateReadOnlyCredential:output_type -> CreateReadOnlyCredentialOut
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_container_registry_proto_init() }
func file_container_registry_proto_init() {
	if File_container_registry_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_container_registry_proto_rawDesc), len(file_container_registry_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_container_registry_proto_goTypes,
		DependencyIndexes: file_container_registry_proto_depIdxs,
		MessageInfos:      file_container_registry_proto_msgTypes,
	}.Build()
	File_container_registry_proto = out.File
	file_container_registry_proto_goTypes = nil
	file_container_registry_proto_depIdxs = nil
}
