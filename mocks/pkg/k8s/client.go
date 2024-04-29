package mocks

import (
	context "context"
	k8s "github.com/kloudlite/api/pkg/k8s"
	io "io"
	corev1 "k8s.io/api/core/v1"
	types "k8s.io/apimachinery/pkg/types"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type ClientCallerInfo struct {
	Args []any
}

type Client struct {
	Calls              map[string][]ClientCallerInfo
	MockApplyYAML      func(ctx context.Context, yamls ...[]byte) error
	MockCreate         func(ctx context.Context, obj client.Object) error
	MockDelete         func(ctx context.Context, obj client.Object) error
	MockDeleteYAML     func(ctx context.Context, yamls ...[]byte) error
	MockGet            func(ctx context.Context, nn types.NamespacedName, obj client.Object) error
	MockList           func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
	MockListSecrets    func(ctx context.Context, namespace string, secretType corev1.SecretType) ([]v1.Secret, error)
	MockReadLogs       func(ctx context.Context, namespace string, name string, writer io.WriteCloser, opts *k8s.ReadLogsOptions) error
	MockUpdate         func(ctx context.Context, obj client.Object) error
	MockValidateObject func(ctx context.Context, obj client.Object) error
}

func (m *Client) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]ClientCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], ClientCallerInfo{Args: args})
}

func (cMock *Client) ApplyYAML(ctx context.Context, yamls ...[]byte) error {
	if cMock.MockApplyYAML != nil {
		cMock.registerCall("ApplyYAML", ctx, yamls)
		return cMock.MockApplyYAML(ctx, yamls...)
	}
	panic("Client: method 'ApplyYAML' not implemented, yet")
}

func (cMock *Client) Create(ctx context.Context, obj client.Object) error {
	if cMock.MockCreate != nil {
		cMock.registerCall("Create", ctx, obj)
		return cMock.MockCreate(ctx, obj)
	}
	panic("Client: method 'Create' not implemented, yet")
}

func (cMock *Client) Delete(ctx context.Context, obj client.Object) error {
	if cMock.MockDelete != nil {
		cMock.registerCall("Delete", ctx, obj)
		return cMock.MockDelete(ctx, obj)
	}
	panic("Client: method 'Delete' not implemented, yet")
}

func (cMock *Client) DeleteYAML(ctx context.Context, yamls ...[]byte) error {
	if cMock.MockDeleteYAML != nil {
		cMock.registerCall("DeleteYAML", ctx, yamls)
		return cMock.MockDeleteYAML(ctx, yamls...)
	}
	panic("Client: method 'DeleteYAML' not implemented, yet")
}

func (cMock *Client) Get(ctx context.Context, nn types.NamespacedName, obj client.Object) error {
	if cMock.MockGet != nil {
		cMock.registerCall("Get", ctx, nn, obj)
		return cMock.MockGet(ctx, nn, obj)
	}
	panic("Client: method 'Get' not implemented, yet")
}

func (cMock *Client) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if cMock.MockList != nil {
		cMock.registerCall("List", ctx, list, opts)
		return cMock.MockList(ctx, list, opts...)
	}
	panic("Client: method 'List' not implemented, yet")
}

func (cMock *Client) ListSecrets(ctx context.Context, namespace string, secretType corev1.SecretType) ([]v1.Secret, error) {
	if cMock.MockListSecrets != nil {
		cMock.registerCall("ListSecrets", ctx, namespace, secretType)
		return cMock.MockListSecrets(ctx, namespace, secretType)
	}
	panic("Client: method 'ListSecrets' not implemented, yet")
}

func (cMock *Client) ReadLogs(ctx context.Context, namespace string, name string, writer io.WriteCloser, opts *k8s.ReadLogsOptions) error {
	if cMock.MockReadLogs != nil {
		cMock.registerCall("ReadLogs", ctx, namespace, name, writer, opts)
		return cMock.MockReadLogs(ctx, namespace, name, writer, opts)
	}
	panic("Client: method 'ReadLogs' not implemented, yet")
}

func (cMock *Client) Update(ctx context.Context, obj client.Object) error {
	if cMock.MockUpdate != nil {
		cMock.registerCall("Update", ctx, obj)
		return cMock.MockUpdate(ctx, obj)
	}
	panic("Client: method 'Update' not implemented, yet")
}

func (cMock *Client) ValidateObject(ctx context.Context, obj client.Object) error {
	if cMock.MockValidateObject != nil {
		cMock.registerCall("ValidateObject", ctx, obj)
		return cMock.MockValidateObject(ctx, obj)
	}
	panic("Client: method 'ValidateObject' not implemented, yet")
}

func NewClient() *Client {
	return &Client{}
}
