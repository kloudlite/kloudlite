package types

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MessageReply struct {
	ClusterId string `json:"clusterId"`

	ProjectId  string `json:"projectId"`
	ResourceId string `json:"resourceId"`

	ChildConditions []metav1.Condition `json:"childConditions,omitempty"`
	Conditions      []metav1.Condition `json:"conditions,omitempty"`
	IsReady         bool               `json:"isReady,omitempty"`
	Key             string             `json:"key"`
	AccountId       string             `json:"accountId"`
	Billing         string             `json:"billing,omitempty"`
}

type MessageSender interface {
	SendMessage(ctx context.Context, key string, message MessageReply) error
}
