package functions

import (
	"bytes"
	"context"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"os/exec"

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
	if err := cli.Update(ctx, obj); err != nil {
		if apiErrors.IsNotFound(err) {
			return cli.Create(ctx, obj)
		}
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
