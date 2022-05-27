package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
)

type statusConditions struct {
	lt         metav1.Time
	conditions []metav1.Condition
}

func (s *statusConditions) Equal(other []metav1.Condition) bool {
	if len(s.conditions) != len(other) {
		return false
	}
	for _, c := range s.conditions {
		st := meta.FindStatusCondition(other, c.Type)
		if st == nil {
			return false
		}
		if st.Reason != c.Reason || st.Message != c.Message || st.Status != c.Status {
			return false
		}
	}
	return true
}

func (s *statusConditions) GetAll() []metav1.Condition {
	return s.conditions
}

func (s *statusConditions) Get(conditionType string) *metav1.Condition {
	return meta.FindStatusCondition(s.conditions, conditionType)
}

func (s *statusConditions) Build(group string, conditions ...metav1.Condition) {
	for _, cond := range conditions {
		if cond.Reason == "" {
			cond.Reason = "NotSpecified"
		}
		if !cond.LastTransitionTime.IsZero() {
			if cond.LastTransitionTime.Time.Sub(s.lt.Time).Seconds() > 0 {
				s.lt = cond.LastTransitionTime
			}
		}
		if cond.LastTransitionTime.IsZero() {
			cond.LastTransitionTime = s.lt
		}
		if group != "" {
			cond.Reason = fmt.Sprintf("%s:%s", group, cond.Reason)
			cond.Type = fmt.Sprintf("%s%s", group, cond.Type)
		}
		meta.SetStatusCondition(&s.conditions, cond)
	}
}

func (s *statusConditions) BuildFromHelmMsvc(
	ctx context.Context,
	apiClient client.Client,
	kind string,
	nn types.NamespacedName,
) error {
	hm := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": constants.MsvcApiVersion,
			"kind":       kind,
		},
	}

	if err := apiClient.Get(ctx, nn, &hm); err != nil {
		if apiErrors.IsNotFound(err) {
			s.Build(
				"Helm", metav1.Condition{
					Type:    "NotCreated",
					Status:  "True",
					Reason:  "HelmResourceNotFound",
					Message: err.Error(),
				},
			)
		}
		return err
	}

	b, err := hm.MarshalJSON()
	if err != nil {
		return err
	}
	var helmSvc struct {
		Status struct {
			Conditions []metav1.Condition `json:"Conditions,omitempty"`
		} `json:"status"`
	}
	if err := json.Unmarshal(b, &helmSvc); err != nil {
		return err
	}
	s.Build("Helm", helmSvc.Status.Conditions...)
	if !meta.IsStatusConditionTrue(helmSvc.Status.Conditions, "Deployed") {
		return errors.Newf("helm not yet deployed")
	}
	return nil
}

func (s *statusConditions) BuildFromStatefulset(
	ctx context.Context,
	apiClient client.Client,
	nn types.NamespacedName,
) error {
	sts := new(appsv1.StatefulSet)
	if err := apiClient.Get(ctx, nn, sts); err != nil {
		if apiErrors.IsNotFound(err) {
			s.Build(
				"StatefulSet", metav1.Condition{
					Type:    "NotCreated",
					Status:  "True",
					Reason:  "StsResourceNotFound",
					Message: err.Error(),
				},
			)
		}
		return err
	}

	fmt.Println("sts:", sts.Status.ReadyReplicas == sts.Status.Replicas)

	s.Build(
		"", metav1.Condition{
			Type: constants.ConditionReady.Type,
			Status: IfThenElse(
				sts.Status.ReadyReplicas == sts.Status.Replicas,
				metav1.ConditionTrue,
				metav1.ConditionFalse,
			),
			Reason:  "AllReplicasReady",
			Message: "StatefulSet Ready",
		},
	)

	podsList := new(corev1.PodList)
	if err := apiClient.List(
		ctx, podsList, &client.ListOptions{
			LabelSelector: apiLabels.SelectorFromValidatedSet(sts.Spec.Template.Labels),
			Namespace:     sts.Namespace,
		},
	); err != nil {
		return err
	}

	err := s.BuildFromPods(podsList.Items...)
	return err
}

func (s *statusConditions) BuildFromDeployment(
	ctx context.Context,
	apiClient client.Client,
	nn types.NamespacedName,
) error {
	depl := new(appsv1.Deployment)
	if err := apiClient.Get(ctx, nn, depl); err != nil {
		if apiErrors.IsNotFound(err) {
			s.Build(
				"Deployment", metav1.Condition{
					Type:    "NotCreated",
					Status:  "True",
					Reason:  "DeploymentNotFound",
					Message: err.Error(),
				},
			)
		}
		return err
	}
	var deplConditions []metav1.Condition
	for _, cond := range depl.Status.Conditions {
		deplConditions = append(
			deplConditions, metav1.Condition{
				Type:    string(cond.Type),
				Status:  metav1.ConditionStatus(cond.Status),
				Reason:  cond.Reason,
				Message: cond.Message,
			},
		)
	}

	s.Build("Deployment", deplConditions...)

	s.Build(
		"", metav1.Condition{
			Type: constants.ConditionReady.Type,
			Status: IfThenElse(
				meta.IsStatusConditionTrue(deplConditions, string(appsv1.DeploymentAvailable)),
				metav1.ConditionTrue,
				metav1.ConditionFalse,
			),
			Reason:  constants.ConditionReady.SuccessReason,
			Message: "Deployment is Available",
		},
	)

	opts := &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(depl.Spec.Template.GetLabels()),
		Namespace:     depl.Namespace,
	}
	podsList := new(corev1.PodList)
	if err := apiClient.List(ctx, podsList, opts); err != nil {
		return errors.NewEf(err, "could not list pods for deployment")
	}
	return s.BuildFromPods(podsList.Items...)
}

func (s *statusConditions) BuildFromPods(pl ...corev1.Pod) error {
	for idx, pod := range pl {
		var podC []metav1.Condition
		fmt.Printf(
			"pod info: Name: %s LenConditions: %d LenContainerStatus: %d\n",
			pod.Name,
			len(pod.Status.Conditions),
			len(pod.Status.ContainerStatuses),
		)
		for _, condition := range pod.Status.Conditions {
			podC = append(
				podC, metav1.Condition{
					Type:               fmt.Sprintf("Pod-idx-%d-%s", idx, condition.Type),
					Status:             metav1.ConditionStatus(condition.Status),
					LastTransitionTime: condition.LastTransitionTime,
					Reason:             fmt.Sprintf("Pod:Idx:%d:NotSpecified", idx),
					Message:            condition.Message,
				},
			)
		}
		s.Build("", podC...)
		var containerC []metav1.Condition
		for _, cs := range pod.Status.ContainerStatuses {
			p := metav1.Condition{
				Type:   fmt.Sprintf("Name-%s", cs.Name),
				Status: StatusFromBool(cs.Ready),
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
		s.Build("Container", containerC...)
		return nil
	}
	return nil
}

func (s *statusConditions) IsTrue(conditionType string) bool {
	return meta.IsStatusConditionTrue(s.conditions, conditionType)
}

func (s *statusConditions) IsFalse(conditionType string) bool {
	return meta.IsStatusConditionFalse(s.conditions, conditionType)
}

func (s *statusConditions) MarkNotReady(err error, reason ...string) {
	r := constants.ConditionReady.ErrorReason
	if len(reason) > 0 {
		r = reason[0]
	}
	s.SetReady(
		metav1.ConditionFalse,
		r,
		err.Error(),
	)
}

func (c *statusConditions) MarkReady(msg string, reason ...string) {
	r := constants.ConditionReady.SuccessReason
	if len(reason) > 0 {
		r = reason[0]
	}
	c.SetReady(
		metav1.ConditionFalse,
		r,
		msg,
	)
}

func (s *statusConditions) SetReady(t metav1.ConditionStatus, reason string, msg string) {
	s.Build(
		"", metav1.Condition{
			Type:    constants.ConditionReady.Type,
			Status:  t,
			Reason:  reason,
			Message: msg,
		},
	)
}

func (s *statusConditions) Reset() {
	s.conditions = []metav1.Condition{}
}

type StatusConditions interface {
	Build(group string, conditions ...metav1.Condition)
	BuildFromHelmMsvc(
		ctx context.Context,
		apiClient client.Client,
		kind string,
		nn types.NamespacedName,
	) error
	BuildFromStatefulset(
		ctx context.Context,
		apiClient client.Client,
		nn types.NamespacedName,
	) error
	BuildFromDeployment(
		ctx context.Context,
		apiClient client.Client,
		nn types.NamespacedName,
	) error
	BuildFromPods(pl ...corev1.Pod) error
	IsTrue(conditionType string) bool
	IsFalse(conditionType string) bool
	MarkNotReady(err error, reason ...string)
	MarkReady(msg string, reason ...string)
	SetReady(t metav1.ConditionStatus, reason string, msg string)
	Reset()

	GetAll() []metav1.Condition
	Get(conditionType string) *metav1.Condition
	Equal(other []metav1.Condition) bool
}

type wConditions struct {
	sc statusConditions
}

func (wc *wConditions) From(conditions []metav1.Condition) StatusConditions {
	wc.sc.conditions = conditions
	return &wc.sc
}

var Conditions = &wConditions{}

type conditions2 struct {
	lt metav1.Time
}

func (c *conditions2) Build(cl *[]metav1.Condition, group string, conditions ...metav1.Condition) {
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
		meta.SetStatusCondition(cl, cond)
	}
}

func (c *conditions2) MarkReady(cl *[]metav1.Condition, reason string, msg ...string) {
	c.Build(
		cl, "", metav1.Condition{
			Type:   constants.ConditionReady.Type,
			Status: metav1.ConditionTrue,
			Reason: reason,
			Message: IfThenElseFn(
				len(msg) > 0,
				func() string { return msg[0] },
				func() string { return "" },
			),
		},
	)
}

func (c *conditions2) MarkNotReady(cl *[]metav1.Condition, err error, reason ...string) {
	c.Build(
		cl, "", metav1.Condition{
			Type:   constants.ConditionReady.Type,
			Status: metav1.ConditionFalse,
			Reason: IfThenElseFn(
				len(reason) > 0,
				func() string { return reason[0] },
				func() string { return constants.ConditionReady.ErrorReason },
			),
			Message: err.Error(),
		},
	)
}

func (c *conditions2) Equal(first []metav1.Condition, second []metav1.Condition) bool {
	if len(first) != len(second) {
		return false
	}
	for _, c := range first {
		st := meta.FindStatusCondition(second, c.Type)
		if st == nil {
			return false
		}
		if st.Reason != c.Reason || st.Message != c.Message || st.Status != c.Status {
			return false
		}
	}
	return true
}

func (c *conditions2) BuildFromHelmMsvc(
	conditions *[]metav1.Condition,
	ctx context.Context,
	apiClient client.Client,
	kind string,
	nn types.NamespacedName,
) error {
	hm := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": constants.MsvcApiVersion,
			"kind":       kind,
		},
	}

	if err := apiClient.Get(ctx, nn, &hm); err != nil {
		if apiErrors.IsNotFound(err) {
			c.Build(
				conditions,
				"Helm", metav1.Condition{
					Type:    "NotCreated",
					Status:  "True",
					Reason:  "HelmResourceNotFound",
					Message: err.Error(),
				},
			)
		}
		return err
	}

	b, err := hm.MarshalJSON()
	if err != nil {
		return err
	}
	var helmSvc struct {
		Status struct {
			Conditions []metav1.Condition `json:"Conditions,omitempty"`
		} `json:"status"`
	}
	if err := json.Unmarshal(b, &helmSvc); err != nil {
		return err
	}
	c.Build(conditions, "Helm", helmSvc.Status.Conditions...)
	if !meta.IsStatusConditionTrue(helmSvc.Status.Conditions, "Deployed") {
		return errors.Newf("helm not yet deployed")
	}
	return nil
}

func (c *conditions2) BuildFromStatefulset(
	conditions *[]metav1.Condition,
	ctx context.Context,
	apiClient client.Client,
	nn types.NamespacedName,
) error {
	sts := new(appsv1.StatefulSet)
	if err := apiClient.Get(ctx, nn, sts); err != nil {
		if apiErrors.IsNotFound(err) {
			c.Build(
				conditions,
				"StatefulSet", metav1.Condition{
					Type:    "NotCreated",
					Status:  "True",
					Reason:  "StsResourceNotFound",
					Message: err.Error(),
				},
			)
		}
		return err
	}

	fmt.Println("sts:", sts.Status.ReadyReplicas == sts.Status.Replicas)

	c.Build(
		conditions,
		"", metav1.Condition{
			Type: constants.ConditionReady.Type,
			Status: IfThenElse(
				sts.Status.ReadyReplicas == sts.Status.Replicas,
				metav1.ConditionTrue,
				metav1.ConditionFalse,
			),
			Reason:  "AllReplicasReady",
			Message: "StatefulSet Ready",
		},
	)

	podsList := new(corev1.PodList)
	if err := apiClient.List(
		ctx, podsList, &client.ListOptions{
			LabelSelector: apiLabels.SelectorFromValidatedSet(sts.Spec.Template.Labels),
			Namespace:     sts.Namespace,
		},
	); err != nil {
		return err
	}

	err := c.BuildFromPods(conditions, podsList.Items...)
	return err
}

func (c *conditions2) BuildFromDeployment(
	conditions *[]metav1.Condition,
	ctx context.Context,
	apiClient client.Client,
	nn types.NamespacedName,
) error {
	depl := new(appsv1.Deployment)
	if err := apiClient.Get(ctx, nn, depl); err != nil {
		if apiErrors.IsNotFound(err) {
			c.Build(
				conditions,
				"Deployment", metav1.Condition{
					Type:    "NotCreated",
					Status:  "True",
					Reason:  "DeploymentNotFound",
					Message: err.Error(),
				},
			)
		}
		return err
	}
	var deplConditions []metav1.Condition
	for _, cond := range depl.Status.Conditions {
		deplConditions = append(
			deplConditions, metav1.Condition{
				Type:    string(cond.Type),
				Status:  metav1.ConditionStatus(cond.Status),
				Reason:  cond.Reason,
				Message: cond.Message,
			},
		)
	}

	c.Build(conditions, "Deployment", deplConditions...)

	c.Build(
		conditions,
		"", metav1.Condition{
			Type: constants.ConditionReady.Type,
			Status: IfThenElse(
				meta.IsStatusConditionTrue(deplConditions, string(appsv1.DeploymentAvailable)),
				metav1.ConditionTrue,
				metav1.ConditionFalse,
			),
			Reason:  constants.ConditionReady.SuccessReason,
			Message: "Deployment is Available",
		},
	)

	opts := &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(depl.Spec.Template.GetLabels()),
		Namespace:     depl.Namespace,
	}
	podsList := new(corev1.PodList)
	if err := apiClient.List(ctx, podsList, opts); err != nil {
		return errors.NewEf(err, "could not list pods for deployment")
	}
	return c.BuildFromPods(conditions, podsList.Items...)
}

func (c *conditions2) BuildFromPods(conditions *[]metav1.Condition, pl ...corev1.Pod) error {
	for idx, pod := range pl {
		var podC []metav1.Condition
		fmt.Printf(
			"pod info: Name: %s LenConditions: %d LenContainerStatus: %d\n",
			pod.Name,
			len(pod.Status.Conditions),
			len(pod.Status.ContainerStatuses),
		)
		for _, condition := range pod.Status.Conditions {
			podC = append(
				podC, metav1.Condition{
					Type:               fmt.Sprintf("Pod-idx-%d-%s", idx, condition.Type),
					Status:             metav1.ConditionStatus(condition.Status),
					LastTransitionTime: condition.LastTransitionTime,
					Reason:             fmt.Sprintf("Pod:Idx:%d:NotSpecified", idx),
					Message:            condition.Message,
				},
			)
		}
		c.Build(conditions, "", podC...)
		var containerC []metav1.Condition
		for _, cs := range pod.Status.ContainerStatuses {
			p := metav1.Condition{
				Type:   fmt.Sprintf("Name-%s", cs.Name),
				Status: StatusFromBool(cs.Ready),
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
		c.Build(conditions, "Container", containerC...)
		return nil
	}
	return nil
}

var Conditions2 = &conditions2{
	lt: metav1.Time{Time: time.UnixMilli(time.Now().Unix())},
}

func init() {
	Conditions.sc.lt = metav1.Time{Time: time.UnixMilli(time.Now().Unix())}
	//Conditions2.lt = metav1.Time{Time: time.Now()}
}
