package conditions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func FromPod(ctx context.Context, client client.Client, nn types.NamespacedName) ([]metav1.Condition, error) {
	var podRes corev1.Pod
	err := client.Get(ctx, nn, &podRes)
	if err != nil {
		return nil, err
	}

	res := make([]metav1.Condition, len(podRes.Status.Conditions))
	for i, pc := range podRes.Status.Conditions {
		res[i] = metav1.Condition{
			Type:               fmt.Sprintf("pod.%s/%s", pc.Type, podRes.Name),
			Status:             metav1.ConditionStatus(pc.Status),
			LastTransitionTime: pc.LastTransitionTime,
			Reason:             pc.Reason,
			Message:            pc.Message,
		}
	}

	// for i, pcs := range podRes.Status.ContainerStatuses {
	// 	res[i] = metav1.Condition{
	// 		Type:    fmt.Sprintf("container"),
	// 		Status:  "",
	// 		Reason:  "",
	// 		Message: "",
	// 	}
	// }

	return res, nil
}

// func FromPod2(
// 	ctx context.Context,
// 	client client.Client,
// 	typeMeta metav1.TypeMeta,
// 	typePrefix string,
// 	nn types.NamespacedName,
// ) ([]metav1.Condition, error) {
// obj := fn.NewUnstructured(typeMeta)
// err := client.Get(ctx, nn, obj)
// if err != nil {
// 	return nil, err
// }
//
// b, err := json.Marshal(obj.Object)
// if err != nil {
// 	return nil, err
// }
//
// var j struct {
// 	Conditions        []metav1.Condition       `json:"conditions,omitempty"`
// 	ContainerStatuses []corev1.ContainerStatus `json:"containerStatuses,omitempty"`
// }
// err = json.Unmarshal(b, &j)
// if err != nil {
// 	return nil, err
// }
//
// res := make([]metav1.Condition, len(j.Conditions)+len(j.ContainerStatuses))
//
// for i, condition := range j.Conditions {
// 	res[i] = condition
// 	res[i].Type = ""
// }
//
// for i, condition := range obj.Object["status"].(statusStruct).Conditions {
// 	condition.Type = fmt.Sprintf("%s-%s", typePrefix, condition.Type)
// 	res[i] = condition
// }
//
// for _, cs := range obj.Object["status"].(statusStruct).ContainerStatuses {
// 	p := metav1.Condition{
// 		Type:   fmt.Sprintf("%s-container-%s", typePrefix, cs.Name),
// 		Status: fn.IfThenElse(cs.Ready, metav1.ConditionTrue, metav1.ConditionFalse),
// 	}
// 	if cs.State.Waiting != nil {
// 		p.Reason = cs.State.Waiting.Reason
// 		p.Message = cs.State.Waiting.Message
// 	}
// 	if cs.State.Running != nil {
// 		p.Reason = "Running"
// 		p.Message = fmt.Sprintf("Container running since %s", cs.State.Running.StartedAt.String())
// 	}
// 	res = append(res, p)
// }
// return res, nil
// }

func FromResource(ctx context.Context, client client.Client, typeMeta metav1.TypeMeta, typePrefix string, nn types.NamespacedName) ([]metav1.Condition, error) {
	obj := fn.NewUnstructured(typeMeta)
	if err := client.Get(ctx, nn, obj); err != nil {
		return nil, err
	}

	m, err := json.Marshal(obj.Object)
	if err != nil {
		return nil, err
	}

	var x struct {
		Status struct {
			Conditions []metav1.Condition `json:"conditions,omitempty"`
		} `json:"status"`
	}

	if err := json.Unmarshal(m, &x); err != nil {
		return nil, err
	}

	c := x.Status.Conditions

	res := make([]metav1.Condition, len(c))
	for i, condition := range c {
		condition.Type = fmt.Sprintf("%s%s", typePrefix, condition.Type)
		condition.Reason = fn.IfThenElse(len(condition.Reason) == 0, "NotSpecified", condition.Reason)
		condition.Message = fn.IfThenElse(len(condition.Message) == 0, "", condition.Message)
		res[i] = condition
	}

	return res, nil
}

func ParseFromResource(resource any, cTypePrefix string) ([]metav1.Condition, error) {
	m, err := json.Marshal(resource)
	if err != nil {
		return nil, err
	}

	var x struct {
		Status struct {
			Conditions []metav1.Condition `json:"conditions,omitempty"`
		} `json:"status"`
	}

	if err := json.Unmarshal(m, &x); err != nil {
		return nil, err
	}

	c := x.Status.Conditions

	res := make([]metav1.Condition, len(c))
	for i, condition := range c {
		condition.Type = fmt.Sprintf("%s%s", cTypePrefix, condition.Type)
		condition.Reason = fn.IfThenElse(len(condition.Reason) == 0, "NotSpecified", condition.Reason)
		condition.Message = fn.IfThenElse(len(condition.Message) == 0, "", condition.Message)
		res[i] = condition
	}

	return res, nil
}

func New[K Type | string, V Reason | string](cType K, status bool, reason V, msg ...string) metav1.Condition {
	s := metav1.ConditionFalse
	if status {
		s = metav1.ConditionTrue
	}

	msg = append(msg, "")

	return metav1.Condition{
		Type:    string(cType),
		Status:  s,
		Reason:  string(reason),
		Message: msg[0],
	}
}

type Type string

func (t Type) String() string {
	return string(t)
}

const ()

type Reason string
