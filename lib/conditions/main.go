package conditions

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	fn "operators.kloudlite.io/lib/functions"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func Patch(dest []metav1.Condition, source []metav1.Condition) ([]metav1.Condition, bool, error) {
	res := make([]metav1.Condition, 0)
	updated := false
	for _, c := range dest {
		sourceCondition := meta.FindStatusCondition(source, c.Type)
		if sourceCondition != nil {
			if sourceCondition.Status != c.Status || sourceCondition.Reason != c.Reason || sourceCondition.Message != c.Message {
				updated = true
				if c.LastTransitionTime.IsZero() {
					sourceCondition.LastTransitionTime = metav1.Time{
						Time: time.UnixMilli(time.Now().Unix()),
					}
				}
				res = append(res, *sourceCondition)
			}
			res = append(res, c)
		}
	}
	return res, updated, nil
}

func FromPod(
	ctx context.Context,
	client client.Client,
	groupVersionKind metav1.GroupVersionKind,
	typePrefix string,
	nn types.NamespacedName) ([]metav1.Condition, error) {
	type statusStruct struct {
		Conditions        []metav1.Condition       `json:"conditions,omitempty"`
		ContainerStatuses []corev1.ContainerStatus `json:"containerStatuses,omitempty"`
	}
	obj := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": strings.Split(groupVersionKind.String(), ", ")[0],
			"kind":       groupVersionKind.Kind,
			"status": statusStruct{
				Conditions: []metav1.Condition{},
			},
		},
	}

	err := client.Get(ctx, nn, &obj)
	if err != nil {
		return nil, err
	}
	res := make([]metav1.Condition, len(obj.Object["status"].(statusStruct).Conditions))

	for i, condition := range obj.Object["status"].(statusStruct).Conditions {
		condition.Type = fmt.Sprintf("%s-%s", typePrefix, condition.Type)
		res[i] = condition
	}

	for _, cs := range obj.Object["status"].(statusStruct).ContainerStatuses {
		p := metav1.Condition{
			Type:   fmt.Sprintf("%s-container-%s", typePrefix, cs.Name),
			Status: fn.IfThenElse(cs.Ready, metav1.ConditionTrue, metav1.ConditionFalse),
		}
		if cs.State.Waiting != nil {
			p.Reason = cs.State.Waiting.Reason
			p.Message = cs.State.Waiting.Message
		}
		if cs.State.Running != nil {
			p.Reason = "Running"
			p.Message = fmt.Sprintf("Container running since %s", cs.State.Running.StartedAt.String())
		}
		res = append(res, p)
	}
	return res, nil
}

func FromResource(
	ctx context.Context,
	client client.Client,
	groupVersionKind metav1.GroupVersionKind,
	typePrefix string,
	nn types.NamespacedName) ([]metav1.Condition, error) {
	type statusStruct struct {
		Conditions []metav1.Condition `json:"conditions,omitempty"`
	}
	obj := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": strings.Split(groupVersionKind.String(), ", ")[0],
			"kind":       groupVersionKind.Kind,
			"status": statusStruct{
				Conditions: []metav1.Condition{},
			},
		},
	}

	err := client.Get(ctx, nn, &obj)
	if err != nil {
		return nil, err
	}
	res := make([]metav1.Condition, len(obj.Object["status"].(statusStruct).Conditions))
	for i, condition := range obj.Object["status"].(statusStruct).Conditions {
		condition.Type = fmt.Sprintf("%s-%s", typePrefix, condition.Type)
		res[i] = condition
	}
	return res, nil
}
