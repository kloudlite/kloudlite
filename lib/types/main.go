package types

import (
	"context"
	"encoding/json"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
)

// +kubebuilder:object:generate=true
type Conditions struct {
	lt         metav1.Time        `json:"lt,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

func (c *Conditions) GetConditions() []metav1.Condition {
	return c.Conditions
}

func (c *Conditions) GetCondition(t string) *metav1.Condition {
	return meta.FindStatusCondition(c.Conditions, t)
}

func (c *Conditions) IsTrue(t string) bool {
	return meta.IsStatusConditionTrue(c.Conditions, t)
}

func (c *Conditions) IsFalse(t string) bool {
	return meta.IsStatusConditionFalse(c.Conditions, t)
}

func (c *Conditions) Reset() {
	c.Conditions = []metav1.Condition{}
}

func (c *Conditions) Remove(t string) {
	meta.RemoveStatusCondition(&c.Conditions, t)
}

func (c *Conditions) Build(group string, conditions ...metav1.Condition) {
	for _, cond := range conditions {
		if cond.Reason == "" {
			cond.Reason = "NotSpecified"
		}
		if !cond.LastTransitionTime.IsZero() {
			if cond.LastTransitionTime.Time.Sub(c.lt.Time).Seconds() > 0 {
				c.lt = cond.LastTransitionTime
			}
		}
		if cond.LastTransitionTime.IsZero() {
			cond.LastTransitionTime = c.lt
		}
		if group != "" {
			cond.Reason = fmt.Sprintf("%s:%s", group, cond.Reason)
			cond.Type = fmt.Sprintf("%s%s", group, cond.Type)
		}
		meta.SetStatusCondition(&c.Conditions, cond)
	}
}

type HelmResource struct {
	Status struct {
		Conditions []metav1.Condition `json:"conditions,omitempty"`
	} `json:"status"`
}

func (c *Conditions) BuildFromHelmMsvc(ctx context.Context, apiClient client.Client, kind string, nn types.NamespacedName) error {
	hm := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": constants.MsvcApiVersion,
			"kind":       kind,
		},
	}
	if err := apiClient.Get(ctx, nn, &hm); err != nil {
		return err
	}
	b, err := hm.MarshalJSON()
	if err != nil {
		return err
	}
	var helmSvc struct {
		Status struct {
			Conditions []metav1.Condition `json:"conditions,omitempty"`
		} `json:"status"`
	}
	if err := json.Unmarshal(b, &helmSvc); err != nil {
		return err
	}
	c.Build("Helm", helmSvc.Status.Conditions...)
	if !meta.IsStatusConditionTrue(helmSvc.Status.Conditions, "Deployed") {
		return errors.Newf("helm not yet deployed")
	}
	return nil
}

func (c *Conditions) BuildFromStatefulset(ctx context.Context, apiClient client.Client, nn types.NamespacedName) error {
	sts := new(appsv1.StatefulSet)
	if err := apiClient.Get(ctx, nn, sts); err != nil {
		return err
	}

	fmt.Println("sts:", sts.Status.ReadyReplicas == sts.Status.Replicas)

	c.Build("", metav1.Condition{
		Type:    constants.ConditionReady.Type,
		Status:  fn.IfThenElse(sts.Status.ReadyReplicas == sts.Status.Replicas, metav1.ConditionTrue, metav1.ConditionFalse).(metav1.ConditionStatus),
		Reason:  "AllReplicasReady",
		Message: "StatefulSet Ready",
	})

	fmt.Println("sts2 :", meta.IsStatusConditionTrue(c.Conditions, "Ready"))

	podsList := new(corev1.PodList)
	if err := apiClient.List(ctx, podsList, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(sts.Spec.Template.Labels),
		Namespace:     sts.Namespace,
	}); err != nil {
		return err
	}

	err := c.BuildFromPods(podsList.Items...)
	return err
}

func (c *Conditions) BuildFromDeployment(ctx context.Context, apiClient client.Client, nn types.NamespacedName) error {
	depl := new(appsv1.Deployment)
	if err := apiClient.Get(ctx, nn, depl); err != nil {
		return err
	}
	var deplConditions []metav1.Condition
	for _, cond := range depl.Status.Conditions {
		deplConditions = append(deplConditions, metav1.Condition{
			Type:    string(cond.Type),
			Status:  metav1.ConditionStatus(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	c.Build("Deployment", deplConditions...)

	c.Build("", metav1.Condition{
		Type:    constants.ConditionReady.Type,
		Status:  fn.IfThenElse(meta.IsStatusConditionTrue(deplConditions, string(appsv1.DeploymentAvailable)), metav1.ConditionTrue, metav1.ConditionFalse).(metav1.ConditionStatus),
		Reason:  constants.ConditionReady.SuccessReason,
		Message: "Deployment is Available",
	})

	opts := &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(depl.Spec.Template.GetLabels()),
		Namespace:     depl.Namespace,
	}
	podsList := new(corev1.PodList)
	if err := apiClient.List(ctx, podsList, opts); err != nil {
		return errors.NewEf(err, "could not list pods for deployment")
	}
	return c.BuildFromPods(podsList.Items...)
}

func (c *Conditions) BuildFromPods(pl ...corev1.Pod) error {
	for idx, pod := range pl {
		var podC []metav1.Condition
		fmt.Printf("pod info: Name: %s LenConditions: %d LenContainerStatus: %d\n", pod.Name, len(pod.Status.Conditions), len(pod.Status.ContainerStatuses))
		for _, condition := range pod.Status.Conditions {
			podC = append(podC, metav1.Condition{
				Type:               fmt.Sprintf("Pod-idx-%d-%s", idx, condition.Type),
				Status:             metav1.ConditionStatus(condition.Status),
				LastTransitionTime: condition.LastTransitionTime,
				Reason:             fmt.Sprintf("Pod:Idx:%d:NotSpecified", idx),
				Message:            condition.Message,
			})
		}
		c.Build("", podC...)
		var containerC []metav1.Condition
		for _, cs := range pod.Status.ContainerStatuses {
			p := metav1.Condition{
				Type:   fmt.Sprintf("Name-%s", cs.Name),
				Status: fn.StatusFromBool(cs.Ready),
			}
			if cs.State.Waiting != nil {
				p.Reason = cs.State.Waiting.Reason
				p.Message = cs.State.Waiting.Message
			}
			if cs.State.Running != nil {
				p.Reason = "Running"
				p.Message = fmt.Sprintf("Container running since %s", cs.State.Running.StartedAt.String())
			}
			containerC = append(containerC, p)
		}
		c.Build("Container", containerC...)
		return nil
	}
	return nil
}
