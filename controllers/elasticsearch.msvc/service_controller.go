package elasticsearchmsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	ct "operators.kloudlite.io/apis/common-types"
	"operators.kloudlite.io/lib/conditions"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	elasticSearch "operators.kloudlite.io/apis/elasticsearch.msvc/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	env    *env.Env
	Name   string
}

func (r *ServiceReconciler) GetName() string {
	return r.Name
}

const (
	KeyHelmReady   string = "helm-ready"
	KeyOutputReady string = "output-ready"
	KeyStsReady    string = "sts-ready"
)

// +kubebuilder:rbac:groups=elasticsearch.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=elasticsearch.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=elasticsearch.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &elasticSearch.Service{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
	}

	req.Logger.Infof("-------------------- NEW RECONCILATION------------------")

	if req.Object.GetAnnotations()["kloudlite.io/re-check"] == "true" {
		ann := req.Object.Annotations
		delete(ann, "kloudlite.io/re-check")
		req.Object.SetAnnotations(ann)

		if err := r.Update(ctx, req.Object); err != nil {
			return ctrl.Result{}, err
		}

		req.Object.Status.Checks = nil
		if err := r.Status().Update(ctx, req.Object); err != nil {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{RequeueAfter: 0}, nil
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureChecks(req, KeyOutputReady, KeyHelmReady, KeyStsReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ouptut(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.helm(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.statefulSetsAndPods(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{RequeueAfter: r.env.ReconcilePeriod * time.Second}, r.Status().Update(ctx, req.Object)
}

func (r *ServiceReconciler) ensureChecks(req *rApi.Request[*elasticSearch.Service], names ...string) stepResult.Result {
	obj, ctx, checks := req.Object, req.Context(), req.Object.Status.Checks
	nChecks := len(checks)

	if checks == nil {
		checks = map[string]rApi.Check{}
	}

	for i := range names {
		if _, ok := checks[names[i]]; !ok {
			checks[names[i]] = rApi.Check{}
		}
	}

	if nChecks != len(checks) {
		obj.Status.Checks = checks
		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return stepResult.New().RequeueAfter(1 * time.Second)
	}
	return req.Next()
}

func (r *ServiceReconciler) ouptut(req *rApi.Request[*elasticSearch.Service]) stepResult.Result {
	obj, ctx, checks := req.Object, req.Context(), req.Object.Status.Checks

	output, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, "msvc-"+obj.Name), &corev1.Secret{})
	if err != nil {
		output = nil
	}
	if output == nil || obj.Generation > checks[KeyOutputReady].Generation {
		check := rApi.Check{Generation: obj.Generation}

		username := "elastic"
		host := fmt.Sprintf("%s.%s.svc.cluster.local:9200", obj.Name, obj.Namespace)

		b, err := templates.Parse(
			templates.Secret, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "msvc-" + obj.Name,
					Namespace: obj.Namespace,
				},
				Immutable: fn.New(true),
				StringData: map[string]string{
					"USERNAME": username,
					"PASSWORD": obj.Spec.Auth.Password,
					"HOSTS":    host,
					"URI":      fmt.Sprintf("http://%s:%s@%s", username, obj.Spec.Auth.Password, host),
				},
			},
		)
		if err != nil {
			return req.CheckFailed(KeyOutputReady, check, err.Error())
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			check.Message = err.Error()
			return req.CheckFailed(KeyOutputReady, check, err.Error())
		}

		check.Status = true
		if check != checks[KeyOutputReady] {
			checks[KeyOutputReady] = check
			if err := r.Status().Update(ctx, obj); err != nil {
				return req.FailWithOpError(err)
			}
			return req.Done().RequeueAfter(2 * time.Second)
		}
	}

	check := rApi.Check{Generation: obj.Generation}
	check.Status = true
	if check != checks[KeyOutputReady] {
		checks[KeyOutputReady] = check
		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
	}

	return req.Next()
}

func (r *ServiceReconciler) helm(req *rApi.Request[*elasticSearch.Service]) stepResult.Result {
	obj, ctx, checks := req.Object, req.Context(), req.Object.Status.Checks

	helmRes, _ := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.HelmElasticType))

	if helmRes == nil || obj.Generation > checks[KeyHelmReady].Generation {
		check := rApi.Check{Generation: obj.Generation}

		sc, err := obj.Spec.CloudProvider.GetStorageClass(ct.Ext4)
		if err != nil {
			return req.CheckFailed(KeyHelmReady, check, err.Error())
		}

		b, err := templates.Parse(
			templates.ElasticSearch, map[string]any{
				"object":        obj,
				"storage-class": sc,
				"owner-refs":    []metav1.OwnerReference{fn.AsOwner(obj, true)},
			},
		)
		if err != nil {
			return req.CheckFailed(KeyHelmReady, check, err.Error())
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(KeyHelmReady, check, err.Error())
		}

		obj.Status.Checks[KeyHelmReady] = check
		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}

		return req.Done().RequeueAfter(2 * time.Second)
	}

	check := rApi.Check{Generation: obj.Generation}
	cds, err := conditions.FromObject(helmRes)
	if err != nil {
		return req.CheckFailed(KeyHelmReady, check, err.Error())
	}

	deployedC := meta.FindStatusCondition(cds, "Deployed")
	if deployedC == nil || deployedC.Status == metav1.ConditionUnknown {
		return req.CheckFailed(KeyHelmReady, check, err.Error()).RequeueAfter(2 * time.Second)
	}

	if deployedC.Status == metav1.ConditionFalse {
		check.Status = false
		check.Message = deployedC.Message
		return req.CheckFailed(KeyHelmReady, check, err.Error())
	}

	if deployedC.Status == metav1.ConditionTrue {
		check.Status = true
	}

	if check != checks[KeyHelmReady] {
		checks[KeyHelmReady] = check
		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done().RequeueAfter(2 * time.Second)
	}

	return req.Next()
}

func (r *ServiceReconciler) statefulSetsAndPods(req *rApi.Request[*elasticSearch.Service]) stepResult.Result {
	obj, ctx, checks := req.Object, req.Context(), req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	var stsList appsv1.StatefulSetList
	if err := r.List(
		ctx, &stsList, &client.ListOptions{
			Namespace: obj.Namespace,
			LabelSelector: labels.SelectorFromValidatedSet(
				map[string]string{
					constants.MsvcNameKey: obj.Name,
				},
			),
		},
	); err != nil {
		return req.CheckFailed(KeyStsReady, check, err.Error())
	}

	for i := range stsList.Items {
		item := stsList.Items[i]
		if item.Status.ReadyReplicas != item.Status.Replicas {
			check.Status = false

			var podsList corev1.PodList
			if err := r.List(
				ctx, &podsList, &client.ListOptions{
					LabelSelector: labels.SelectorFromValidatedSet(
						map[string]string{constants.MsvcNameKey: obj.Name},
					),
				},
			); err != nil {
				return req.FailWithOpError(err)
			}

			messages := rApi.GetMessagesFromPods(podsList.Items...)
			if len(messages) > 0 {
				b, err := json.Marshal(messages)
				if err != nil {
					check.Message = err.Error()
					return req.CheckFailed(KeyStsReady, check, err.Error())
				}

				check.Message = string(b)
			}
		}
	}

	if check != checks[KeyStsReady] {
		checks[KeyStsReady] = check
		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
	}

	if !checks[KeyStsReady].Status {
		return req.Done()
	}

	return req.Next()
}

func (r *ServiceReconciler) finalize(req *rApi.Request[*elasticSearch.Service]) stepResult.Result {
	return req.Finalize()
}

func eventFilters() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(ev event.UpdateEvent) bool {
			return ev.ObjectNew.GetGeneration() != ev.ObjectOld.GetGeneration()
		},
	}
}

// SetupWithManager sets up the controllers with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.env = envVars

	builder := ctrl.NewControllerManagedBy(mgr)
	builder.For(&elasticSearch.Service{})
	builder.Owns(&corev1.Secret{})

	watchList := []client.Object{
		&appsv1.StatefulSet{},
		&corev1.Pod{},
	}

	for i := range watchList {
		builder.Watches(
			&source.Kind{Type: watchList[i]}, handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					msvcName := obj.GetLabels()[constants.MsvcNameKey]
					if msvcName != "" {
						return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), msvcName)}}
					}
					return nil
				},
			),
		)
	}

	builder.WithEventFilter(eventFilters())

	return builder.Complete(r)
}
