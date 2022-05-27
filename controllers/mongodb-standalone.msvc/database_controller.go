package mongodbstandalonemsvc

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	mongodbStandalone "operators.kloudlite.io/apis/mongodb-standalone.msvc/v1"
	"operators.kloudlite.io/controllers/crds"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	libMongo "operators.kloudlite.io/lib/mongo"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type DatabaseReconReq struct {
	ctrl.Request

	state map[string]any

	logger   *zap.SugaredLogger
	database *mongodbStandalone.Database
}

const (
	DbUser     string = "DB_USER"
	DbPassword string = "DB_PASSWORD"
	DbHosts    string = "DB_HOSTS"
	DbUrl      string = "DB_URL"
)

const (
	DbPasswordKey string = "db-password"
	RootUri       string = "DB_URL"
)

// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *DatabaseReconciler) Reconcile(ctx context.Context, orgReq ctrl.Request) (ctrl.Result, error) {
	req := &DatabaseReconReq{
		logger:   crds.GetLogger(orgReq.NamespacedName),
		Request:  orgReq,
		state:    map[string]any{},
		database: new(mongodbStandalone.Database),
	}

	req.logger.Debugf("------------------------------------------ NEW RECONCILATION ------------------------")

	if err := r.Client.Get(ctx, req.NamespacedName, req.database); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.database.HasLabels() {
		req.database.EnsureLabels()
		if err := r.Update(ctx, req.database); err != nil {
			return reconcileResult.FailedE(err)
		}
	}

	if req.database.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, req)
	}

	reconResult, err := r.reconcileStatus(ctx, req)
	if err != nil {
		return r.failWithErr(ctx, req, err)
	}
	if reconResult != nil {
		return *reconResult, nil
	}

	return r.reconcileOperations(ctx, req)
}

func (r *DatabaseReconciler) finalize(ctx context.Context, req *DatabaseReconReq) (ctrl.Result, error) {
	return reconcileResult.OK()
}

func (r *DatabaseReconciler) failWithErr(ctx context.Context, req *DatabaseReconReq, err error) (ctrl.Result, error) {
	meta.SetStatusCondition(
		&req.database.Status.OpsConditions, metav1.Condition{
			Type:    "ReconFailedWithErr",
			Status:  metav1.ConditionFalse,
			Reason:  "ErrDuringReconcile",
			Message: err.Error(),
		},
	)
	if err2 := r.Status().Update(ctx, req.database); err2 != nil {
		return ctrl.Result{}, err2
	}
	return reconcileResult.FailedE(err)
}

func (r *DatabaseReconciler) reconcileStatus(ctx context.Context, req *DatabaseReconReq) (*ctrl.Result, error) {
	isReady := true
	var cs []metav1.Condition
	// STEP: check managed service is ready
	msvc := new(mongodbStandalone.Service)
	if err := r.Get(
		ctx,
		types.NamespacedName{Namespace: req.Namespace, Name: req.database.Spec.ManagedSvcName},
		msvc,
	); err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
		isReady = false
		cs = append(
			cs, metav1.Condition{
				Type:    "MsvcExists",
				Status:  "False",
				Reason:  "MsvcNotFound",
				Message: err.Error(),
			},
		)
		msvc = nil
	}

	if msvc != nil {
		isReady = msvc.Status.IsReady
		cs = append(
			cs, metav1.Condition{
				Type:   "MsvcIsReady",
				Status: fn.IfThenElse(msvc.Status.IsReady, metav1.ConditionTrue, metav1.ConditionFalse),
			},
		)
	}

	// STEP: check managed service output is ready
	msvcOutput := new(corev1.Secret)
	if err := r.Get(
		ctx,
		fn.NN(req.Namespace, fmt.Sprintf("msvc-%s", req.database.Spec.ManagedSvcName)),
		msvcOutput,
	); err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
		isReady = false
		cs = append(
			cs, metav1.Condition{
				Type:    "MsvcOutputExists",
				Status:  "False",
				Reason:  "SecretNotFound",
				Message: err.Error(),
			},
		)
		msvcOutput = nil
	}

	if msvcOutput != nil {
		req.state[RootUri] = string(msvcOutput.Data[RootUri])
		req.state[DbHosts] = string(msvcOutput.Data[DbHosts])

		// STEP: mongo user userExists
		mc, err := libMongo.NewClient(req.state[RootUri].(string))
		if err != nil {
			return nil, err
		}
		req.state["MongoCli"] = mc

		userExists, err := mc.UserExists(ctx, req.Name)
		if err != nil {
			return nil, err
		}

		if !userExists {
			cs = append(
				cs, metav1.Condition{
					Type:   "MongoUserExists",
					Status: metav1.ConditionFalse,
					Reason: "UserNotFound",
				},
			)
		}

		isReady = userExists
	}

	// STEP: check generated vars
	req.logger.Debugf(req.database.Status.GeneratedVars.GetString(DbPasswordKey))
	if _, ok := req.database.Status.GeneratedVars.GetString(DbPasswordKey); !ok {
		cs = append(
			cs, metav1.Condition{
				Type:    "GeneratedVars",
				Status:  "False",
				Reason:  "NotGeneratedYet",
				Message: "",
			},
		)
		isReady = false
	}

	// STEP: Output Exists
	outputSecret := new(corev1.Secret)
	if err := r.Get(ctx, fn.NN(req.Namespace, fmt.Sprintf("msvc-%s", req.database.Name)), outputSecret); err != nil {
		isReady = false
	}

	req.logger.Debugf("cs: %+v", cs)

	req.database.Status.IsReady = isReady
	newConditions, updated, err := conditions.Patch(req.database.Status.Conditions, cs)
	if err != nil {
		return nil, err
	}
	if !updated {
		return nil, nil
	}

	req.logger.Infof("status is different, so updating status ...")
	req.database.Status.IsReady = isReady
	req.database.Status.Conditions = newConditions
	if err := r.Status().Update(ctx, req.database); err != nil {
		req.logger.Debugf("err: %v", err)
		return nil, err
	}
	return reconcileResult.OKP()
}

func (r *DatabaseReconciler) reconcileOperations(ctx context.Context, req *DatabaseReconReq) (ctrl.Result, error) {
	req.database.Status.OpsConditions = []metav1.Condition{}

	if meta.IsStatusConditionFalse(req.database.Status.Conditions, "GeneratedVars") {
		if err := req.database.Status.GeneratedVars.Set(DbPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, r.Status().Update(ctx, req.database)
	}

	mc, ok := req.state["MongoCli"].(*libMongo.Client)
	if !ok {
		return r.failWithErr(ctx, req, errors.Newf("mongo client not found"))
	}

	dbPasswd, ok := req.database.Status.GeneratedVars.GetString(DbPasswordKey)
	if !ok {
		return r.failWithErr(ctx, req, errors.Newf("%s not found in gen-vars", DbPasswordKey))
	}
	if err := mc.UpsertUser(ctx, req.database.Name, req.database.Name, dbPasswd); err != nil {
		return r.failWithErr(ctx, req, err)
	}

	hosts := req.state[DbHosts].(string)
	outScrt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: req.database.Namespace,
			Name:      fmt.Sprintf("mres-%s", req.database.Name),
			OwnerReferences: []metav1.OwnerReference{
				fn.AsOwner(req.database, true),
			},
			Labels: req.database.GetLabels(),
		},
		StringData: map[string]string{
			DbPassword: dbPasswd,
			DbUser:     req.database.Name,
			DbHosts:    hosts,
			DbUrl: fmt.Sprintf(
				"mongodb://%s:%s@%s/%s",
				req.database.Name, dbPasswd, hosts, req.database.Name,
			),
		},
	}

	if err := fn.KubectlApply(ctx, r.Client, outScrt); err != nil {
		return r.failWithErr(ctx, req, err)
	}

	return reconcileResult.OK()
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mongodbStandalone.Database{}).
		Complete(r)
}
