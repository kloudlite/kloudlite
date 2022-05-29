package conditions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	fn "operators.kloudlite.io/lib/functions"
)

func reasondiff(r1, r2 string) bool {
	if r1 == "" {
		r1 = "NotSpecified"
	}
	if r2 == "" {
		r2 = "NotSpecified"
	}
	return r1 != r2
}

func Patch(dest []metav1.Condition, source []metav1.Condition) ([]metav1.Condition, bool, error) {
	res := make([]metav1.Condition, 0)
	x := metav1.Time{Time: time.UnixMilli(time.Now().UnixMilli())}
	// x := metav1.Now()
	updated := false
	for _, c := range dest {
		sourceCondition := meta.FindStatusCondition(source, c.Type)
		if sourceCondition == nil {
			updated = true
			continue
		}

		if sourceCondition.Status != c.Status || reasondiff(
			sourceCondition.Reason,
			c.Reason,
		) || sourceCondition.Message != c.Message {
			updated = true

			if sourceCondition.LastTransitionTime.IsZero() {
				sourceCondition.LastTransitionTime = x
			}

			if sourceCondition.Reason == "" {
				sourceCondition.Reason = "NotSpecified"
			}

			res = append(res, *sourceCondition)
			continue
		}
		res = append(res, c)
	}

	for _, c := range source {
		if meta.FindStatusCondition(dest, c.Type) == nil {
			updated = true
			if c.LastTransitionTime.IsZero() {
				c.LastTransitionTime = x
			}
			if c.Reason == "" {
				c.Reason = "NotSpecified"
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
	obj := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": strings.Split(groupVersionKind.String(), ", ")[0],
			"kind":       groupVersionKind.Kind,
		},
	}

	if err := client.Get(ctx, nn, &obj); err != nil {
		return nil, err
	}

	type X struct {
		Status struct {
			Conditions []metav1.Condition `json:"conditions,omitempty"`
		} `json:"status"`
	}

	m, err := json.Marshal(obj.Object)
	if err != nil {
		return nil, err
	}
	var x X
	if err := json.Unmarshal(m, &x); err != nil {
		return nil, err
	}

	c := x.Status.Conditions

	res := make([]metav1.Condition, len(c))
	for i, condition := range c {
		condition.Type = fmt.Sprintf("%s-%s", typePrefix, condition.Type)
		condition.Reason = fn.IfThenElse(len(condition.Reason) == 0, "NotSpecified", condition.Reason)
		condition.Message = fn.IfThenElse(len(condition.Message) == 0, "", condition.Message)
		res[i] = condition
	}
	return res, nil
}

func New(cType string, status bool, reason string, msg ...string) metav1.Condition {
	s := metav1.ConditionFalse
	if status {
		s = metav1.ConditionTrue
	}

	msg = append(msg, "")

	return metav1.Condition{
		Type:    cType,
		Status:  s,
		Reason:  reason,
		Message: msg[0],
	}
}
