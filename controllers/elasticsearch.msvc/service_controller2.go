package elasticsearchmsvc

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"time"
//
// 	"sigs.k8s.io/controllers-runtime/pkg/event"
// 	"sigs.k8s.io/controllers-runtime/pkg/predicate"
//
// 	appsv1 "k8s.io/api/apps/v1"
// 	corev1 "k8s.io/api/core/v1"
// 	"k8s.io/apimachinery/pkg/api/meta"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/labels"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	ct "operators.kloudlite.io/apis/common-types"
// 	elasticSearch "operators.kloudlite.io/apis/elasticsearch.msvc/v1"
// 	"operators.kloudlite.io/env"
// 	"operators.kloudlite.io/lib/conditions"
// 	"operators.kloudlite.io/lib/constants"
// 	"operators.kloudlite.io/lib/errors"
// 	fn "operators.kloudlite.io/lib/functions"
// 	"operators.kloudlite.io/lib/logging"
// 	rApi "operators.kloudlite.io/lib/operator"
// 	stepResult "operators.kloudlite.io/lib/operator/step-result"
// 	"operators.kloudlite.io/lib/templates"
// 	ctrl "sigs.k8s.io/controllers-runtime"
// 	"sigs.k8s.io/controllers-runtime/pkg/client"
// 	"sigs.k8s.io/controllers-runtime/pkg/handler"
// 	"sigs.k8s.io/controllers-runtime/pkg/reconcile"
// 	"sigs.k8s.io/controllers-runtime/pkg/source"
// )

// ServiceReconciler reconciles a Service object
// type ServiceReconciler struct {
// 	client.Client
// 	Scheme *runtime.Scheme
// 	logger logging.Logger
// 	Name   string
// }
//
// func (r *ServiceReconciler) GetName() string {
// 	return r.Name
// }
//
// // const (
// // 	HelmElasticExists string = "helm.elasticSearch/Exists"
// // 	HelmElasticReady  string = "helm.elasticSearch/Ready"
// // )
//
// const (
// 	KeyVarsGenerated string = "vars-generated"
// 	KeyHelmExists    string = "helm-exists"
// 	KeyHelmReady     string = "helm-ready"
// 	KeyOutputExists  string = "output-exists"
// 	KeyStsReady      string = "sts-ready"
//
// 	KeyPassword string = "password"
// )
//
// // +kubebuilder:rbac:groups=elasticsearch.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// // +kubebuilder:rbac:groups=elasticsearch.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// // +kubebuilder:rbac:groups=elasticsearch.msvc.kloudlite.io,resources=services/finalizers,verbs=update
//
// func (r *ServiceReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
// 	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &elasticSearch.Service{})
// 	if err != nil {
// 		return ctrl.Result{}, client.IgnoreNotFound(err)
// 	}
//
// 	if req.Object.GetDeletionTimestamp() != nil {
// 		if x := r.finalize(req); !x.ShouldProceed() {
// 			return x.ReconcilerResponse()
// 		}
// 	}
//
// 	req.Logger.Infof("-------------------- NEW RECONCILATION------------------")
//
// 	if req.Object.GetAnnotations()["kloudlite.io/re-check"] == "true" {
// 		ann := req.Object.Annotations
// 		delete(req.Object.Annotations, "kloudlite.io/re-check")
// 		req.Object.SetAnnotations(ann)
//
// 		if err := r.Update(ctx, req.Object); err != nil {
// 			return ctrl.Result{}, err
// 		}
//
// 		req.Object.Status.Checks = nil
// 		if err := r.Status().Update(ctx, req.Object); err != nil {
// 			return ctrl.Result{}, nil
// 		}
// 		return ctrl.Result{RequeueAfter: 0}, nil
// 	}
//
// 	checks := req.Object.Status.Checks
// 	if checks == nil {
// 		checks = map[string]rApi.Check{}
// 	}
//
// 	nChecks := len(checks)
// 	if _, ok := checks[KeyHelmExists]; !ok {
// 		checks[KeyHelmExists] = rApi.Check{}
// 	}
// 	if _, ok := checks[KeyHelmReady]; !ok {
// 		checks[KeyHelmReady] = rApi.Check{}
// 	}
// 	if _, ok := checks[KeyOutputExists]; !ok {
// 		checks[KeyOutputExists] = rApi.Check{}
// 	}
// 	if _, ok := checks[KeyStsReady]; !ok {
// 		checks[KeyStsReady] = rApi.Check{}
// 	}
//
// 	if nChecks != len(checks) {
// 		req.Object.Status.Checks = checks
// 		if err := r.Status().Update(ctx, req.Object); err != nil {
// 			return ctrl.Result{}, nil
// 		}
// 		return ctrl.Result{RequeueAfter: 0}, nil
// 	}
//
// 	if x := req.EnsureLabelsAndAnnotations(); !x.ShouldProceed() {
// 		return x.ReconcilerResponse()
// 	}
//
// 	if x := r.GenerateVars(req); !x.ShouldProceed() {
// 		return x.ReconcilerResponse()
// 	}
//
// 	// checks for helm elastic search
// 	if x := r.reconcileHelm(req); !x.ShouldProceed() {
// 		return x.ReconcilerResponse()
// 	}
//
// 	if x := r.reconcileOutput(req); !x.ShouldProceed() {
// 		return x.ReconcilerResponse()
// 	}
//
// 	if x := r.statefulsetsAndPods(req); !x.ShouldProceed() {
// 		return x.ReconcilerResponse()
// 	}
//
// 	// if x := r.reconcileStatus(req); !x.ShouldProceed() {
// 	// 	return x.ReconcilerResponse()
// 	// }
// 	//
// 	// if x := r.reconcileOperations(req); !x.ShouldProceed() {
// 	// 	return x.ReconcilerResponse()
// 	// }
//
// 	req.Object.Status.IsReady = true
// 	return ctrl.Result{RequeueAfter: 30 * time.Second}, r.Status().Update(ctx, req.Object)
// }
//
// func (r *ServiceReconciler) GenerateVars(req *rApi.Request[*elasticSearch.Service]) stepResult.Result {
// 	obj := req.Object
// 	nItems := obj.Status.GeneratedVars.Len()
//
// 	if !obj.Status.GeneratedVars.Exists(KeyPassword) {
// 		if err := obj.Status.GeneratedVars.Set(KeyPassword, fn.CleanerNanoid(40)); err != nil {
// 			return req.FailWithOpError(err)
// 		}
// 	}
//
// 	if nItems != obj.Status.GeneratedVars.Len() {
// 		if err := r.Status().Update(req.Context(), obj); err != nil {
// 			return req.FailWithStatusError(err)
// 		}
// 		return req.Done().RequeueAfter(1)
// 	}
//
// 	return req.Next()
// }
//
// func (r *ServiceReconciler) reconcileHelm(req *rApi.Request[*elasticSearch.Service]) stepResult.Result {
// 	ctx := req.Context()
// 	obj := req.Object
//
// 	checks := obj.Status.Checks
//
// 	// Check: KeyHelmExists
// 	// if obj.Generation > checks[KeyHelmExists].Generation || time.Since(checks[KeyHelmExists].LastCheckedAt.Time).Seconds() > 30 {
// 	if obj.Generation > checks[KeyHelmExists].Generation && !checks[KeyHelmExists].Status {
// 		check := rApi.Check{Generation: obj.Generation, LastCheckedAt: metav1.Time{Time: time.Now()}}
// 		storageClass, err := obj.Spec.CloudProvider.GetStorageClass(ct.Ext4)
// 		if err != nil {
// 			check.Message = errors.NewEf(err, "could not storage class for fstype=%s", ct.Ext4).Error()
// 			return req.CheckFailed(KeyHelmExists, check)
// 		}
//
// 		var password string
// 		if err := obj.Status.GeneratedVars.Get(KeyPassword, &password); err != nil {
// 			check.Message = err.Error()
// 			return req.CheckFailed(KeyHelmExists, check)
// 		}
//
// 		b, err := templates.Parse(
// 			templates.ElasticSearch, map[string]any{
// 				"object":        obj,
// 				"storage-class": storageClass,
// 				"owner-refs": []metav1.OwnerReference{
// 					fn.AsOwner(obj, true),
// 				},
// 				"password": password,
// 			},
// 		)
// 		if err != nil {
// 			check.Message = err.Error()
// 			return req.CheckFailed(KeyHelmExists, check)
// 		}
//
// 		if err := fn.KubectlApplyExec(ctx, b); err != nil {
// 			check.Message = err.Error()
// 			return req.CheckFailed(KeyHelmExists, check)
// 		}
//
// 		checks[KeyHelmExists] = check
// 		if err := r.Status().Update(ctx, obj); err != nil {
// 			return req.FailWithOpError(err)
// 		}
//
// 		return req.Done().RequeueAfter(0)
// 	}
//
// 	helmElastic, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.HelmElasticType))
// 	if err != nil {
// 		return req.FailWithStatusError(err)
// 	}
//
// 	if c := checks[KeyHelmExists]; !c.Status {
// 		c.Status = true
// 		checks[KeyHelmExists] = c
// 		if err := r.Status().Update(ctx, obj); err != nil {
// 			return req.FailWithOpError(err)
// 		}
// 		return req.Done().RequeueAfter(0)
// 	}
//
// 	// Check: KeyHelmReady
// 	check := rApi.Check{Generation: obj.Generation, LastCheckedAt: metav1.Time{Time: time.Now()}}
// 	cds, err := conditions.FromObject(helmElastic)
//
// 	releaseFailedC := meta.FindStatusCondition(cds, "ReleaseFailed")
// 	if releaseFailedC != nil && releaseFailedC.Status == metav1.ConditionTrue {
// 		if releaseFailedC.Status == metav1.ConditionTrue {
// 			check.Status = false
// 			check.Message = releaseFailedC.Message
// 		}
//
// 		if check != checks[KeyHelmReady] {
// 			checks[KeyHelmReady] = check
// 			if err := r.Status().Update(ctx, obj); err != nil {
// 				return req.FailWithOpError(err)
// 			}
// 			if check.Diff(checks[KeyHelmExists]) {
// 				return req.Done().RequeueAfter(1)
// 			}
// 			return req.Done()
// 		}
// 	}
//
// 	deployedC := meta.FindStatusCondition(cds, "Deployed")
// 	if deployedC == nil {
// 		return req.Done().RequeueAfter(2 * time.Second)
// 	}
// 	if deployedC.Status == metav1.ConditionFalse {
// 		check.Status = false
// 		check.Message = deployedC.Message
// 	}
//
// 	if deployedC.Status == metav1.ConditionTrue {
// 		check.Status = true
// 	}
//
// 	if check != checks[KeyHelmReady] {
// 		checks[KeyHelmReady] = check
// 		if err := r.Status().Update(ctx, obj); err != nil {
// 			return req.FailWithOpError(err)
// 		}
// 	}
//
// 	return req.Next()
// }
//
// func (r *ServiceReconciler) reconcileOutput(req *rApi.Request[*elasticSearch.Service]) stepResult.Result {
// 	obj := req.Object
// 	ctx := req.Context()
// 	checks := obj.Status.Checks
//
// 	if obj.Generation > checks[KeyOutputExists].Generation || time.Since(checks[KeyHelmExists].LastCheckedAt.Time).Seconds() > 30 {
// 		check := rApi.Check{Generation: obj.Generation, LastCheckedAt: metav1.Time{Time: time.Now()}}
// 		var password string
// 		if err := obj.Status.GeneratedVars.Get(KeyPassword, &password); err != nil {
// 			check.Message = err.Error()
// 			return req.CheckFailed(KeyOutputExists, check)
// 			// return req.FailWithOpError(err)
// 		}
//
// 		host := fmt.Sprintf("%s.%s.svc.cluster.local:9200", obj.Name, obj.Namespace)
//
// 		b, err := templates.Parse(
// 			templates.Secret, &corev1.Secret{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "msvc-" + obj.Name,
// 					Namespace: obj.Namespace,
// 					OwnerReferences: []metav1.OwnerReference{
// 						fn.AsOwner(obj, true),
// 					},
// 				},
// 				StringData: map[string]string{
// 					"USERNAME": "elastic",
// 					"PASSWORD": password,
// 					"HOSTS":    host,
// 					"URI":      fmt.Sprintf("http://%s:%s@%s", "elastic", password, host),
// 				},
// 			},
// 		)
//
// 		if err != nil {
// 			check.Message = err.Error()
// 			return req.CheckFailed(KeyOutputExists, check)
// 		}
//
// 		if err := fn.KubectlApplyExec(ctx, b); err != nil {
// 			check.Message = err.Error()
// 			return req.CheckFailed(KeyOutputExists, check)
// 		}
//
// 		check.Status = true
// 		if check != checks[KeyOutputExists] {
// 			checks[KeyOutputExists] = check
// 			if err := r.Status().Update(ctx, obj); err != nil {
// 				return req.FailWithOpError(err)
// 			}
// 		}
// 	}
//
// 	return req.Next()
// }
//
// func (r *ServiceReconciler) statefulsetsAndPods(req *rApi.Request[*elasticSearch.Service]) stepResult.Result {
// 	obj := req.Object
// 	ctx := req.Context()
//
// 	var sts appsv1.StatefulSetList
// 	if err := r.List(
// 		ctx, &sts, &client.ListOptions{
// 			LabelSelector: labels.SelectorFromValidatedSet(
// 				map[string]string{
// 					"kloudlite.io/msvc.name": obj.Name,
// 				},
// 			),
// 			Namespace: obj.Namespace,
// 		},
// 	); err != nil {
// 		return nil
// 	}
//
// 	for i := range sts.Items {
// 		check := rApi.Check{Generation: obj.Generation, LastCheckedAt: metav1.Time{Time: time.Now()}}
// 		check.Status = sts.Items[i].Status.ReadyReplicas == sts.Items[i].Status.Replicas
//
// 		// if !check.Status {
// 		//	return req.CheckFailed(KeyStsReady, check)
// 		// }
// 		//
// 		// if check != obj.Status.Checks[KeyStsReady] {
// 		//	obj.Status.Checks[KeyStsReady] = check
// 		//	if err := r.Status().Update(ctx, obj); err != nil {
// 		//		return req.FailWithOpError(err)
// 		//	}
// 		// }
// 		//
//
// 		if !check.Status {
// 			var podsList corev1.PodList
// 			if err := r.List(
// 				ctx, &podsList, &client.ListOptions{
// 					LabelSelector: labels.SelectorFromValidatedSet(
// 						map[string]string{"kloudlite.io/msvc.name": obj.Name},
// 					),
// 				},
// 			); err != nil {
// 				return req.FailWithOpError(err)
// 			}
//
// 			messages := rApi.GetMessagesFromPods(podsList.Items...)
// 			if len(messages) > 0 {
// 				b, err := json.Marshal(messages)
// 				if err != nil {
// 					check.Message = err.Error()
// 					return req.CheckFailed(KeyStsReady, check)
// 				}
//
// 				check.Message = string(b)
// 				if check != obj.Status.Checks[KeyStsReady] {
// 					obj.Status.Checks[KeyStsReady] = check
// 					if err := r.Status().Update(ctx, obj); err != nil {
// 						return req.FailWithOpError(err)
// 					}
// 				}
// 			}
// 		}
//
// 		if check != obj.Status.Checks[KeyStsReady] {
// 			obj.Status.Checks[KeyStsReady] = check
// 			if err := r.Status().Update(ctx, obj); err != nil {
// 				return req.FailWithOpError(err)
// 			}
// 		}
// 	}
//
// 	return req.Next()
// }
//
// func (r *ServiceReconciler) finalize(req *rApi.Request[*elasticSearch.Service]) stepResult.Result {
// 	return req.Finalize()
// }
//
// func eventFilters() predicate.Predicate {
// 	return predicate.Funcs{
// 		UpdateFunc: func(ev event.UpdateEvent) bool {
// 			return ev.ObjectNew.GetGeneration() != ev.ObjectOld.GetGeneration()
// 		},
// 	}
// }
//
// // SetupWithManager sets up the controllers with the Manager.
// func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
// 	r.Client = mgr.GetClient()
// 	r.Scheme = mgr.GetScheme()
// 	r.logger = logger.WithName(r.Name)
//
// 	builder := ctrl.NewControllerManagedBy(mgr)
// 	builder.For(&elasticSearch.Service{})
// 	builder.Owns(&corev1.Secret{})
//
// 	watchList := []client.Object{
// 		&appsv1.StatefulSet{},
// 		&corev1.Pod{},
// 	}
//
// 	for i := range watchList {
// 		builder.Watches(
// 			&source.Kind{Type: watchList[i]}, handler.EnqueueRequestsFromMapFunc(
// 				func(obj client.Object) []reconcile.Request {
// 					msvcName := obj.GetLabels()[constants.MsvcNameKey]
// 					if msvcName != "" {
// 						return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), msvcName)}}
// 					}
// 					return nil
// 				},
// 			),
// 		)
// 	}
//
// 	builder.WithEventFilter(eventFilters())
//
// 	return builder.Complete(r)
// }
