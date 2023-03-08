package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/xeipuuv/gojsonschema"
	apiExtensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"kloudlite.io/pkg/errors"
	fn "kloudlite.io/pkg/functions"
)

type InvalidSchemaError struct {
	err     error
	errMsgs []string
}

func (ise InvalidSchemaError) Error() string {
	m := map[string]any{
		"message":          ise.err.Error(),
		"type":             "InvalidData",
		"validationErrors": ise.errMsgs,
	}
	b, err := json.Marshal(m)
	if err != nil {
		fmt.Println("[UNEXPECTED] failed to marshal InvalidSchemaError", err)
		return ise.err.Error()
	}
	return string(b)
}

func NewInvalidSchemaError(err error, errMsgs []string) InvalidSchemaError {
	return InvalidSchemaError{err: err, errMsgs: errMsgs}
}

type ExtendedK8sClient interface {
	GetCRDJsonSchema(ctx context.Context, name string) (*apiExtensionsV1.JSONSchemaProps, error)
	ValidateStruct(ctx context.Context, s any, crdName string) error
}

type extendedK8sClient struct {
	client *clientset.Clientset
}

func (e extendedK8sClient) GetCRDJsonSchema(ctx context.Context, name string) (*apiExtensionsV1.JSONSchemaProps, error) {
	crd, err := e.client.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return crd.Spec.Versions[0].Schema.OpenAPIV3Schema, nil
}

func (e extendedK8sClient) ValidateStruct(ctx context.Context, s any, crdName string) error {
	input, err := json.Marshal(s)

	if err != nil {
		return errors.NewEf(err, "failed to marshal input struct")
	}
	documentLoader := gojsonschema.NewBytesLoader(input)

	crd, err := e.client.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	var props apiExtensionsV1.JSONSchemaProps = crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["metadata"]
	props.Required = []string{"name"}
	if crd.Spec.Scope == apiExtensionsV1.NamespaceScoped {
		props.Required = append(props.Required, "namespace")
	}

	props.Properties = map[string]apiExtensionsV1.JSONSchemaProps{
		"name": {
			Type:      gojsonschema.TYPE_STRING,
			MinLength: fn.New(int64(1)),
			MaxLength: fn.New(int64(63)),
		},
		"namespace": {
			Type:      gojsonschema.TYPE_STRING,
			MinLength: fn.New(int64(1)),
			MaxLength: fn.New(int64(63)),
		},
	}
	crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["metadata"] = props

	b, err := json.Marshal(crd.Spec.Versions[0].Schema.OpenAPIV3Schema)
	if err != nil {
		return err
	}

	schemaLoader := gojsonschema.NewBytesLoader(b)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		errMsgs := make([]string, 0, len(result.Errors()))
		for _, err := range result.Errors() {
			errMsgs = append(errMsgs, err.String())
		}
		return NewInvalidSchemaError(fmt.Errorf("document is invalid"), errMsgs)
	}
	return nil
}

func NewExtendedK8sClient(config *rest.Config) (ExtendedK8sClient, error) {
	client, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &extendedK8sClient{
		client: client,
	}, nil
}
