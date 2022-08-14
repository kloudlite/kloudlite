package operator

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/logging"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rawJson "operators.kloudlite.io/lib/raw-json"
)

// +kubebuilder:object:generate=true

type Status struct {
	// +kubebuilder:validation:Optional
	IsReady         bool                `json:"isReady"`
	DisplayVars     rawJson.KubeRawJson `json:"displayVars,omitempty"`
	GeneratedVars   rawJson.KubeRawJson `json:"generatedVars,omitempty"`
	Conditions      []metav1.Condition  `json:"conditions,omitempty"`
	ChildConditions []metav1.Condition  `json:"childConditions,omitempty"`
	OpsConditions   []metav1.Condition  `json:"opsConditions,omitempty"`
	Generation      int                 `json:"generation,omitempty"`
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
	r.locals[key] = value
}

func Get[T client.Object](ctx context.Context, cli client.Client, nn types.NamespacedName, obj T) (T, error) {
	if err := cli.Get(ctx, nn, obj); err != nil {
		return obj, err
	}
	return obj, nil
}
