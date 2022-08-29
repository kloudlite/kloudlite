package elasticsearchmsvc

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ct "operators.kloudlite.io/apis/common-types"
	elasticSearch "operators.kloudlite.io/apis/elasticsearch.msvc/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	Name   string
}

func (r *ServiceReconciler) GetName() string {
	return r.Name
}

// const (
// 	HelmElasticExists string = "helm.elasticSearch/Exists"
// 	HelmElasticReady  string = "helm.elasticSearch/Ready"
// )

const (
	KeyVarsGenerated string = "vars-generated"
	KeyHelmExists    string = "helm-exists"
	KeyHelmReady     string = "helm-ready"
	KeyOutputExists  string = "output-exists"

	KeyPassword string = "password"
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
		delete(req.Object.Annotations, "kloudlite.io/re-check")
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

	checks := req.Object.Status.Checks
	if checks == nil {
		checks = map[string]rApi.Check{}
	}
	nChecks := len(checks)
	if _, ok := checks[KeyOutputExists]; !ok {
		checks[KeyOutputExists] = rApi.Check{}
	}
	if _, ok := checks[KeyHelmReady]; !ok {
		checks[KeyHelmReady] = rApi.Check{}
	}
	if _, ok := checks[KeyOutputExists]; !ok {
		checks[KeyOutputExists] = rApi.Check{}
	}

	if nChecks != len(checks) {
		req.Object.Status.Checks = checks
		if err := r.Status().Update(ctx, req.Object); err != nil {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{RequeueAfter: 0}, nil
	}

	if x := req.EnsureLabelsAndAnnotations(); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.GenerateVars(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	// checks for helm elastic search
	if x := r.reconcileHelm(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileOutput(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	// if x := r.reconcileStatus(req); !x.ShouldProceed() {
	// 	return x.ReconcilerResponse()
	// }
	//
	// if x := r.reconcileOperations(req); !x.ShouldProceed() {
	// 	return x.ReconcilerResponse()
	// }

	req.Object.Status.IsReady = true
	return ctrl.Result{}, r.Status().Update(ctx, req.Object)
}

func (r *ServiceReconciler) GenerateVars(req *rApi.Request[*elasticSearch.Service]) stepResult.Result {
	obj := req.Object
	nItems := obj.Status.GeneratedVars.Len()

	if !obj.Status.GeneratedVars.Exists(KeyPassword) {
		if err := obj.Status.GeneratedVars.Set(KeyPassword, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithOpError(err)
		}
	}

	if nItems != obj.Status.GeneratedVars.Len() {
		if err := r.Status().Update(req.Context(), obj); err != nil {
			return req.FailWithStatusError(err)
		}
		return req.Done().RequeueAfter(1)
	}

	return req.Next()
}

func (r *ServiceReconciler) reconcileHelm(req *rApi.Request[*elasticSearch.Service]) stepResult.Result {
	ctx := req.Context()
	obj := req.Object

	checks := obj.Status.Checks

	// Check: KeyHelmExists
	if obj.Generation > checks[KeyHelmExists].Generation || time.Since(checks[KeyHelmExists].LastCheckedAt.Time).Seconds() > 30 {
		check := rApi.Check{Generation: obj.Generation, LastCheckedAt: metav1.Time{Time: time.Now()}}
		storageClass, err := obj.Spec.CloudProvider.GetStorageClass(ct.Ext4)
		if err != nil {
			check.Error = errors.NewEf(err, "could not storage class for fstype=%s", ct.Ext4).Error()
			return req.CheckFailed(KeyHelmExists, check)
		}

		var password string
		if err := obj.Status.GeneratedVars.Get(KeyPassword, &password); err != nil {
			check.Error = err.Error()
			return req.CheckFailed(KeyHelmExists, check)
		}

		b, err := templates.Parse(
			templates.ElasticSearch, map[string]any{
				"object":        obj,
				"storage-class": storageClass,
				"owner-refs": []metav1.OwnerReference{
					fn.AsOwner(obj, true),
				},
				"password": password,
			},
		)
		if err != nil {
			check.Error = err.Error()
			return req.CheckFailed(KeyHelmExists, check)
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			check.Error = err.Error()
			return req.CheckFailed(KeyHelmExists, check)
			// return req.FailWithOpError(err)
		}

		checks[KeyHelmExists] = check
		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}

		return req.Done().RequeueAfter(0)
	}

	helmElastic, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.HelmElasticType))
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if c := checks[KeyHelmExists]; !c.Status {
		c.Status = true
		checks[KeyHelmExists] = c
		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done().RequeueAfter(0)
	}

	// Check: KeyHelmReady
	check := rApi.Check{Generation: obj.Generation, LastCheckedAt: metav1.Time{Time: time.Now()}}

	cds, err := conditions.FromObject(helmElastic)
	deployedC := meta.FindStatusCondition(cds, "Deployed")
	if deployedC == nil {
		return req.Done().RequeueAfter(2 * time.Second)
	}
	if deployedC.Status == metav1.ConditionFalse {
		check.Status = false
		check.Error = deployedC.Message
	}

	if deployedC.Status == metav1.ConditionTrue {
		check.Status = true
	}

	if check != checks[KeyHelmReady] {
		checks[KeyHelmReady] = check
		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
	}

	return req.Next()
}

func (r *ServiceReconciler) reconcileOutput(req *rApi.Request[*elasticSearch.Service]) stepResult.Result {
	obj := req.Object
	ctx := req.Context()
	checks := obj.Status.Checks

	if obj.Generation > checks[KeyOutputExists].Generation || time.Since(checks[KeyHelmExists].LastCheckedAt.Time).Seconds() > 30 {
		check := rApi.Check{Generation: obj.Generation, LastCheckedAt: metav1.Time{Time: time.Now()}}
		var password string
		if err := obj.Status.GeneratedVars.Get(KeyPassword, &password); err != nil {
			check.Error = err.Error()
			return req.CheckFailed(KeyOutputExists, check)
			// return req.FailWithOpError(err)
		}

		host := fmt.Sprintf("%s.%s.svc.cluster.local:9200", obj.Name, obj.Namespace)

		b, err := templates.Parse(
			templates.Secret, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "msvc-" + obj.Name,
					Namespace: obj.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						fn.AsOwner(obj, true),
					},
				},
				StringData: map[string]string{
					"USERNAME": "elastic",
					"PASSWORD": password,
					"HOSTS":    host,
					"URI":      fmt.Sprintf("http://%s:%s@%s", "elastic", password, host),
				},
			},
		)

		if err != nil {
			check.Error = err.Error()
			return req.CheckFailed(KeyOutputExists, check)
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			check.Error = err.Error()
			return req.CheckFailed(KeyOutputExists, check)
		}

		check.Status = true
		if check != checks[KeyOutputExists] {
			checks[KeyOutputExists] = check
			if err := r.Status().Update(ctx, obj); err != nil {
				return req.FailWithOpError(err)
			}
		}
	}

	return req.Next()
}

func (r *ServiceReconciler) finalize(req *rApi.Request[*elasticSearch.Service]) stepResult.Result {
	return req.Finalize()
}

// func (r *ServiceReconciler) reconcileStatus(req *rApi.Request[*elasticSearch.Service]) stepResult.Result {
// 	ctx := req.Context()
// 	obj := req.Object
//
// 	isReady := true
// 	cs := make([]metav1.Condition, 0, 4)
//
// 	// 1. check if helm service is available
// 	helmConditions, err := conditions.FromResource(ctx, r.Client, constants.HelmElasticType, "Helm", fn.NN(obj.Namespace, obj.Name))
// 	if err != nil {
// 		if !apiErrors.IsNotFound(err) {
// 			return req.FailWithStatusError(err)
// 		}
// 		isReady = false
// 		cs = append(cs, conditions.New(HelmElasticExists, false, conditions.NotFound, err.Error()))
// 	}
//
// 	cs = append(cs, conditions.New(HelmElasticExists, true, conditions.Found))
//
// 	if c := meta.FindStatusCondition(helmConditions, "Deployed"); c != nil {
// 		if c.Status == metav1.ConditionFalse {
// 			isReady = false
// 			cs = append(cs, conditions.New(HelmElasticReady, false, c.Status, c.Message))
// 		}
// 	}
//
// 	// cs = append(cs, helmConditions...)
//
// 	// 2. check if reconciler output exists
// 	_, err = rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, fmt.Sprintf("msvc-%s", obj.Name)), &corev1.Secret{})
// 	if err != nil {
// 		if !apiErrors.IsNotFound(err) {
// 			return req.FailWithStatusError(err)
// 		}
// 		isReady = false
// 		cs = append(cs, conditions.New("ReconcilerOutputExists", false, "NotFound", err.Error()))
// 	} else {
// 		cs = append(cs, conditions.New("ReconcilerOutputExists", true, "Found"))
// 	}
//
// 	// 3. generated vars
// 	if obj.Status.GeneratedVars.Exists(SvcRootPasswordKey) {
// 		cs = append(cs, conditions.New("GeneratedVars", true, "Exists"))
// 	} else {
// 		isReady = false
// 		cs = append(cs, conditions.New("GeneratedVars", false, "NotReconciledYet"))
// 	}
//
// 	newConditions, hasUpdated, err := conditions.Patch(obj.Status.Conditions, cs)
// 	if err != nil {
// 		return req.FailWithStatusError(err)
// 	}
// 	if !hasUpdated && isReady == obj.Status.IsReady {
// 		return req.Next()
// 	}
//
// 	obj.Status.IsReady = isReady
// 	obj.Status.Conditions = newConditions
//
// 	if err := r.Status().Update(ctx, obj); err != nil {
// 		return req.FailWithStatusError(err)
// 	}
// 	return req.Done()
// }
//
// func (r *ServiceReconciler) reconcileOperations(req *rApi.Request[*elasticSearch.Service]) stepResult.Result {
// 	ctx := req.Context()
// 	svcObj := req.Object
//
// 	if meta.IsStatusConditionFalse(svcObj.Status.Conditions, "GeneratedVars") {
// 		if err := svcObj.Status.GeneratedVars.Set(SvcRootPasswordKey, fn.CleanerNanoid(40)); err != nil {
// 			return req.FailWithOpError(err)
// 		}
// 		if err := r.Status().Update(ctx, svcObj); err != nil {
// 			return req.FailWithOpError(err)
// 		}
// 		return req.Done()
// 	}
//
// 	storageClass, err := svcObj.Spec.CloudProvider.GetStorageClass(ct.Ext4)
// 	if err != nil {
// 		return req.FailWithOpError(errors.NewEf(err, "could not storage class for fstype=%s", ct.Ext4)).Err(nil)
// 	}
//
// 	b, err := templates.Parse(
// 		templates.ElasticSearch, map[string]any{
// 			"object":        svcObj,
// 			"storage-class": storageClass,
// 			"owner-refs": []metav1.OwnerReference{
// 				fn.AsOwner(svcObj, true),
// 			},
// 		},
// 	)
// 	if err != nil {
// 		return req.FailWithOpError(err)
// 	}
//
// 	if err := fn.KubectlApplyExec(ctx, b); err != nil {
// 		return req.FailWithOpError(err)
// 	}
//
// 	rootPassword, ok := svcObj.Status.GeneratedVars.GetString(SvcRootPasswordKey)
// 	if !ok {
// 		return req.FailWithOpError(
// 			errors.Newf(
// 				"key=%s must have been present in .Status.GeneratedVars",
// 				SvcRootPasswordKey,
// 			),
// 		)
// 	}
//
// 	host := fmt.Sprintf("%s.%s.svc.cluster.local:9200", svcObj.Name, svcObj.Namespace)
// 	b, err = templates.Parse(
// 		templates.Secret, &corev1.Secret{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Name:      fmt.Sprintf("msvc-%s", svcObj.Name),
// 				Namespace: svcObj.Namespace,
// 				OwnerReferences: []metav1.OwnerReference{
// 					fn.AsOwner(svcObj, true),
// 				},
// 			},
// 			StringData: map[string]string{
// 				"USERNAME": "elastic",
// 				"PASSWORD": rootPassword,
// 				"HOSTS":    host,
// 				"URI":      fmt.Sprintf("http://%s:%s@%s", "elastic", rootPassword, host),
// 			},
// 		},
// 	)
// 	if err != nil {
// 		return req.FailWithOpError(err)
// 	}
//
// 	if err := fn.KubectlApplyExec(ctx, b); err != nil {
// 		return req.FailWithOpError(err)
// 	}
//
// 	svcObj.Status.OpsConditions = []metav1.Condition{}
// 	if err := r.Status().Update(ctx, svcObj); err != nil {
// 		return req.FailWithOpError(err)
// 	}
// 	return req.Next()
// }
//

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	return ctrl.NewControllerManagedBy(mgr).
		For(&elasticSearch.Service{}).
		Owns(fn.NewUnstructured(constants.HelmElasticType)).
		// Owns(&appsv1.Deployment{}).
		// Owns(&appsv1.StatefulSet{}).
		// Owns(&corev1.Service{}).
		Complete(r)
}
