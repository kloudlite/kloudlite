package crds

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"time"
//
// 	"github.com/goombaio/namegenerator"
// 	appsv1 "k8s.io/api/apps/v1"
// 	corev1 "k8s.io/api/core/v1"
// 	"k8s.io/apimachinery/pkg/api/meta"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
// 	"k8s.io/apimachinery/pkg/labels"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	crdsv1 "operators.kloudlite.io/apis/crds/v1"
// 	"operators.kloudlite.io/env"
// 	"operators.kloudlite.io/lib/conditions"
// 	"operators.kloudlite.io/lib/constants"
// 	fn "operators.kloudlite.io/lib/functions"
// 	"operators.kloudlite.io/lib/logging"
// 	rApi "operators.kloudlite.io/lib/operator"
// 	stepResult "operators.kloudlite.io/lib/operator/step-result"
// 	"operators.kloudlite.io/lib/templates"
// 	ctrl "sigs.k8s.io/controller-runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/controller-runtime/pkg/handler"
// 	"sigs.k8s.io/controller-runtime/pkg/reconcile"
// 	"sigs.k8s.io/controller-runtime/pkg/source"
// 	"sigs.k8s.io/yaml"
// )
//
// // AccountRouterReconciler reconciles a AccountRouter object
// type AccountRouterReconciler struct {
// 	client.Client
// 	Scheme *runtime.Scheme
// 	Name   string
// 	logger logging.Logger
// 	env    *env.Env
// }
//
// func (r *AccountRouterReconciler) GetName() string {
// 	return r.Name
// }
//
// const (
// 	ConfigMapName string = "account-router"
// )
//
// const (
// 	KeyConfigMapReady         string = "ingress-configmap-ready"
// 	KeyIngressControllerReady string = "ingress-controllers-ready"
// 	KeyIngressRoutesReady     string = "ingress-routes-ready"
//
// 	KeyRouterCfg        string = "router-cfg"
// 	KeyConfigMapAsOwner string = "configmap-as-owner"
// )
//
// type AccountRouterConfig struct {
// 	ControllerName string   `json:"controllers-name"`
// 	ExtraDomains   []string `json:"extra-domains"`
// }
//
// // +kubebuilder:rbac:groups=crds.kloudlite.io,resources=accountrouters,verbs=get;list;watch;create;update;patch;delete
// // +kubebuilder:rbac:groups=crds.kloudlite.io,resources=accountrouters/status,verbs=get;update;patch
// // +kubebuilder:rbac:groups=crds.kloudlite.io,resources=accountrouters/finalizers,verbs=update
//
// func (r *AccountRouterReconciler) Reconcile(ctx context.Context, oReq reconcile.Request) (reconcile.Result, error) {
// 	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &crdsv1.AccountRouter{})
// 	if err != nil {
// 		return ctrl.Result{}, client.IgnoreNotFound(err)
// 	}
//
// 	if req.Object.GetDeletionTimestamp() != nil {
// 		if x := r.finalize(req); !x.ShouldProceed() {
// 			return x.ReconcilerResponse()
// 		}
// 		return ctrl.Result{}, nil
// 	}
//
// 	req.Logger.Infof("-------------------- NEW RECONCILATION------------------")
//
// 	if step := req.EnsureChecks(KeyConfigMapReady, KeyIngressControllerReady, KeyIngressRoutesReady); !step.ShouldProceed() {
// 		return step.ReconcilerResponse()
// 	}
//
// 	if step := r.reconcileConfigmap(req); !step.ShouldProceed() {
// 		return step.ReconcilerResponse()
// 	}
//
// 	if step := r.reconcileIngressController(req); !step.ShouldProceed() {
// 		return step.ReconcilerResponse()
// 	}
//
// 	if step := r.reconcileIngressRoutes(req); !step.ShouldProceed() {
// 		return step.ReconcilerResponse()
// 	}
//
// 	req.Object.Status.IsReady = true
// 	return ctrl.Result{RequeueAfter: r.env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
// }
//
// func (r *AccountRouterReconciler) finalize(req *rApi.Request[*crdsv1.AccountRouter]) stepResult.Result {
// 	return req.Finalize()
// }
//
// func (r *AccountRouterReconciler) reconcileConfigmap(req *rApi.Request[*crdsv1.AccountRouter]) stepResult.Result {
// 	obj, ctx, checks := req.Object, req.Context(), req.Object.Status.Checks
//
// 	cfgMap, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, ConfigMapName), &corev1.ConfigMap{})
// 	if err != nil {
// 		cfgMap = nil
// 		req.Logger.Infof("ingress-nginx-config not found, would be creating now")
// 	}
//
// 	if cfgMap == nil {
// 		check := rApi.Check{Generation: obj.Generation}
// 		seed := time.Now().UTC().UnixNano()
// 		nameGenerator := namegenerator.NewNameGenerator(seed)
//
// 		rCfg, err := yaml.Marshal(
// 			AccountRouterConfig{
// 				ControllerName: fmt.Sprintf("ingress-%s", nameGenerator.Generate()),
// 				ExtraDomains:   []string{},
// 			},
// 		)
// 		if err != nil {
// 			return req.CheckFailed(KeyConfigMapReady, check, err.Error())
// 		}
//
// 		b, err := templates.Parse(
// 			templates.CoreV1.ConfigMap, map[string]any{
// 				"name":       ConfigMapName,
// 				"namespace":  obj.Namespace,
// 				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
// 				"data":       map[string]string{"config.yaml": string(rCfg)},
// 			},
// 		)
// 		if err != nil {
// 			check.Message = err.Error()
// 			return req.CheckFailed(KeyConfigMapReady, check).Err(nil)
// 		}
//
// 		if err := fn.KubectlApplyExec(ctx, b); err != nil {
// 			check.Message = err.Error()
// 			return req.CheckFailed(KeyConfigMapReady, check).Err(nil)
// 		}
//
// 		checks[KeyConfigMapReady] = check
// 		if err := r.Status().Update(ctx, obj); err != nil {
// 			return req.FailWithOpError(err)
// 		}
//
// 		return req.Done().RequeueAfter(2 * time.Second)
// 	}
//
// 	check := rApi.Check{Generation: obj.Generation}
// 	check.Status = cfgMap != nil
//
// 	var routerCfg AccountRouterConfig
// 	if err := yaml.Unmarshal([]byte(cfgMap.Data["config.yaml"]), &routerCfg); err != nil {
// 		check.Message = err.Error()
// 		return req.CheckFailed(KeyConfigMapReady, check).Err(nil)
// 	}
//
// 	if check != checks[KeyConfigMapReady] {
// 		checks[KeyConfigMapReady] = check
// 		if check.Status {
// 			if err := obj.Status.Message.Delete(KeyConfigMapReady); err != nil {
// 				return req.FailWithOpError(err)
// 			}
// 		}
// 		if err := r.Status().Update(ctx, obj); err != nil {
// 			return req.FailWithOpError(err)
// 		}
//
// 		return req.Done().RequeueAfter(1 * time.Second)
// 	}
//
// 	rApi.SetLocal(req, KeyRouterCfg, routerCfg)
// 	rApi.SetLocal(req, KeyConfigMapAsOwner, fn.AsOwner(cfgMap, true))
// 	return req.Next()
// }
//
// func (r *AccountRouterReconciler) reconcileIngressController(req *rApi.Request[*crdsv1.AccountRouter]) stepResult.Result {
// 	obj, ctx, checks := req.Object, req.Context(), req.Object.Status.Checks
//
// 	check := rApi.Check{Generation: obj.Generation}
//
// 	routerCfg, ok := rApi.GetLocal[AccountRouterConfig](req, KeyRouterCfg)
// 	if !ok {
// 		check.Message = fmt.Sprintf("key %s not present in locals", KeyRouterCfg)
// 		return req.CheckFailed(KeyIngressControllerReady, check)
// 	}
//
// 	ingressController := func() *unstructured.Unstructured {
// 		ingC := fn.NewUnstructured(constants.HelmIngressNginx)
// 		if err := r.Get(ctx, fn.NN(obj.Namespace, routerCfg.ControllerName), ingC); err != nil {
// 			return nil
// 		}
// 		return ingC
// 	}()
//
// 	if ingressController == nil || obj.Generation > checks[KeyIngressControllerReady].Generation {
// 		cfgAsOwner, ok := rApi.GetLocal[metav1.OwnerReference](req, KeyConfigMapAsOwner)
// 		if !ok {
// 			check.Message = fmt.Sprintf("key %s not present in locals", KeyConfigMapAsOwner)
// 			return req.CheckFailed(KeyConfigMapAsOwner, check)
// 		}
//
// 		b, err := templates.Parse(
// 			templates.HelmIngressNginx, map[string]any{
// 				"obj":              obj,
// 				"controllers-name": routerCfg.ControllerName,
// 				"owner-refs":       []metav1.OwnerReference{cfgAsOwner},
// 				"labels": map[string]string{
// 					constants.AccountRouterNameKey: obj.Name,
// 				},
// 			},
// 		)
// 		if err != nil {
// 			check.Message = err.Error()
// 			return req.CheckFailed(KeyIngressControllerReady, check).Err(nil)
// 		}
//
// 		if err := fn.KubectlApplyExec(ctx, b); err != nil {
// 			check.Message = err.Error()
// 			return req.CheckFailed(KeyIngressControllerReady, check).Err(nil)
// 		}
//
// 		obj.Status.Checks[KeyIngressControllerReady] = check
// 		if err := r.Status().Update(ctx, obj); err != nil {
// 			return req.FailWithOpError(err)
// 		}
//
// 		return req.Done().RequeueAfter(1 * time.Second)
// 	}
//
// 	cds, err := conditions.FromObject(ingressController)
// 	if err != nil {
// 		check.Message = err.Error()
// 		return req.CheckFailed(KeyIngressControllerReady, check)
// 	}
//
// 	condition := meta.FindStatusCondition(cds, "Deployed")
// 	if condition == nil || condition.Status == metav1.ConditionUnknown {
// 		return req.Done().RequeueAfter(2 * time.Second)
// 	}
//
// 	if condition.Status == metav1.ConditionFalse {
// 		check.Message = condition.Message
// 		return req.CheckFailed(KeyIngressControllerReady, check)
// 	}
//
// 	if condition.Status == metav1.ConditionTrue {
// 		check.Status = true
// 	}
//
// 	if check != checks[KeyIngressControllerReady] {
// 		checks[KeyIngressControllerReady] = check
// 		if check.Status {
// 			if err := obj.Status.Message.Delete(KeyIngressControllerReady); err != nil {
// 				return req.FailWithOpError(err)
// 			}
// 		}
// 		if err := r.Status().Update(ctx, obj); err != nil {
// 			return req.FailWithOpError(err)
// 		}
// 		return req.Done().RequeueAfter(1 * time.Second)
// 	}
//
// 	return req.Next()
// }
//
// func (r *AccountRouterReconciler) parseAccountDomains(ctx context.Context, accountRef string) []string {
// 	klAcc := fn.NewUnstructured(constants.KloudliteAccountType)
// 	if err := r.Get(ctx, fn.NN("", accountRef), klAcc); err != nil {
// 		return nil
// 	}
// 	b, err := json.Marshal(klAcc.Object)
// 	if err != nil {
// 		return nil
// 	}
//
// 	var j struct {
// 		Spec struct {
// 			OwnedDomains []string `json:"ownedDomains,omitempty"`
// 		}
// 	}
// 	if err := json.Unmarshal(b, &j); err != nil {
// 		return nil
// 	}
// 	return j.Spec.OwnedDomains
// }
//
// func (r *AccountRouterReconciler) reconcileIngressRoutes(req *rApi.Request[*crdsv1.AccountRouter]) stepResult.Result {
// 	obj, ctx, checks := req.Object, req.Context(), req.Object.Status.Checks
//
// 	check := rApi.Check{Generation: obj.Generation}
// 	accDomains := r.parseAccountDomains(ctx, obj.Spec.AccountRef)
//
// 	routerCfg, ok := rApi.GetLocal[AccountRouterConfig](req, KeyRouterCfg)
// 	if !ok {
// 		check.Message = fmt.Sprintf("key %s is not present in locals", KeyRouterCfg)
// 		return req.CheckFailed(KeyIngressRoutesReady, check)
// 	}
//
// 	domains := append(accDomains, routerCfg.ExtraDomains...)
//
// 	if len(domains) == 0 {
// 		check.Status = true
// 		if check != checks[KeyIngressRoutesReady] {
// 			checks[KeyIngressRoutesReady] = check
// 			if err := r.Status().Update(ctx, obj); err != nil {
// 				return req.FailWithOpError(err)
// 			}
// 		}
// 		return req.Next()
// 	}
//
// 	b, err := templates.Parse(
// 		templates.AccountIngressBridge, map[string]any{
// 			"obj":                  obj,
// 			"controllers-name":     routerCfg.ControllerName,
// 			"global-ingress-class": "ingress-nginx",
// 			"domains":              domains,
// 			"labels": map[string]string{
// 				constants.AccountRouterNameKey: obj.Name,
// 			},
// 			"cluster-issuer":   r.env.ClusterCertIssuer,
// 			"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
// 			"ingress-svc-name": routerCfg.ControllerName + "-controllers",
// 		},
// 	)
// 	if err != nil {
// 		check.Message = err.Error()
// 		return req.CheckFailed(KeyIngressRoutesReady, check).Err(nil)
// 	}
//
// 	if err := fn.KubectlApplyExec(ctx, b); err != nil {
// 		check.Message = err.Error()
// 		return req.CheckFailed(KeyIngressRoutesReady, check).Err(nil)
// 	}
//
// 	check.Status = true
// 	if check != checks[KeyIngressRoutesReady] {
// 		checks[KeyIngressRoutesReady] = check
// 		if check.Status {
// 			if err := obj.Status.Message.Delete(KeyIngressRoutesReady); err != nil {
// 				return req.FailWithOpError(err)
// 			}
// 		}
// 		if err := r.Status().Update(ctx, obj); err != nil {
// 			return req.FailWithOpError(err)
// 		}
// 	}
//
// 	return req.Next()
// }
//
// // SetupWithManager sets up the controllers with the Manager.
// func (r *AccountRouterReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
// 	r.Client = mgr.GetClient()
// 	r.Scheme = mgr.GetScheme()
// 	r.env = envVars
// 	r.logger = logger.WithName(r.Name)
//
// 	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.AccountRouter{})
// 	builder.Owns(&corev1.ConfigMap{})
//
// 	watchList := []client.Object{
// 		fn.NewUnstructured(constants.HelmIngressNginx),
// 		&appsv1.Deployment{},
// 	}
//
// 	for i := range watchList {
// 		builder.Watches(
// 			&source.Kind{Type: watchList[i]}, handler.EnqueueRequestsFromMapFunc(
// 				func(obj client.Object) []reconcile.Request {
// 					routerName := obj.GetLabels()[constants.AccountRouterNameKey]
// 					if routerName == "" {
// 						return nil
// 					}
// 					return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), routerName)}}
// 				},
// 			),
// 		)
// 	}
//
// 	builder.Watches(
// 		&source.Kind{Type: fn.NewUnstructured(constants.KloudliteAccountType)}, handler.EnqueueRequestsFromMapFunc(
// 			func(obj client.Object) []reconcile.Request {
// 				accId := obj.GetLabels()[constants.AccountRef]
// 				if accId == "" {
// 					return nil
// 				}
// 				var accRoutersList crdsv1.AccountRouterList
// 				if err := r.List(
// 					context.TODO(), &accRoutersList, &client.ListOptions{
// 						LabelSelector: labels.SelectorFromValidatedSet(
// 							map[string]string{constants.AccountRef: accId},
// 						),
// 					},
// 				); err != nil {
// 					return nil
// 				}
//
// 				rr := make([]reconcile.Request, 0, len(accRoutersList.Items))
// 				for i := range accRoutersList.Items {
// 					rr = append(
// 						rr,
// 						reconcile.Request{NamespacedName: fn.NN(accRoutersList.Items[i].GetNamespace(), accRoutersList.Items[i].GetName())},
// 					)
// 				}
// 				return rr
// 			},
// 		),
// 	)
//
// 	return builder.Complete(r)
// }
