package pipeline_run_watcher

import (
	"context"
	"encoding/json"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	"operators.kloudlite.io/lib/redpanda"
	"operators.kloudlite.io/operators/status-n-billing/internal/env"
	"operators.kloudlite.io/operators/status-n-billing/internal/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Reconciler reconciles a StatusWatcher object
type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Producer   redpanda.Producer
	KafkaTopic string
	logger     logging.Logger
	Name       string
	Env        *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

func (r *Reconciler) SendStatusEvents(ctx context.Context, obj client.Object) error {
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	var j struct {
		Status struct {
			CompletionTime *time.Time         `json:"completionTime,omitempty"`
			Conditions     []metav1.Condition `json:"conditions"`
		}
	}

	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}

	successCondition := meta.FindStatusCondition(j.Status.Conditions, "Succeeded")
	if successCondition == nil {
		return nil
	}

	if obj.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(obj, constants.StatusWatcherFinalizer) {
			if err := r.dispatchMsg(
				ctx, types.GetMsgKey(obj), Msg{
					PipelineRunId: obj.GetName(),
					PipelineId:    obj.GetLabels()["app"],
					StartTime:     obj.GetCreationTimestamp().Time,
					EndTime:       j.Status.CompletionTime,
					Success:       successCondition.Status == metav1.ConditionTrue,
					Message:       successCondition.Message,
				},
			); err != nil {
				return err
			}
			return r.RemoveWatcherFinalizer(ctx, obj)
		}
		return nil
	}

	if !controllerutil.ContainsFinalizer(obj, constants.StatusWatcherFinalizer) {
		return r.AddWatcherFinalizer(ctx, obj)
	}

	return r.dispatchMsg(
		ctx, types.GetMsgKey(obj), Msg{
			PipelineRunId: obj.GetName(),
			PipelineId:    obj.GetLabels()["app"],
			StartTime:     obj.GetCreationTimestamp().Time,
			EndTime:       j.Status.CompletionTime,
			Success:       successCondition.Status == metav1.ConditionTrue,
			Message:       successCondition.Message,
		},
	)
}

type Msg struct {
	PipelineRunId string     `json:"pipeline_run_id"`
	PipelineId    string     `json:"pipeline_id"`
	StartTime     time.Time  `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	Success       bool       `json:"success"`
	Message       string     `json:"message,omitempty"`
}

func (r *Reconciler) dispatchMsg(ctx context.Context, key string, msg Msg) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = r.Producer.Produce(ctx, r.KafkaTopic, key, b)
	return err
}

// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=statuswatchers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=statuswatchers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=statuswatchers/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.logger.WithName(req.NamespacedName.String())
	logger.Infof("request received ...")
	defer func() {
		logger.Infof("processed request ...")
	}()

	obj := fn.NewUnstructured(constants.TektonPipelineRunKind)
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.SendStatusEvents(ctx, obj); err != nil {
		return ctrl.Result{}, err
	}

	// return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
	return ctrl.Result{}, nil
}

func (r *Reconciler) AddWatcherFinalizer(ctx context.Context, obj client.Object) error {
	controllerutil.AddFinalizer(obj, constants.StatusWatcherFinalizer)
	return r.Update(ctx, obj)
}

func (r *Reconciler) RemoveWatcherFinalizer(ctx context.Context, obj client.Object) error {
	controllerutil.RemoveFinalizer(obj, constants.StatusWatcherFinalizer)
	return r.Update(ctx, obj)
}

// SetupWithManager sets up the controllers with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr)
	builder.For(fn.NewUnstructured(constants.TektonPipelineRunKind))

	return builder.Complete(r)
}
