package lib

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MessageReply struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	Status     bool               `json:"status"`
}
