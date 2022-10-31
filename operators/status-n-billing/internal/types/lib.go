package types

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

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
	Key      string           `json:"key"`
	Billing  *ResourceBilling `json:"billing-watcher,omitempty"`
	Metadata KlMetadata       `json:"metadata,omitempty"`
	Stage    stageTT          `json:"stage"`
}

type Notifier struct {
	clusterId string
	producer  redpanda.Producer
	topic     string
}

func (n *Notifier) Notify(ctx context.Context, key string, metadata KlMetadata, status rApi.Status, stage stageTT) error {
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

	_, err = n.producer.Produce(ctx, n.topic, key, b)
	return err
}

func (n *Notifier) NotifyBilling(ctx context.Context, key string, metadata KlMetadata, billing *ResourceBilling, stage stageTT) error {
	metadata.ClusterId = n.clusterId
	msg := MessageReply{
		Metadata: metadata,
		Billing:  billing,
		Key:      key,
		Stage:    stage,
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = n.producer.Produce(ctx, n.topic, key, b)
	return err
}

func NewNotifier(clusterId string, producer redpanda.Producer, topic string) *Notifier {
	return &Notifier{producer: producer, topic: topic, clusterId: clusterId}
}

type Plan string

const (
	ComputeBasic      Plan = "Plan"
	ComputeGeneral    Plan = "General"
	ComputeHighMemory Plan = "HighMemory"

	BlockStorageDefault Plan = "Default"
	LambdaDefault       Plan = "Default"
)

type k8sResource string

const (
	Compute       k8sResource = "Compute"
	BlockStorage  k8sResource = "BlockStorage"
	ObjectStorage k8sResource = "ObjectStorage"
	Lambda        k8sResource = "Lambda"
	Ci            k8sResource = "Ci"
)

type K8sItem struct {
	Type         k8sResource `json:"type"`
	Count        int         `json:"count,omitempty"`
	Plan         Plan        `json:"plan,omitempty"`
	IsShared     string      `json:"isShared,omitempty"`
	PlanQuantity float32     `json:"planQuantity,omitempty"`
}

func NewK8sItem(obj client.Object, resType k8sResource, planQuantity float32, count int) K8sItem {
	kItem := K8sItem{
		Type:         resType,
		Count:        count,
		PlanQuantity: planQuantity,
	}

	switch resType {
	case Compute:
		{
			kItem.Plan = Plan(obj.GetAnnotations()[constants.AnnotationKeys.BillingPlan])
			kItem.IsShared = obj.GetAnnotations()[constants.AnnotationKeys.IsShared]
		}
	case BlockStorage:
		{
			kItem.Plan = BlockStorageDefault
		}
	case Lambda:
		{
			kItem.Plan = LambdaDefault
		}
	}

	return kItem
}

type ResourceBilling struct {
	Name  string    `json:"name,omitempty"`
	Items []K8sItem `json:"items,omitempty"`
}

type KlMetadata struct {
	ClusterId        string                  `json:"clusterId"`
	AccountId        string                  `json:"accountId"`
	ProjectId        string                  `json:"projectId"`
	ResourceId       string                  `json:"resourceId"`
	GroupVersionKind schema.GroupVersionKind `json:"groupVersionKind"`
	Labels           map[string]string       `json:"labels"`
}

func ExtractMetadata(obj client.Object) KlMetadata {
	ann := obj.GetAnnotations()
	return KlMetadata{
		AccountId:        ann[constants.AnnotationKeys.AccountRef],
		ProjectId:        ann[constants.AnnotationKeys.ProjectRef],
		ResourceId:       ann[constants.AnnotationKeys.ResourceRef],
		GroupVersionKind: obj.GetObjectKind().GroupVersionKind(),
		Labels:           obj.GetLabels(),
	}
}

type WrappedName struct {
	Name  string `json:"name"`
	Group string `json:"group"`
}

func (w WrappedName) ParseGroup() (*schema.GroupVersionKind, error) {
	gName, err := base64.StdEncoding.DecodeString(w.Group)
	if err != nil {
		return nil, err
	}
	var gvk schema.GroupVersionKind
	s := strings.Split(string(gName), ", ")
	gv := strings.Split(s[0], "/")
	gvk.Group = strings.TrimSpace(gv[0])
	gvk.Version = strings.TrimSpace(gv[1])
	if _, err := fmt.Sscanf(s[1], "Kind=%s", &gvk.Kind); err != nil {
		return nil, err
	}
	return &gvk, nil
}

func GetMsgKey(c client.Object) string {
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
