package mongodbstandalonemsvc

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	mongodbStandalone "operators.kloudlite.io/apis/mongodb-standalone.msvc/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	libMongo "operators.kloudlite.io/lib/mongo"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	Name   string
}

func (r *DatabaseReconciler) GetName() string {
	return r.Name
}

const (
	MongoUserExists conditions.Type = "MongoUserExists"
)

const (
	DbPasswordKey string = "db-password"
)

type MsvcOutputRef struct {
	Hosts string
	Uri   string
}

func parseMsvcOutput(s *corev1.Secret) *MsvcOutputRef {
	return &MsvcOutputRef{
		Hosts: string(s.Data["DB_HOSTS"]),
		Uri:   string(s.Data["DB_URL"]),
	}
}

// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *DatabaseReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &mongodbStandalone.Database{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if step := r.finalize(req); !step.ShouldProceed() {
			return step.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.Logger.Infof("---------------- database reconciler -- NEW RECONCILATION -----------------")

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconcileStatus(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconcileOperations(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	return ctrl.Result{}, nil
}

func (r *DatabaseReconciler) finalize(req *rApi.Request[*mongodbStandalone.Database]) stepResult.Result {
	return req.Finalize()
}

func (r *DatabaseReconciler) reconcileStatus(req *rApi.Request[*mongodbStandalone.Database]) stepResult.Result {
	ctx := req.Context()
	obj := req.Object

	isReady := true
	var cs []metav1.Condition

	// STEP: 1. check managed service is ready
	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.ManagedSvcName),
		&mongodbStandalone.Service{},
	)

	if err != nil {
		isReady = false
		cs = append(cs, conditions.New(conditions.ManagedSvcExists, false, conditions.NotFound, err.Error()))
		return req.FailWithStatusError(err, cs...).Err(nil)
	}
	cs = append(cs, conditions.New(conditions.ManagedSvcExists, true, conditions.Found))
	cs = append(cs, conditions.New(conditions.ManagedSvcReady, msvc.Status.IsReady, conditions.Empty))
	if !msvc.Status.IsReady {
		return req.FailWithStatusError(errors.Newf("msvc %s is not ready", msvc.Name), cs...).Err(nil)
	}

	// STEP: 2. retrieve managed svc output (usually secret)
	if msvc != nil {
		msvcRef, err2 := func() (*MsvcOutputRef, error) {
			msvcOutput, err := rApi.Get(ctx, r.Client, fn.NN(msvc.Namespace, fmt.Sprintf("msvc-%s", msvc.Name)), &corev1.Secret{})
			if err != nil {
				isReady = false
				cs = append(cs, conditions.New(conditions.ManagedSvcOutputExists, false, conditions.NotFound, err.Error()))
				return nil, err
			}
			cs = append(cs, conditions.New(conditions.ManagedSvcOutputExists, true, conditions.Found))
			outputRef := parseMsvcOutput(msvcOutput)
			rApi.SetLocal(req, "msvc-output-ref", outputRef)
			return outputRef, nil
		}()
		if err2 != nil {
			return req.FailWithStatusError(err2)
		}

		if err2 := func() error {
			// STEP: 3. check reconciler (child components e.g. mongo account, s3 bucket, redis ACL user) exists
			mc, err := libMongo.NewClient(msvcRef.Uri)
			if err != nil {
				return err
			}
			if err := mc.Connect(ctx); err != nil {
				return err
			}
			defer mc.Close()

			userExists, err := mc.UserExists(ctx, obj.Name)
			if err != nil {
				cs = append(cs, conditions.New(MongoUserExists, false, conditions.NotFound, err.Error()))
				return err
			}
			if !userExists {
				cs = append(cs, conditions.New(MongoUserExists, false, conditions.NotFound))
				return nil
			}
			cs = append(cs, conditions.New(MongoUserExists, true, conditions.Found))
			return nil
		}(); err2 != nil {
			return req.FailWithStatusError(err2, cs...)
		}
	}

	// STEP: 4. check generated vars
	if msvc != nil && !obj.Status.GeneratedVars.Exists(DbPasswordKey) {
		cs = append(cs, conditions.New(conditions.GeneratedVars, false, conditions.NotReconciledYet))
	} else {
		cs = append(cs, conditions.New(conditions.GeneratedVars, true, conditions.Found))
	}

	// STEP: 5. patch conditions
	newConditions, updated, err := conditions.Patch(obj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !updated && isReady == obj.Status.IsReady {
		return req.Next()
	}

	obj.Status.IsReady = isReady
	obj.Status.Conditions = newConditions
	if err := r.Status().Update(ctx, obj); err != nil {
		return req.FailWithStatusError(err)
	}
	return req.Done()
}

func (r *DatabaseReconciler) reconcileOperations(req *rApi.Request[*mongodbStandalone.Database]) stepResult.Result {
	ctx := req.Context()
	obj := req.Object

	// STEP: 1. add finalizers if needed
	if !fn.ContainsFinalizers(obj, constants.CommonFinalizer, constants.ForegroundFinalizer) {
		controllerutil.AddFinalizer(obj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(obj, constants.ForegroundFinalizer)

		if err := r.Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}

		return req.Done()
	}

	// STEP: 2. generate vars if needed to
	if meta.IsStatusConditionFalse(obj.Status.Conditions, conditions.GeneratedVars.String()) {
		if err := obj.Status.GeneratedVars.Set(DbPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithStatusError(err)
		}
		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	// STEP: 3. retrieve msvc output, need it in creating reconciler output
	msvcRef, ok := rApi.GetLocal[*MsvcOutputRef](req, "msvc-output-ref")
	if !ok {
		return req.FailWithOpError(errors.Newf("err=%s key not found in req locals", "msvc-output-ref"))
	}

	// STEP: 4. create child components like mongo-user, redis-acl etc.
	dbPasswd, ok := obj.Status.GeneratedVars.GetString(DbPasswordKey)
	if !ok {
		return req.FailWithOpError(errors.Newf("key %s not found in GeneratedVars", DbPasswordKey))
	}
	err4 := func() error {
		mc, err := libMongo.NewClient(msvcRef.Uri)
		if err != nil {
			return err
		}
		if err := mc.Connect(ctx); err != nil {
			return err
		}
		defer mc.Close()

		return mc.UpsertUser(ctx, obj.Name, obj.Name, dbPasswd)
	}()
	if err4 != nil {
		// TODO:(user) might need to reconcile with retry with timeout error
		return req.FailWithOpError(err4)
	}

	// STEP: 5. create reconciler output (eg. secret)
	if errt := func() error {
		b, err := templates.Parse(
			templates.Secret, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("mres-%s", obj.Name),
					Namespace: obj.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						fn.AsOwner(obj, true),
					},
				},
				StringData: map[string]string{
					"DB_PASSWORD": dbPasswd,
					"DB_USER":     obj.Name,
					"DB_HOSTS":    msvcRef.Hosts,
					"DB_NAME":     obj.Name,
					"DB_URL": fmt.Sprintf(
						"mongodb://%s:%s@%s/%s",
						obj.Name, dbPasswd, msvcRef.Hosts, obj.Name,
					),
				},
			},
		)
		if err != nil {
			req.Logger.Errorf(err, "failed parsing template %s", templates.Secret)
			return nil
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			req.Logger.Errorf(err, "kubectl apply failed for template %s", templates.Secret)
			return nil
		}
		return nil
	}(); errt != nil {
		return req.FailWithOpError(errt)
	}

	return req.Next()
}

func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	return ctrl.NewControllerManagedBy(mgr).
		For(&mongodbStandalone.Database{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
