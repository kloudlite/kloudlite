package types

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MessageReply struct {
	ChildConditions []metav1.Condition `json:"childConditions,omitempty"`
	Conditions      []metav1.Condition `json:"conditions,omitempty"`
	IsReady         bool               `json:"isReady,omitempty"`
	Key             string             `json:"key"`
}

type MessageSender interface {
	SendMessage(ctx context.Context, key string, message MessageReply) error
}
