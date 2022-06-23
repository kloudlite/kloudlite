package mysqlstandalonemsvc

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	mysqlStandalone "operators.kloudlite.io/apis/mysql-standalone.msvc/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	MysqlPasswordKey     = "mysql-password"
	MysqlRootPasswordKey = "mysql-root-password"
)

func (r *ServiceReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &mysqlStandalone.Service{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.Result(), x.Err()
		}
	}

	req.Logger.Info("-------------------- NEW RECONCILATION------------------")

	if x := req.EnsureLabels(); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) finalize(req *rApi.Request[*mysqlStandalone.Service]) rApi.StepResult {
	return req.Finalize()
}

func (r *ServiceReconciler) reconcileStatus(req *rApi.Request[*mysqlStandalone.Service]) rApi.StepResult {
	ctx := req.Context()
	svcObj := req.Object

	isReady := true
	var cs []metav1.Condition
	var childC []metav1.Condition

	// STEP: 1. sync conditions from CRs of helm/custom controllers
	helmResource, err := rApi.Get(
		ctx, r.Client, fn.NN(svcObj.Namespace, svcObj.Name), fn.NewUnstructured(constants.HelmMysqlType),
	)

	if err != nil {
		isReady = false
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New(conditions.HelmResourceExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.HelmResourceExists, true, conditions.Found))

		rConditions, err := conditions.ParseFromResource(helmResource, "Helm")
		if err != nil {
			return req.FailWithStatusError(err)
		}
		childC = append(childC, rConditions...)
		rReady := meta.IsStatusConditionTrue(rConditions, "HelmDeployed")
		if !rReady {
			isReady = false
		}
		cs = append(
			cs, conditions.New(conditions.HelmResourceReady, rReady, conditions.Empty),
		)
	}

	// STEP: 2. sync conditions from deployments/statefulsets
	stsRes, err := rApi.Get(ctx, r.Client, fn.NN(svcObj.Namespace, svcObj.Name), &appsv1.StatefulSet{})
	if err != nil {
		isReady = false
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New(conditions.StsExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.StsExists, true, conditions.Found))
		rConditions, err := conditions.ParseFromResource(stsRes, "Sts")
		if err != nil {
			return req.FailWithStatusError(err)
		}
		childC = append(childC, rConditions...)

		if stsRes.Status.Replicas != stsRes.Status.ReadyReplicas {
			isReady = false
			cs = append(cs, conditions.New(conditions.StsReady, false, conditions.Empty))
		} else {
			cs = append(cs, conditions.New(conditions.StsReady, true, conditions.Empty))
		}
	}

	// STEP: 3. if vars generated ?
	if !svcObj.Status.GeneratedVars.Exists(MysqlRootPasswordKey, MysqlPasswordKey) {
		isReady = false
		cs = append(
			cs, conditions.New(
				conditions.GeneratedVars, false, conditions.NotReconciledYet,
			),
		)
	} else {
		cs = append(cs, conditions.New(conditions.GeneratedVars, true, conditions.Found))
	}

	// STEP: 4. if reconciler output exists
	_, err = rApi.Get(
		ctx, r.Client, fn.NN(svcObj.Namespace, fmt.Sprintf("msvc-%s", svcObj.Name)), &corev1.Secret{},
	)
	if err != nil {
		isReady = false
		cs = append(cs, conditions.New(conditions.ReconcilerOutputExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.ReconcilerOutputExists, true, conditions.Found))
	}

	// STEP: 5. patch aggregated conditions
	nConditionsC, hasUpdatedC, err := conditions.Patch(svcObj.Status.ChildConditions, childC)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	nConditions, hasSUpdated, err := conditions.Patch(svcObj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !hasUpdatedC && !hasSUpdated && isReady == svcObj.Status.IsReady {
		return req.Next()
	}

	svcObj.Status.IsReady = isReady
	svcObj.Status.Conditions = nConditions
	svcObj.Status.ChildConditions = nConditionsC
	svcObj.Status.OpsConditions = []metav1.Condition{}

	return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, svcObj))
}

func (r *ServiceReconciler) reconcileStatus2(req *rApi.Request[*mysqlStandalone.Service]) rApi.StepResult {
	svcObj := req.Object
	ctx := req.Context()

	isReady := true
	var cs []metav1.Condition

	// helm resource exists, and fetch conditions from it
	helmConditions, err := conditions.FromResource(
		ctx, r.Client, constants.HelmMysqlType, "Helm", fn.NN(svcObj.Namespace, svcObj.Name),
	)
	if err != nil {
		isReady = false
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New("HelmResourceExists", false, "NotFound", err.Error()))
	}
	cs = append(cs, helmConditions...)

	// helm (owned) statefulset exists
	stsConditions, err := conditions.FromResource(
		ctx, r.Client, constants.StatefulsetType, "Sts", fn.NN(svcObj.Namespace, svcObj.Name),
	)
	if err != nil {
		isReady = false
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New("STSExists", false, "NotFound", err.Error()))
	}
	cs = append(cs, stsConditions...)

	// reconciler output exists (secret)
	_, err = rApi.Get(ctx, r.Client, fn.NN(svcObj.Namespace, svcObj.Name), &corev1.Secret{})
	if err != nil {
		isReady = false
		cs = append(cs, conditions.New("ReconcilerOutputExists", false, "NotFound", err.Error()))
	} else {
		cs = append(cs, conditions.New("ReconcilerOutputExists", true, "Found"))
	}

	// generated vars like root password
	if svcObj.Status.GeneratedVars.Exists(MysqlRootPasswordKey, MysqlPasswordKey) {
		cs = append(cs, conditions.New("GeneratedVars", true, "Exists"))
	} else {
		isReady = false
		cs = append(cs, conditions.New("GeneratedVars", false, "NotReconciledYet"))
	}

	nConditions, hasUpdated, err := conditions.Patch(svcObj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !hasUpdated && isReady == svcObj.Status.IsReady {
		return req.Next()
	}

	svcObj.Status.IsReady = isReady
	svcObj.Status.Conditions = nConditions
	return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, svcObj))
}

func (r *ServiceReconciler) reconcileOperations(req *rApi.Request[*mysqlStandalone.Service]) rApi.StepResult {
	ctx := req.Context()
	svcObj := req.Object

	// STEP: 1. add finalizers if needed
	if !controllerutil.ContainsFinalizer(svcObj, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(svcObj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(svcObj, constants.ForegroundFinalizer)

		return rApi.NewStepResult(&ctrl.Result{}, r.Update(ctx, svcObj))
	}

	// STEP: 2. generate vars if needed to
	if meta.IsStatusConditionFalse(svcObj.Status.Conditions, conditions.GeneratedVars.String()) {
		if err := svcObj.Status.GeneratedVars.Set(MysqlRootPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithStatusError(err)
		}
		if err := svcObj.Status.GeneratedVars.Set(MysqlPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithStatusError(err)
		}
		return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, svcObj))
	}

	// STEP: 3. apply CRs of helm/custom controller
	if errP := func() error {
		b1, err := templates.Parse(
			templates.MySqlStandalone, map[string]any{
				"object": svcObj,
				// TODO: storage-class
				"storage-class": constants.DoBlockStorage,
				"owner-refs": []metav1.OwnerReference{
					fn.AsOwner(svcObj, true),
				},
			},
		)

		if err != nil {
			return err
		}

		// STEP: 4. create output
		rootPassword, ok := svcObj.Status.GeneratedVars.GetString(MysqlRootPasswordKey)
		if !ok {
			return errors.Newf("key=%s is not present in .Status.GeneratedVars", MysqlRootPasswordKey)
		}

		mysqlHost := fmt.Sprintf("%s.%s.%s:%d", svcObj.Name, svcObj.Namespace, "svc.cluster.local", 3306)
		b2, err := templates.Parse(
			templates.Secret, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("msvc-%s", svcObj.Name),
					Namespace: svcObj.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						fn.AsOwner(svcObj, true),
					},
				},
				StringData: map[string]string{
					"ROOT_PASSWORD": rootPassword,
					"HOSTS":         mysqlHost,
					"DSN":           fmt.Sprintf("%s:%s@tcp(%s)/%s", "root", rootPassword, mysqlHost, "mysql"),
					"URI":           fmt.Sprintf("mysqlx://%s:%s@%s/%s", "root", rootPassword, mysqlHost, "mysql"),
				},
			},
		)
		if err != nil {
			return err
		}

		if _, err := fn.KubectlApplyExec(b1, b2); err != nil {
			return err
		}
		return nil
	}(); errP != nil {
		req.FailWithOpError(errP)
	}

	return req.Done()
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {

	builder := ctrl.NewControllerManagedBy(mgr).For(&mysqlStandalone.Service{})

	builder.Owns(fn.NewUnstructured(constants.HelmMysqlType))
	builder.Owns(&corev1.Secret{})

	refWatchList := []client.Object{
		&appsv1.StatefulSet{},
		&corev1.Pod{},
	}

	for _, item := range refWatchList {
		builder.Watches(
			&source.Kind{Type: item}, handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					value, ok := obj.GetLabels()[fmt.Sprintf("%s/ref", mysqlStandalone.GroupVersion.Group)]
					if !ok {
						return nil
					}
					return []reconcile.Request{
						{NamespacedName: fn.NN(obj.GetNamespace(), value)},
					}
				},
			),
		)
	}

	return builder.Complete(r)
}
