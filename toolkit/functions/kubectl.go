package functions

import (
	"context"
	"encoding/json"
	"strings"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kloudlite/operator/toolkit/errors"

	"github.com/gobuffalo/flect"
)

func toUnstructured(obj client.Object) (*unstructured.Unstructured, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	t := &unstructured.Unstructured{Object: m}
	return t, nil
}

func KubectlApply(ctx context.Context, cli client.Client, obj client.Object) error {
	t, err := toUnstructured(obj)
	if err != nil {
		return err
	}
	if err := cli.Get(
		ctx, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, t,
	); err != nil {
		if !apiErrors.IsNotFound(err) {
			return errors.NewEf(err, "could not get k8s resource")
		}
		// CREATE it
		return cli.Create(ctx, obj)
	}

	// UPDATE it
	x, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	var j map[string]any
	if err := json.Unmarshal(x, &j); err != nil {
		return err
	}

	if m, ok := j["metadata"].(map[string]any); ok {
		annotations := map[string]string{}

		for k, v := range t.GetAnnotations() {
			annotations[k] = v
		}

		if m2, ok := m["annotations"].(map[string]string); ok {
			for k, v := range m2 {
				annotations[k] = v
			}
		}
		t.SetAnnotations(annotations)

		labels := map[string]string{}

		for k, v := range t.GetLabels() {
			labels[k] = v
		}

		if m2, ok := m["labels"].(map[string]string); ok {
			for k, v := range m2 {
				labels[k] = v
			}
		}

		t.SetLabels(labels)
	}

	// for general types
	if _, ok := j["spec"]; ok {
		t.Object["spec"] = j["spec"]
	}

	// For Configmap/secret
	if _, ok := j["data"]; ok {
		t.Object["data"] = j["data"]
	}

	// for secret
	if _, ok := j["stringData"]; ok {
		t.Object["stringData"] = j["stringData"]
	}
	return cli.Update(ctx, t)
}

func AsOwner(r client.Object, controller ...bool) metav1.OwnerReference {
	ctrler := false
	if len(controller) > 0 {
		ctrler = controller[0]
	}
	return metav1.OwnerReference{
		APIVersion: r.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:       r.GetObjectKind().GroupVersionKind().Kind,
		Name:       r.GetName(),
		UID:        r.GetUID(),
		Controller: &ctrler,
		// BlockOwnerDeletion: New(false),
		BlockOwnerDeletion: &ctrler,
	}
}

func IsOwner(obj client.Object, ownerRef metav1.OwnerReference) bool {
	for _, ref := range obj.GetOwnerReferences() {
		if ref.Name == ownerRef.Name &&
			ref.UID == ownerRef.UID &&
			ref.Kind == ownerRef.Kind && ref.
			APIVersion == ownerRef.APIVersion {
			return true
		}
	}
	return false
}

func HasOwner(obj client.Object) bool {
	return len(obj.GetOwnerReferences()) > 0
}

func ParseGVK(apiVersion string, kind string) schema.GroupVersionKind {
	gv, _ := schema.ParseGroupVersion(apiVersion)
	return schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    kind,
	}
}

func GVK(obj client.Object) metav1.GroupVersionKind {
	gvk := obj.GetObjectKind().GroupVersionKind()
	return metav1.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
	}
}

// RegularPlural is used to pluralize group of k8s CRDs from kind
// It is copied from https://github.com/kubernetes-sigs/kubebuilder/blob/afce6a0e8c2a6d5682be07bbe502e728dd619714/pkg/model/resource/utils.go#L71
func RegularPlural(singular string) string {
	return flect.Pluralize(strings.ToLower(singular))
}
