package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/xeipuuv/gojsonschema"
	apiExtensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	// client go like
	Get(ctx context.Context, nn types.NamespacedName, obj client.Object) error
	Create(ctx context.Context, obj client.Object) error

	// custom ones
	ValidateObject(ctx context.Context, obj client.Object) error

	ApplyYAML(ctx context.Context, yamls ...[]byte) error
	DeleteYAML(ctx context.Context, yamls ...[]byte) error
}

type clientHandler struct {
	kclient    client.Client
	kclientset *clientset.Clientset
	yamlclient kubectl.YAMLClient
}

// CreateOrUpdate implements Client.
func (ch *clientHandler) Create(ctx context.Context, obj client.Object) error {
	return ch.kclient.Create(ctx, obj)
}

// Get implements Client.
func (c *clientHandler) Get(ctx context.Context, nn types.NamespacedName, obj client.Object) error {
	return c.kclient.Get(ctx, nn, obj)
}

// ValidateObject implements Client.
func (c *clientHandler) ValidateObject(ctx context.Context, obj client.Object) error {
	gvk := obj.GetObjectKind().GroupVersionKind()

	input, err := json.Marshal(obj)
	if err != nil {
		return errors.NewEf(err, "failed to marshal input struct")
	}
	documentLoader := gojsonschema.NewBytesLoader(input)

	crd, err := c.kclientset.ApiextensionsV1().CustomResourceDefinitions().Get(ctx,
		fmt.Sprintf("%s.%s", fn.RegularPlural(gvk.Kind), gvk.Group), metav1.GetOptions{})
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
			Pattern:   `^[a-z0-9]([-a-z0-9]*[a-z0-9])?([.][a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`, // source, kubectl apply with an incorrect name
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

// ApplyYAML implements Client.
func (c *clientHandler) ApplyYAML(ctx context.Context, yamls ...[]byte) error {
	if _, err := c.yamlclient.ApplyYAML(ctx, yamls...); err != nil {
		return err
	}
	return nil
}

// DeleteYAML implements Client.
func (c *clientHandler) DeleteYAML(ctx context.Context, yamls ...[]byte) error {
	return c.yamlclient.DeleteYAML(ctx, yamls...)
}

func NewClient(cfg *rest.Config, scheme *runtime.Scheme) (Client, error) {
	if scheme == nil {
		scheme = runtime.NewScheme()
		clientgoscheme.AddToScheme(scheme)
	}

	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: nil,
	})
	if err != nil {
		return nil, err
	}

	clientset, err := clientset.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	yamlclient, err := kubectl.NewYAMLClient(cfg)
	if err != nil {
		return nil, err
	}

	return &clientHandler{
		kclient:    c,
		kclientset: clientset,
		yamlclient: yamlclient,
	}, nil
}
