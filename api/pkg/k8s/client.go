package k8s

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/xeipuuv/gojsonschema"
	corev1 "k8s.io/api/core/v1"
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
	List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
	Create(ctx context.Context, obj client.Object) error
	Update(ctx context.Context, obj client.Object) error
	Delete(ctx context.Context, obj client.Object) error

	ListSecrets(ctx context.Context, namespace string, secretType corev1.SecretType) ([]corev1.Secret, error)

	// custom ones
	ValidateObject(ctx context.Context, obj client.Object) error

	ApplyYAML(ctx context.Context, yamls ...[]byte) error
	DeleteYAML(ctx context.Context, yamls ...[]byte) error

	ReadLogs(ctx context.Context, namespace, name string, writer io.WriteCloser, opts *ReadLogsOptions) error
}

type ReadLogsOptions struct {
	ContainerName string
	SinceSeconds  *int64
	TailLines     *int64
}

type clientHandler struct {
	kclient    client.Client
	kclientset *clientset.Clientset
	yamlclient kubectl.YAMLClient
}

// List implements Client.
func (ch *clientHandler) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return ch.kclient.List(ctx, list, opts...)
}

type LogLine struct {
	Message       string `json:"message"`
	Timestamp     string `json:"timestamp"`
	PodName       string `json:"podName"`
	ContainerName string `json:"containerName"`
}

// ReadLogs implements Client.
func (ch *clientHandler) ReadLogs(ctx context.Context, namespace, name string, writer io.WriteCloser, opts *ReadLogsOptions) error {
	defer writer.Close()
	if err := ctx.Err(); err != nil {
		return err
	}

	req := ch.yamlclient.Client().CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{
		Container:    opts.ContainerName,
		Follow:       true,
		Previous:     false,
		SinceSeconds: opts.SinceSeconds,
		// SinceTime:    nil,
		Timestamps: true,
		TailLines:  opts.TailLines,
	})

	rc, err := req.Stream(ctx)
	if err != nil {
		fmt.Println("err:", err)
		return err
	}
	defer rc.Close()

	r := bufio.NewReader(rc)

	for {
		b, err := r.ReadBytes('\n')
		if err != nil {
			return errors.NewE(err)
		}

		s := bytes.SplitN(b[:len(b)-1], []byte(" "), 2)
		if len(s) != 2 {
			return fmt.Errorf("invalid log line")
		}

		line := fmt.Sprintf(`{"timestamp": %q, "podName": %q, "containerName": %q, "message": %q}`, s[0], name, opts.ContainerName, s[1])
		// fmt.Printf("[DEBUG] line: %s\n", line)
		if _, err := writer.Write([]byte(line)); err != nil {
			return err
		}
		if _, err := writer.Write([]byte("\n")); err != nil {
			return err
		}
	}
}

// ListSecrets implements Client.
func (ch *clientHandler) ListSecrets(ctx context.Context, namespace string, secretType corev1.SecretType) ([]corev1.Secret, error) {
	out, err := ch.yamlclient.Client().CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("type=%s", secretType),
	})
	if err != nil {
		return nil, err
	}

	return out.Items, nil
}

// Delete implements Client.
func (ch *clientHandler) Delete(ctx context.Context, obj client.Object) error {
	return ch.kclient.Delete(ctx, obj)
}

// Update implements Client.
func (ch *clientHandler) Update(ctx context.Context, obj client.Object) error {
	return ch.kclient.Update(ctx, obj)
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
		return errors.NewE(err)
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
		return errors.NewE(err)
	}

	schemaLoader := gojsonschema.NewBytesLoader(b)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return errors.NewE(err)
	}

	if !result.Valid() {
		errMsgs := make([]string, 0, len(result.Errors()))
		for _, err := range result.Errors() {
			errMsgs = append(errMsgs, err.String())
		}
		return NewInvalidSchemaError(errors.Newf("document is invalid"), errMsgs)
	}
	return nil
}

// ApplyYAML implements Client.
func (c *clientHandler) ApplyYAML(ctx context.Context, yamls ...[]byte) error {
	if _, err := c.yamlclient.ApplyYAML(ctx, yamls...); err != nil {
		return errors.NewE(err)
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
	}

	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		fmt.Println(err)
	}

	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: nil,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	cs, err := clientset.NewForConfig(cfg)
	if err != nil {
		return nil, errors.NewE(err)
	}

	yamlclient, err := kubectl.NewYAMLClient(cfg, kubectl.YAMLClientOpts{})
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &clientHandler{
		kclient:    c,
		kclientset: cs,
		yamlclient: yamlclient,
	}, nil
}
