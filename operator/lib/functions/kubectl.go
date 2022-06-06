package functions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"os/exec"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"operators.kloudlite.io/lib/errors"
)

func KubectlApplyExec(stdin ...[]byte) (stdout *bytes.Buffer, err error) {
	c := exec.Command("kubectl", "apply", "-f", "-")
	outStream, errStream := bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
	c.Stdin = bytes.NewBuffer(bytes.Join(stdin, []byte("\n---\n")))
	c.Stdout = outStream
	c.Stderr = errStream
	if err := c.Run(); err != nil {
		return outStream, errors.NewEf(err, errStream.String())
	}
	fmt.Printf("stdout: %s\n", outStream.Bytes())
	return outStream, nil
}

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

	var j map[string]interface{}
	if err := json.Unmarshal(x, &j); err != nil {
		return err
	}

	if _, ok := j["spec"]; ok {
		t.Object["spec"] = j["spec"]
	}

	if _, ok := j["data"]; ok {
		t.Object["data"] = j["data"]
	}

	if _, ok := j["stringData"]; ok {
		t.Object["stringData"] = j["stringData"]
	}
	return cli.Update(ctx, t)
}

func KubectlGet(namespace string, resourceRef string) ([]byte, error) {
	c := exec.Command("kubectl", "get", "-o", "json", "-n", namespace, resourceRef)
	errB := bytes.NewBuffer([]byte{})
	outB := bytes.NewBuffer([]byte{})
	c.Stderr = errB
	c.Stdout = outB
	if err := c.Run(); err != nil {
		return nil, errors.NewEf(err, errB.String())
	}
	return outB.Bytes(), nil
}

func KubectlDelete(namespace, resourceRef string) error {
	c := exec.Command("kubectl", "delete", "-n", namespace, resourceRef)
	errB := bytes.NewBuffer([]byte{})
	outB := bytes.NewBuffer([]byte{})
	c.Stderr = errB
	c.Stdout = outB
	if err := c.Run(); err != nil {
		return errors.NewEf(err, errB.String())
	}
	return nil
}

func AsOwner(r client.Object, controller bool) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion:         r.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:               r.GetObjectKind().GroupVersionKind().Kind,
		Name:               r.GetName(),
		UID:                r.GetUID(),
		Controller:         NewBool(controller),
		BlockOwnerDeletion: NewBool(true),
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

func NamespacedName(obj client.Object) types.NamespacedName {
	return types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}
}

func ToYaml(obj client.Object) ([]byte, error) {
	b, err := json.Marshal(obj)
	return b, err
}
