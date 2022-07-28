package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"operators.kloudlite.io/lib/constants"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/redpanda"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type stageTT string

var Stages = struct {
	Exists  stageTT `json:"exists"`
	Deleted stageTT `json:"deleted"`
}{
	Exists:  "EXISTS",
	Deleted: "DELETED",
}

type MessageReply struct {
	ChildConditions []metav1.Condition `json:"childConditions,omitempty"`
	Conditions      []metav1.Condition `json:"conditions,omitempty"`
	IsReady         bool               `json:"isReady"`
	// ToBeDeleted     bool               `json:"toBeDeleted,omitempty"`
	Key      string          `json:"key"`
	Billing  ResourceBilling `json:"billing,omitempty"`
	Metadata KlMetadata      `json:"metadata,omitempty"`
	Stage    stageTT         `json:"stage"`
}

type Notifier struct {
	clusterId string
	producer  *redpanda.Producer
	topic     string
}

func (n *Notifier) notify(ctx context.Context, key string, metadata KlMetadata, status rApi.Status, stage stageTT) error {
	metadata.ClusterId = n.clusterId
	msg := MessageReply{
		Metadata:        metadata,
		ChildConditions: status.ChildConditions,
		Conditions:      status.Conditions,
		IsReady:         status.IsReady,
		Key:             key,
		Stage:           stage,
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return n.producer.Produce(ctx, n.topic, key, b)
}

func (n *Notifier) notifyBilling(ctx context.Context, key string, metadata KlMetadata, billing *ResourceBilling, stage stageTT) error {
	metadata.ClusterId = n.clusterId
	msg := MessageReply{
		Metadata: metadata,
		Billing:  *billing,
		Key:      key,
		Stage:    stage,
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return n.producer.Produce(ctx, n.topic, key, b)
}

func NewNotifier(clusterId string, producer *redpanda.Producer, topic string) *Notifier {
	return &Notifier{producer: producer, topic: topic, clusterId: clusterId}
}

type Plan string

type k8sResource string

const (
	Pod k8sResource = "Pod"
	Pvc k8sResource = "Pvc"
)

type k8sItem struct {
	Type     k8sResource `json:"type"`
	Count    int         `json:"count,omitempty"`
	Plan     Plan        `json:"plan,omitempty"`
	PlanQ    string      `json:"planQ,omitempty"`
	IsShared string      `json:"isShared,omitempty"`
}

func newK8sItem(obj client.Object, resType k8sResource, value int) k8sItem {
	return k8sItem{
		Type:     resType,
		Count:    value,
		Plan:     Plan(obj.GetAnnotations()[constants.AnnotationKeys.BillingPlan]),
		PlanQ:    obj.GetAnnotations()[constants.AnnotationKeys.BillableQuantity],
		IsShared: obj.GetAnnotations()[constants.AnnotationKeys.IsShared],
	}
}

type ResourceBilling struct {
	Name  string    `json:"name,omitempty"`
	Items []k8sItem `json:"items,omitempty"`
}

type KlMetadata struct {
	ClusterId        string                  `json:"clusterId"`
	AccountId        string                  `json:"accountId"`
	ProjectId        string                  `json:"projectId"`
	ResourceId       string                  `json:"resourceId"`
	GroupVersionKind schema.GroupVersionKind `json:"groupVersionKind"`
}

func ExtractMetadata(obj client.Object) KlMetadata {
	items := obj.GetAnnotations()
	return KlMetadata{
		AccountId:        items[constants.AnnotationKeys.Account],
		ProjectId:        items[constants.AnnotationKeys.Project],
		ResourceId:       items[constants.AnnotationKeys.Resource],
		GroupVersionKind: obj.GetObjectKind().GroupVersionKind(),
	}
}

type WrappedName struct {
	Name  string `json:"name"`
	Group string `json:"group"`
}

func getMsgKey(c client.Object) string {
	kind := c.GetObjectKind().GroupVersionKind().Kind
	return fmt.Sprintf("Kind=%s/Namespace=%s/Name=%s", kind, c.GetNamespace(), c.GetName())
}

func (w WrappedName) String() (string, error) {
	b, err := json.Marshal(w)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
