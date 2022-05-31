package functions

import (
	"bytes"
	"context"
	"encoding/json"

	"os/exec"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	return outStream, nil
}

func KubectlApply(ctx context.Context, cli client.Client, obj client.Object) error {
	// b, err := json.Marshal(obj.DeepCopyObject())
	// if err != nil {
	// 	return err
	// }
	// var j map[string]any
	// if err := json.Unmarshal(b, &j); err != nil {
	// 	return err
	// }
	// x := unstructured.Unstructured{Object: j}
	//
	x := obj

	// cli.Update(
	// 	ctx, obj, &client.UpdateOptions{
	// 		DryRun:       nil,
	// 		FieldManager: "",
	// 		Raw:          nil,
	// 	},
	// )

	if _, err := controllerutil.CreateOrUpdate(
		ctx, cli, x, func() error {
			b1, err := json.Marshal(x.DeepCopyObject())
			if err != nil {
				return err
			}
			var j map[string]any
			if err := json.Unmarshal(b1, &j); err != nil {
				return err
			}
			serverX := unstructured.Unstructured{Object: j}

			b2, err := json.Marshal(obj)
			if err != nil {
				return err
			}
			y := unstructured.Unstructured{Object: map[string]any{}}
			if err := json.Unmarshal(b2, &y.Object); err != nil {
				return err
			}

			y.DeepCopyInto(&serverX)
			// serverX.DeepCopyInto(&y)
			// x = &y
			// x.SetAnnotations(MapMerge(x.GetAnnotations(), y.GetAnnotations()))
			// x.SetLabels(MapMerge(x.GetLabels(), y.GetLabels()))
			// x.SetOwnerReferences(y.GetOwnerReferences())
			// x.Object["spec"] = y.Object["spec"]
			// x.Object["status"] = y.Object["status"]
			return nil
		},
	); err != nil {
		return err
	}
	return nil
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

func NamespacedName(obj client.Object) types.NamespacedName {
	return types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}
}
