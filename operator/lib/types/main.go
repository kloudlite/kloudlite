package types

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:object:generate=true
type Conditions struct {
	lt         v1.Time        `json:"lt,omitempty"`
	Conditions []v1.Condition `json:"conditions,omitempty"`
}

func (c *Conditions) GetConditions() []v1.Condition {
	return c.Conditions
}

func (c *Conditions) GetCondition(t string) *v1.Condition {
	return meta.FindStatusCondition(c.Conditions, t)
}

func (c *Conditions) IsTrue(t string) bool {
	return meta.IsStatusConditionTrue(c.Conditions, t)
}

func (c *Conditions) IsFalse(t string) bool {
	return meta.IsStatusConditionFalse(c.Conditions, t)
}

func (c *Conditions) Reset() {
	c.Conditions = []v1.Condition{}
}

func (c *Conditions) Build(group string, conditions ...v1.Condition) {
	meta.SetStatusCondition(&c.Conditions, v1.Condition{
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

type WatchList []types.NamespacedName

func (wl *WatchList) Reset() {
	*wl = []types.NamespacedName{}
}

func (wl WatchList) Exists(nn types.NamespacedName) bool {
	for _, nName := range wl {
		if nName.String() == nn.String() {
			return true
		}
	}
	return false
}

func (wl *WatchList) Add(nn types.NamespacedName) {
	*wl = append(*wl, nn)
}
