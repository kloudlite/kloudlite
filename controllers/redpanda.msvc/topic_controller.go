package redpandamsvc

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	redpandamsvcv1 "operators.kloudlite.io/apis/redpanda.msvc/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// TopicReconciler reconciles a Topic object
type TopicReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	Name   string
}

func (r *TopicReconciler) GetName() string {
	return r.Name
}

const (
	TopicExists conditions.Type = "repanda.topics/exists"
)

// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=topics,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=topics/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=topics/finalizers,verbs=update

func (r *TopicReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &redpandamsvcv1.Topic{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.Logger.Infof("--------------------NEW RECONCILATION------------------")

	if x := req.EnsureLabelsAndAnnotations(); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	req.Logger.Infof("--------------------RECONCILATION FINISH------------------")

	return ctrl.Result{}, nil
}

func checkTopicExists(req *rApi.Request[*redpandamsvcv1.Topic], topicName string, redpandaHosts string) bool {
	cmd := exec.Command("rpk", "topic", "describe", topicName, "--brokers", redpandaHosts)
	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		req.Logger.Debugf(fmt.Sprintf("topic %s does not exist", topicName))
		return false
	}
	return true
}

func createTopic(req *rApi.Request[*redpandamsvcv1.Topic], topicName string, redpandaHosts string) error {
	cmd := exec.Command("rpk", "topic", "describe", topicName, "--brokers", redpandaHosts)
	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		req.Logger.Debugf(fmt.Sprintf("topic %s does not exist", topicName))
		return err
	}
	return nil
}

func (r *TopicReconciler) finalize(req *rApi.Request[*redpandamsvcv1.Topic]) stepResult.Result {
	return req.Finalize()
}

func (r *TopicReconciler) reconcileStatus(req *rApi.Request[*redpandamsvcv1.Topic]) stepResult.Result {
	ctx := req.Context()
	obj := req.Object

	isReady := true

	if value := obj.GetAnnotations()["kloudlite.io/reset-status-watcher"]; value == "true" {
		ann := obj.GetAnnotations()
		delete(ann, "kloudlite.io/reset-status-watcher")
		obj.SetAnnotations(ann)
		if err := r.Update(ctx, obj); err != nil {
			return req.FailWithStatusError(err)
		}

		obj.Status = rApi.Status{}
		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithStatusError(err)
		}
		return req.Done().RequeueAfter(0)
	}

	cs := make([]metav1.Condition, 0, 4)

	// managed service exists?
	msvc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.ManagedSvcName), &redpandamsvcv1.Service{})
	if err != nil {
		isReady = false
		if apiErrors.IsNotFound(err) {
			cs = append(cs, conditions.New(MsvcExists, false, conditions.NotFound, err.Error()))
			return req.FailWithStatusError(err, cs...).Err(nil)
		}
		return req.FailWithStatusError(err).Err(nil)
	} else {
		cs = append(cs, conditions.New(MsvcExists, true, conditions.Found))
	}

	if !msvc.Status.IsReady {
		isReady = false
		cs = append(cs, conditions.New(MsvcReady, false, conditions.NotReady))
		return req.FailWithStatusError(err, cs...).Err(nil)
	} else {
		cs = append(cs, conditions.New(MsvcReady, true, conditions.Ready))
	}

	// reading secret for accessing redpanda
	msvcOutput, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, "msvc-"+msvc.Name), &corev1.Secret{})
	if err != nil {
		isReady = false
		cs = append(cs, conditions.New(OutputExists, false, err.Error()))
		return req.FailWithStatusError(err, cs...).Err(nil)
	} else {
		cs = append(cs, conditions.New(OutputExists, true, conditions.Found))
	}

	b, ok := msvcOutput.Data[OutputRedpandaHostsKey]
	if !ok {
		isReady = false
		return req.FailWithStatusError(errors.NewEf(err, "could not read key %s from secret %s", OutputRedpandaHostsKey, obj.Namespace))
	}
	rApi.SetLocal(req, "redpanda-hosts", string(b))

	// check redpanda hosts
	if checkTopicExists(req, obj.Name, string(b)) {
		isReady = false
		cs = append(cs, conditions.New(TopicExists, false, conditions.NotFound))
	} else {
		cs = append(cs, conditions.New(TopicExists, true, conditions.Found))
	}

	nConditions, hasUpdated, err := conditions.Patch(obj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}
	if !hasUpdated && isReady == obj.Status.IsReady {
		return req.Next()
	}

	obj.Status.IsReady = isReady
	obj.Status.Conditions = nConditions
	if err := r.Status().Update(ctx, obj); err != nil {
		return req.FailWithStatusError(err)
	}
	return req.Done()
}

func (r *TopicReconciler) reconcileOperations(req *rApi.Request[*redpandamsvcv1.Topic]) stepResult.Result {
	ctx := req.Context()
	obj := req.Object

	if fn.ContainsFinalizers(obj, constants.CommonFinalizer, constants.ForegroundFinalizer) {
		controllerutil.AddFinalizer(obj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(obj, constants.ForegroundFinalizer)
		if err := r.Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	redpandaHosts, ok := rApi.GetLocal[string](req, "redpanda-hosts")
	if !ok {
		return req.FailWithOpError(errors.Newf("key 'redpanda-hosts' not found in req locals"))
	}

	if meta.IsStatusConditionFalse(obj.Status.Conditions, TopicExists.String()) {
		if err := createTopic(req, obj.Name, redpandaHosts); err != nil {
			return req.FailWithOpError(err)
		}
	}

	return req.Next()
}

// SetupWithManager sets up the controllers with the Manager.
func (r *TopicReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	return ctrl.NewControllerManagedBy(mgr).
		For(&redpandamsvcv1.Topic{}).
		Watches(
			&source.Kind{Type: &redpandamsvcv1.Service{}}, handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					var topicsList redpandamsvcv1.TopicList
					if err := r.List(
						context.TODO(), &topicsList, &client.ListOptions{
							LabelSelector: labels.SelectorFromValidatedSet(
								map[string]string{
									"kloudlite.io/msvc.name": obj.GetName(),
								},
							),
						},
					); err != nil {
						return nil
					}
					reqs := make([]reconcile.Request, 0, len(topicsList.Items))
					for _, item := range topicsList.Items {
						reqs = append(reqs, reconcile.Request{NamespacedName: fn.NN(item.Namespace, item.Name)})
					}
					return reqs
				},
			),
		).
		Complete(r)
}
