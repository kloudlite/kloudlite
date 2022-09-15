package operator

import (
	"context"
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/logging"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rawJson "operators.kloudlite.io/lib/raw-json"
)

type ResourceRef struct {
	Name string `json:"name"`
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace"`
	// +kubebuilder:validation:Optional
	ApiVersion string `json:"apiVersion"`
	// +kubebuilder:validation:Optional
	Kind string `json:"kind"`
}

// +kubebuilder:object:generate=true

type Check struct {
	Status     bool   `json:"status"`
	Message    string `json:"message,omitempty"`
	Generation int64  `json:"generation,omitempty"`
	// Resources     []ResourceRef `json:"resources,omitempty"`
	// LastCheckedAt metav1.Time `json:"lastCheckedAt,omitempty"`
}

// +kubebuilder:object:generate=true

type Status struct {
	// +kubebuilder:validation:Optional
	IsReady           bool               `json:"isReady"`
	Message           rawJson.RawJson    `json:"message,omitempty"`
	Messages          []ContainerMessage `json:"messages,omitempty"`
	DisplayVars       rawJson.RawJson    `json:"displayVars,omitempty"`
	GeneratedVars     rawJson.RawJson    `json:"generatedVars,omitempty"`
	Conditions        []metav1.Condition `json:"conditions,omitempty"`
	ChildConditions   []metav1.Condition `json:"childConditions,omitempty"`
	OpsConditions     []metav1.Condition `json:"opsConditions,omitempty"`
	Checks            map[string]Check   `json:"checks,omitempty"`
	LastReconcileTime metav1.Time        `json:"lastReconcileTime,omitempty"`
}

type Reconciler interface {
	reconcile.Reconciler
	SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error
	GetName() string
}

type Resource interface {
	client.Object
	runtime.Object
	GetStatus() *Status
	GetEnsuredLabels() map[string]string
	GetEnsuredAnnotations() map[string]string
}

func GetLocal[T any, V Resource](r *Request[V], key string) (T, bool) {
	x := r.locals[key]
	t, ok := x.(T)
	if !ok {
		return *new(T), ok
	}
	return t, ok
}

func SetLocal[T any, V Resource](r *Request[V], key string, value T) {
	if r.locals == nil {
		r.locals = map[string]any{}
	}
	r.locals[key] = value
}

func Get[T client.Object](ctx context.Context, cli client.Client, nn types.NamespacedName, obj T) (T, error) {
	if err := cli.Get(ctx, nn, obj); err != nil {
		// return obj, err
		return *new(T), err
	}
	return obj, nil
}

func GetRaw[T any](ctx context.Context, cli client.Client, nn types.NamespacedName, obj *unstructured.Unstructured) (*T, error) {
	// b, err := json.Marshal(obj)
	// if err != nil {
	// 	return nil, err
	// }
	// var m map[string]any
	// if err := json.Unmarshal(b, &m); err != nil {
	// 	return nil, err
	// }
	//
	// k := unstructured.Unstructured{
	// 	Object: obj,
	// }
	if err := cli.Get(ctx, nn, obj); err != nil {
		return nil, err
	}

	b, err := json.Marshal(obj.Object)
	if err != nil {
		return nil, err
	}
	var result T
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
