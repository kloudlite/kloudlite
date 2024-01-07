package kubectl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kloudlite/operator/pkg/logging"
	"io"
	"time"

	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	"sigs.k8s.io/controller-runtime/pkg/client"
	myaml "sigs.k8s.io/yaml"

	fn "github.com/kloudlite/operator/pkg/functions"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type YAMLClient interface {
	Apply(ctx context.Context, obj client.Object) ([]rApi.ResourceRef, error)
	ApplyYAML(ctx context.Context, yamls ...[]byte) ([]rApi.ResourceRef, error)
	DeleteResource(ctx context.Context, obj client.Object) error
	DeleteYAML(ctx context.Context, yamls ...[]byte) error
	RolloutRestart(ctx context.Context, kind Restartable, namespace string, labels map[string]string) error

	Client() *kubernetes.Clientset
}

type yamlClient struct {
	k8sClient     *kubernetes.Clientset
	dynamicClient dynamic.Interface
	mapper        meta.RESTMapper
	logger        logging.Logger
}

func (yc *yamlClient) Client() *kubernetes.Clientset {
	return yc.k8sClient
}

func (yc *yamlClient) Apply(ctx context.Context, obj client.Object) ([]rApi.ResourceRef, error) {
	b, err := myaml.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return yc.ApplyYAML(ctx, b)
}

func (yc *yamlClient) ApplyYAML(ctx context.Context, yamls ...[]byte) ([]rApi.ResourceRef, error) {
	b := bytes.Join(yamls, []byte("\n---\n"))
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(b), 100)

	var resources []rApi.ResourceRef

	for {
		var obj unstructured.Unstructured
		if err := decoder.Decode(&obj); err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}

		if obj.Object == nil {
			continue
		}

		resources = append(resources, rApi.ParseResourceRef(&obj))

		gvk := obj.GroupVersionKind()
		mapping, err := yc.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return nil, err
		}

		resourceClient := func() dynamic.ResourceInterface {
			if obj.GetNamespace() == "" {
				return yc.dynamicClient.Resource(mapping.Resource).Namespace("default")
			}
			return yc.dynamicClient.Resource(mapping.Resource)
		}()

		ann := obj.GetAnnotations()
		if ann == nil {
			ann = make(map[string]string)
		}

		labels := obj.GetLabels()
		if labels == nil {
			labels = make(map[string]string)
		}

		b, err := json.Marshal(obj.Object)
		if err != nil {
			return resources, err
		}

		ann[constants.LastAppliedKey] = string(b)

		// Check if the resource exists
		cobj, err := resourceClient.Get(ctx, obj.GetName(), metav1.GetOptions{})
		if err != nil && apiErrors.IsNotFound(err) {
			// If not exists, create it
			obj.SetAnnotations(ann)
			obj.SetLabels(labels)
			_, err = resourceClient.Create(ctx, &obj, metav1.CreateOptions{})
			if err != nil {
				return resources, err
			}
			yc.logger.Infof("created resource (gvk: %s) (%s/%s)", gvk.String(), obj.GetNamespace(), obj.GetName())
			continue
		}

		prevLastApplied, ok := cobj.GetAnnotations()[constants.LastAppliedKey]
		if ok {
			if prevLastApplied == ann[constants.LastAppliedKey] {
				yc.logger.Infof("No changes for resource (gvk: %s) (%s/%s)", gvk.String(), obj.GetNamespace(), obj.GetName())
				continue
			}

			var prevAppliedObj unstructured.Unstructured
			if err := json.Unmarshal([]byte(prevLastApplied), &prevAppliedObj); err != nil {
				return nil, err
			}

			prevAnn := prevAppliedObj.GetAnnotations()

			for k, v := range cobj.GetAnnotations() {
				if !fn.MapHasKey(ann, k) && !fn.MapHasKey(prevAnn, k) {
					ann[k] = v
				}
			}

			prevLabels := prevAppliedObj.GetLabels()

			for k, v := range cobj.GetLabels() {
				if !fn.MapHasKey(labels, k) && !fn.MapHasKey(prevLabels, k) {
					labels[k] = v
				}
			}
		}

		obj.SetAnnotations(ann)
		obj.SetLabels(labels)
		// If exists, update it
		if _, err = resourceClient.Update(ctx, &obj, metav1.UpdateOptions{}); err != nil {
			return resources, err
		}
		yc.logger.Infof("updated resource (gvk: %s) (%s/%s)", gvk.String(), obj.GetNamespace(), obj.GetName())
	}
	return resources, nil
}

func (yc *yamlClient) ApplyYAMLOld(ctx context.Context, yamls ...[]byte) ([]rApi.ResourceRef, error) {
	b := bytes.Join(yamls, []byte("\n---\n"))
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(b), 100)

	var resources []rApi.ResourceRef

	for {
		var rawObj runtime.RawExtension
		if err := decoder.Decode(&rawObj); err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}

		if rawObj.Raw == nil {
			continue
		}

		obj, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			return nil, err
		}

		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return nil, err
		}

		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

		// TODO: (this is failing cross importing)
		resources = append(resources, rApi.ParseResourceRef(unstructuredObj))
		ann := unstructuredObj.GetAnnotations()
		if ann == nil {
			ann = make(map[string]string, 2)
		}

		ann[constants.GVKKey] = gvk.String()
		metadata, ok := unstructuredMap["metadata"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid object format")
		}
		if metadata["annotations"] != nil {
			delete(metadata["annotations"].(map[string]any), constants.LastAppliedKey)
		}

		delete(unstructuredMap, "status")

		b, err := json.Marshal(unstructuredMap)
		if err != nil {
			return nil, err
		}

		ann[constants.LastAppliedKey] = string(b)

		unstructuredObj.SetAnnotations(ann)

		var dri dynamic.ResourceInterface

		mapping, err := yc.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			// log.Fatal(err)
			return nil, err
		}
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			if unstructuredObj.GetNamespace() == "" {
				unstructuredObj.SetNamespace("default")
			}
			dri = yc.dynamicClient.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
		} else {
			dri = yc.dynamicClient.Resource(mapping.Resource)
		}

		resource, err := dri.Get(ctx, unstructuredObj.GetName(), metav1.GetOptions{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return nil, err
			}
		}

		// TODO (nxtcoder17): delete, and recreate deployment if service account has been changed
		if resource != nil && resource.GetAnnotations()[constants.LastAppliedKey] == string(b) {
			continue
		}

		resourceRaw, err := json.Marshal(unstructuredObj.Object)
		if err != nil {
			continue
		}

		if _, err := dri.Patch(
			context.Background(),
			unstructuredObj.GetName(),
			types.MergePatchType,
			resourceRaw,
			metav1.PatchOptions{},
		); err != nil {
			if apiErrors.IsNotFound(err) {
				if _, err := dri.Create(ctx, unstructuredObj, metav1.CreateOptions{}); err != nil {
					// log.Fatal(err)
					return nil, err
				}
				continue
			}
			// log.Fatal(err)
			return nil, err
		}
	}
	return resources, nil
}

func (yc *yamlClient) DeleteResource(ctx context.Context, obj client.Object) error {
	gvk := obj.GetObjectKind().GroupVersionKind()
	gvr := gvk.GroupVersion().WithResource(fn.RegularPlural(gvk.Kind))
	return yc.dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Delete(ctx, obj.GetName(), metav1.DeleteOptions{})
}

func (yc *yamlClient) DeleteYAML(ctx context.Context, yamls ...[]byte) error {
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

		mapping, err := yc.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
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

		if err := dri.Delete(ctx, unstructuredObj.GetName(), metav1.DeleteOptions{}); err != nil {
			if apiErrors.IsNotFound(err) {
				return nil
			}
			return err
		}
	}

	return nil
}

type Restartable string

const (
	Deployment  Restartable = "deployment"
	StatefulSet Restartable = "statefulset"
)

func (yc *yamlClient) RolloutRestart(ctx context.Context, kind Restartable, namespace string, labels map[string]string) error {
	switch kind {
	case Deployment:
		{
			dl, err := yc.k8sClient.AppsV1().Deployments(namespace).List(
				ctx, metav1.ListOptions{
					LabelSelector: apiLabels.FormatLabels(labels),
				},
			)
			if err != nil {
				return err
			}
			for _, d := range dl.Items {
				if d.Annotations == nil {
					d.Annotations = map[string]string{}
				}
				d.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
				if _, err := yc.k8sClient.AppsV1().Deployments(namespace).Update(ctx, &d, metav1.UpdateOptions{}); err != nil {
					return err
				}
			}
		}
	case StatefulSet:
		{
			sl, err := yc.k8sClient.AppsV1().StatefulSets(namespace).List(
				ctx, metav1.ListOptions{
					LabelSelector: apiLabels.FormatLabels(labels),
				},
			)
			if err != nil {
				return err
			}
			for _, d := range sl.Items {
				if d.Annotations == nil {
					d.Annotations = map[string]string{}
				}
				d.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
				if _, err := yc.k8sClient.AppsV1().StatefulSets(namespace).Update(ctx, &d, metav1.UpdateOptions{}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type YAMLClientOpts struct {
	Logger logging.Logger
}

func NewYAMLClient(config *rest.Config, opts YAMLClientOpts) (YAMLClient, error) {
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
		return nil, err
	}

	mapper := restmapper.NewDiscoveryRESTMapper(gr)

	if opts.Logger == nil {
		opts.Logger, err = logging.New(&logging.Options{
			Name:        "k8s-yaml-client",
			CallerTrace: true,
		})
		if err != nil {
			return nil, err
		}
	}

	return &yamlClient{
		k8sClient:     c,
		dynamicClient: dc,
		mapper:        mapper,
		logger:        opts.Logger,
	}, nil
}

func NewYAMLClientOrDie(config *rest.Config, opts YAMLClientOpts) YAMLClient {
	cli, err := NewYAMLClient(config, opts)
	if err != nil {
		panic(err)
	}
	return cli
}
