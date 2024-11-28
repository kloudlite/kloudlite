package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/logging"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func ContainsFinalizers(obj client.Object, finalizers ...string) bool {
	flist := obj.GetFinalizers()
	m := make(map[string]bool, len(flist))
	for i := range flist {
		m[flist[i]] = true
	}

	for i := range finalizers {
		_, ok := m[finalizers[i]]
		if !ok {
			return false
		}
	}
	return true
}

func ParseFromMap[T any, K any](m map[string]K) (*T, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var output T
	if err := json.Unmarshal(b, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func IntoMap(value any, targetMap any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &targetMap)
}

func ParseFromSecret[T any](secret *corev1.Secret) (*T, error) {
	if secret == nil {
		return nil, nil
	}
	x := make(map[string]string, len(secret.Data))
	for k, v := range secret.Data {
		x[k] = string(v)
	}

	b, err := json.Marshal(x)
	if err != nil {
		return nil, err
	}

	var output T
	if err := json.Unmarshal(b, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func ParseFromSecretData[T any](m map[string][]byte) (*T, error) {
	return ParseFromSecret[T](&corev1.Secret{Data: m})
}

type Restartable string

const (
	Deployment  Restartable = "deployment"
	StatefulSet Restartable = "statefulset"
)

func RolloutRestart(c client.Client, kind Restartable, namespace string, labels map[string]string) error {
	switch kind {
	case Deployment:
		{
			ctx, cancelFn := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancelFn()
			var dl appsv1.DeploymentList
			if err := c.List(
				ctx, &dl, &client.ListOptions{
					Namespace:     namespace,
					LabelSelector: apiLabels.SelectorFromValidatedSet(labels),
				},
			); err != nil {
				return err
			}

			for _, d := range dl.Items {
				if d.Spec.Template.ObjectMeta.Annotations == nil {
					d.Spec.Template.ObjectMeta.Annotations = map[string]string{}
				}
				// [source] (https://stackoverflow.com/a/59051313)
				d.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
				if err := c.Update(ctx, &d); err != nil {
					return err
				}
			}
		}
	case StatefulSet:
		{
			ctx, cancelFn := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancelFn()

			var sl appsv1.StatefulSetList

			if err := c.List(
				ctx, &sl, &client.ListOptions{
					Namespace:     namespace,
					LabelSelector: apiLabels.SelectorFromValidatedSet(labels),
				},
			); err != nil {
				return err
			}

			for _, d := range sl.Items {
				if d.Spec.Template.ObjectMeta.Annotations == nil {
					d.Spec.Template.ObjectMeta.Annotations = map[string]string{}
				}
				d.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
				if err := c.Update(ctx, &d); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func FilterObservabilityAnnotations(ann map[string]string) map[string]string {
	m := make(map[string]string, len(ann))
	for k, v := range ann {
		if strings.HasPrefix(k, "kloudlite.io/observability") {
			m[k] = v
		}
	}
	return m
}

func DeleteAndWait[T client.Object](ctx context.Context, logger logging.Logger, kcli client.Client, resources ...T) error {
	deletionStatus := make(map[string]bool)

	for i := range resources {
		resourceRef := fmt.Sprintf("resource (%s/%s) (gvk: %s)", resources[i].GetNamespace(), resources[i].GetName(), resources[i].GetObjectKind().GroupVersionKind().String())

		deletionStatus[resourceRef] = false

		if err := kcli.Get(ctx, client.ObjectKeyFromObject(resources[i]), resources[i]); err != nil {
			if apiErrors.IsNotFound(err) {
				deletionStatus[resourceRef] = true
				continue
			}
			return err
		}

		if resources[i].GetDeletionTimestamp() == nil {
			logger.Infof("deleting %s", resourceRef)

			if err := kcli.Delete(ctx, resources[i], &client.DeleteOptions{
				GracePeriodSeconds: New(int64(30)),
				PropagationPolicy:  New(metav1.DeletePropagationForeground),
			}); err != nil {
				if !apiErrors.IsNotFound(err) {
					return err
				}
				return fmt.Errorf("waiting for deletion for %s", resourceRef)
			}
		}
	}

	for k, v := range deletionStatus {
		if !v {
			return fmt.Errorf("waiting for (%s) to be removed from k8s", k)
		}
	}

	return nil
}

func ForceDelete[T client.Object](ctx context.Context, logger logging.Logger, kcli client.Client, resources ...T) error {
	deletionStatus := make(map[string]bool)

	for i := range resources {
		resourceRef := fmt.Sprintf("resource (%s/%s) (gvk: %s)", resources[i].GetNamespace(), resources[i].GetName(), resources[i].GetObjectKind().GroupVersionKind().String())

		deletionStatus[resourceRef] = false

		if err := kcli.Get(ctx, client.ObjectKeyFromObject(resources[i]), resources[i]); err != nil {
			if apiErrors.IsNotFound(err) {
				deletionStatus[resourceRef] = true
				continue
			}
			return err
		}

		if resources[i].GetDeletionTimestamp() == nil {
			logger.Infof("deleting %s", resourceRef)

			if err := kcli.Delete(ctx, resources[i], &client.DeleteOptions{
				GracePeriodSeconds: New(int64(30)),
				PropagationPolicy:  New(metav1.DeletePropagationForeground),
			}); err != nil {
				if !apiErrors.IsNotFound(err) {
					return err
				}
				return fmt.Errorf("waiting for deletion for %s", resourceRef)
			}
		}

		if controllerutil.ContainsFinalizer(resources[i], constants.CommonFinalizer) {
			controllerutil.RemoveFinalizer(resources[i], constants.CommonFinalizer)
			if err := kcli.Update(ctx, resources[i]); err != nil {
				return err
			}
			return fmt.Errorf("removing finalizers from resource %s", resourceRef)
		}
	}

	for k, v := range deletionStatus {
		if !v {
			return fmt.Errorf("waiting for (%s) to be removed from k8s", k)
		}
	}

	return nil
}

// checks whether CRD are installed on the cluster
func IsGVKInstalled(client client.Client, obj client.Object) bool {
	if _, err := client.IsObjectNamespaced(obj); err != nil && strings.HasPrefix(err.Error(), "failed to get restmapping: failed to find API group") {
		return false
	}
	return true
}
