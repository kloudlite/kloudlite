package functions

import (
	"context"
	"encoding/json"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
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
