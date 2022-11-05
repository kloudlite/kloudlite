package kubectl

import (
	"bytes"
	"context"
	"io"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"

	"k8s.io/apimachinery/pkg/types"
)

type YAMLClient struct {
	k8sClient     *kubernetes.Clientset
	dynamicClient dynamic.Interface
	restMapper    meta.RESTMapper
}

func (yc *YAMLClient) ApplyYAML(ctx context.Context, yamls ...[]byte) error {
	jYamls := bytes.Join(yamls, []byte("\n---\n"))
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(jYamls), 100)
	for {
		var rawObj runtime.RawExtension
		if err := decoder.Decode(&rawObj); err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		if rawObj.Raw == nil {
			continue
		}

		obj, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			return err
		}
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			// log.Fatal(err)
			return err
		}

		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

		var dri dynamic.ResourceInterface

		mapping, err := yc.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			// log.Fatal(err)
			return err
		}
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			if unstructuredObj.GetNamespace() == "" {
				unstructuredObj.SetNamespace("default")
			}
			dri = yc.dynamicClient.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
		} else {
			dri = yc.dynamicClient.Resource(mapping.Resource)
		}

		if _, err := dri.Patch(
			context.Background(),
			unstructuredObj.GetName(),
			types.MergePatchType,
			rawObj.Raw,
			metav1.PatchOptions{},
		); err != nil {
			if errors.IsNotFound(err) {
				if _, err := dri.Create(ctx, unstructuredObj, metav1.CreateOptions{}); err != nil {
					// log.Fatal(err)
					return err
				}
				return nil
			}
			// log.Fatal(err)
			return err
		}
	}
	return nil
}

func NewYAMLClient(config *rest.Config) (*YAMLClient, error) {
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	dc, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	gr, err := restmapper.GetAPIGroupResources(c.Discovery())
	if err != nil {
		// log.Fatal(err)
		return nil, err
	}

	mapper := restmapper.NewDiscoveryRESTMapper(gr)

	return &YAMLClient{
		k8sClient:     c,
		dynamicClient: dc,
		restMapper:    mapper,
	}, nil
}

func NewYAMLClientOrDie(config *rest.Config) *YAMLClient {
	client, err := NewYAMLClient(config)
	if err != nil {
		panic(err)
	}
	return client
}
