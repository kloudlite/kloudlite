package lib

import (
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
