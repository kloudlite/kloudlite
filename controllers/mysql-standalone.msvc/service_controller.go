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
	req, _ := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &mysqlStandalone.Service{})

	if req == nil {
		return ctrl.Result{}, nil
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

	// 1. if not generated, generate vars and reboot reconciler
	if meta.IsStatusConditionFalse(svcObj.Status.Conditions, "GeneratedVars") {
		if err := svcObj.Status.GeneratedVars.Set(MysqlRootPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithOpError(err)
		}
		if err := svcObj.Status.GeneratedVars.Set(MysqlPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithOpError(err)
		}
		return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, svcObj))
	}
	// 2. appply helm template for this resource type
	b, err := templates.Parse(
		templates.MySqlStandalone, map[string]any{
			"object":        svcObj,
			"storage-class": "do-block-storage",
			"owner-refs": []metav1.OwnerReference{
				fn.AsOwner(svcObj, true),
			},
		},
	)
	if err != nil {
		return req.FailWithOpError(err)
	}

	if _, err := fn.KubectlApplyExec(b); err != nil {
		return req.FailWithOpError(errors.NewEf(err, "applying helm template for MysqlStandalone"))
	}

	// 3. create reconciler output (secret)
	rootPassword, ok := svcObj.Status.GeneratedVars.GetString(MysqlRootPasswordKey)
	if !ok {
		return req.FailWithOpError(err)
	}

	mysqlHost := fmt.Sprintf("%s.%s.%s:%d", svcObj.Name, svcObj.Namespace, "svc.cluster.local", 3306)
	b, err = templates.Parse(
		templates.Secret, corev1.Secret{
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
		return req.FailWithOpError(err)
	}

	if _, err := fn.KubectlApplyExec(b); err != nil {
		return req.FailWithOpError(err)
	}

	return req.Done()
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlStandalone.Service{}).
		Owns(fn.NewUnstructured(constants.HelmMysqlType)).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
