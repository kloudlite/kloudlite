package lib

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MessageReply struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	Status     bool               `json:"status"`
	Key        string             `json:"key"`
}

type Notifier interface {
	client.Client
	MessageSender
}

type MessageSender interface {
	SendMessage(key string, msg MessageReply) error
}

type LabelKey string

func (lk LabelKey) String() string {
	return string(lk)
}

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

func (c *Conditions) Build(group string, conditions ...metav1.Condition) {
	meta.SetStatusCondition(&c.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             "False",
		Reason:             "ChecksNotCompleted",
		LastTransitionTime: c.lt,
		Message:            "Not All Checks completed",
	})

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
