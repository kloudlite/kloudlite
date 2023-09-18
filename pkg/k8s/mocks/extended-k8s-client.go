package mocks

import (
	context "context"
	apiExtensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type ExtendedK8sClientCallerInfo struct {
	Args []any
}

type ExtendedK8sClient struct {
	Calls                map[string][]ExtendedK8sClientCallerInfo
	MockGetCRDJsonSchema func(ctx context.Context, name string) (*apiExtensionsV1.JSONSchemaProps, error)
	MockValidateStruct   func(ctx context.Context, obj client.Object) error
}

func (m *ExtendedK8sClient) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]ExtendedK8sClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], ExtendedK8sClientCallerInfo{Args: args})
}

func (eMock *ExtendedK8sClient) GetCRDJsonSchema(ctx context.Context, name string) (*apiExtensionsV1.JSONSchemaProps, error) {
	if eMock.MockGetCRDJsonSchema != nil {
		eMock.registerCall("GetCRDJsonSchema", ctx, name)
		return eMock.MockGetCRDJsonSchema(ctx, name)
	}
	panic("method 'GetCRDJsonSchema' not implemented, yet")
}

func (eMock *ExtendedK8sClient) ValidateStruct(ctx context.Context, obj client.Object) error {
	if eMock.MockValidateStruct != nil {
		eMock.registerCall("ValidateStruct", ctx, obj)
		return eMock.MockValidateStruct(ctx, obj)
	}
	panic("method 'ValidateStruct' not implemented, yet")
}

func NewExtendedK8sClient() *ExtendedK8sClient {
	return &ExtendedK8sClient{}
}
