package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/redpanda"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	Billing         ResourceBilling    `json:"billing,omitempty"`
}

type Notifier struct {
	clusterId string
	producer  *redpanda.Producer
	topic     string
}

func (n *Notifier) notify(ctx context.Context, key string, metadata *KlMetadata, status rApi.Status) error {
	msg := MessageReply{
		ClusterId:       n.clusterId,
		AccountId:       metadata.AccountId,
		ProjectId:       metadata.ProjectId,
		ResourceId:      metadata.ResourceId,
		ChildConditions: status.ChildConditions,
		Conditions:      status.Conditions,
		IsReady:         status.IsReady,
		Key:             key,
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return n.producer.Produce(ctx, n.topic, key, b)
}

func (n *Notifier) notifyBilling(ctx context.Context, key string, metadata *KlMetadata, billing *ResourceBilling) error {
	msg := MessageReply{
		ClusterId:  n.clusterId,
		AccountId:  metadata.AccountId,
		ProjectId:  metadata.ProjectId,
		ResourceId: metadata.ResourceId,
		Billing:    *billing,
		Key:        key,
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

const (
	Shared_1x1    Plan = "shared_1x1"
	Dedicated_1x1 Plan = "dedicated_1x1"

	Shared_1x2    Plan = "shared_1x2"
	Dedicated_1x2 Plan = "dedicated_1x2"

	Shared_1x4    Plan = "shared_1x4"
	Dedicated_1x4 Plan = "dedicated_1x4"

	Storage Plan = "storage"
)

type k8sResource string

const (
	Pod k8sResource = "Pod"
	Pvc k8sResource = "Pvc"
)

type k8sItem struct {
	Type  k8sResource
	Count int
	Plan  Plan
	PlanQ float64
}

type ResourceBilling struct {
	Name        string    `json:"name"`
	ToBeDeleted bool      `json:"toBeDeleted,omitempty"`
	Items       []k8sItem `json:"items"`
}

type KlMetadata struct {
	AccountId  string `json:"accountId"`
	ProjectId  string
	ResourceId string
	Plan       string
}

func ExtractMetadata(obj client.Object) *KlMetadata {
	items := obj.GetAnnotations()
	return &KlMetadata{
		AccountId:  items["kloudlite.io/account-ref"],
		ProjectId:  items["kloudlite.io/project-ref"],
		ResourceId: items["kloudlite.io/ResourceBilling-ref"],
		Plan:       items["kloudkite.io/billing-plan"],
	}
}

type WrappedName struct {
	Name  string                  `json:"name"`
	Group schema.GroupVersionKind `json:"group"`
}

func getMsgKey(c client.Object) string {
	kind := c.GetObjectKind().GroupVersionKind().Kind
	return fmt.Sprintf("Kind=%s/Namespace=%s/Type=%s", kind, c.GetNamespace(), c.GetName())
}

func (w WrappedName) String() (string, error) {
	b, err := json.Marshal(w)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
