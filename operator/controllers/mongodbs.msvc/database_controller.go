package mongodbsmsvc

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	crdsv1 "operators.kloudlite.io/api/v1"
	mongodb "operators.kloudlite.io/apis/mongodbs.msvc/v1"
	"operators.kloudlite.io/controllers"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger *zap.SugaredLogger
	lt     metav1.Time
	mdb    *mongodb.Database
	lib.MessageSender
}

const (
	DbUser     string = "DB_USER"
	DbPassword string = "DB_PASSWORD"
	DbHosts    string = "DB_HOSTS"
	DbUrl      string = "DB_URL"
)

// Ref: mongo commands referenced from [https://www.mongodb.com/docs/manual/reference/command/]

type UsersInfo struct {
	Users []interface{} `json:"users" bson:"users"`
}

func (r *DatabaseReconciler) buildConditions(source string, conditions ...metav1.Condition) {
	meta.SetStatusCondition(&r.mdb.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             "False",
		Reason:             "ChecksNotCompleted",
		LastTransitionTime: r.lt,
		Message:            "Not All Checks completed",
	})
	for _, c := range conditions {
		if c.Reason == "" {
			c.Reason = "NotSpecified"
		}
		if !c.LastTransitionTime.IsZero() {
			if c.LastTransitionTime.Time.Sub(r.lt.Time).Seconds() > 0 {
				r.lt = c.LastTransitionTime
			}
		}
		if c.LastTransitionTime.IsZero() {
			c.LastTransitionTime = r.lt
		}
		if source != "" {
			c.Reason = fmt.Sprintf("%s:%s", source, c.Reason)
			c.Type = fmt.Sprintf("%s%s", source, c.Type)
		}
		meta.SetStatusCondition(&r.mdb.Status.Conditions, c)
	}
}

func connectToDB(ctx context.Context, uri, dbName string) (*mongo.Database, error) {
	cli, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, errors.NewEf(err, "could not create mongodb client")
	}

	if err := cli.Connect(ctx); err != nil {
		return nil, errors.NewEf(err, "could not connect to specified mongodb service")
	}
	db := cli.Database(dbName)
	return db, nil
}

func (r *DatabaseReconciler) notifyAndDie(ctx context.Context, err error) (ctrl.Result, error) {
	r.buildConditions("", metav1.Condition{
		Type:    "Ready",
		Status:  "False",
		Reason:  "ErrWhileReconcilation",
		Message: err.Error(),
	})
	return r.notify(ctx)
}

func (r *DatabaseReconciler) notify(ctx context.Context) (ctrl.Result, error) {
	if err := r.Status().Update(ctx, r.mdb); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not update status for (db=%s)", r.mdb.LogRef()))
	}
	return reconcileResult.OK()
}

// +kubebuilder:rbac:groups=mongodbs.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodbs.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodbs.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = controllers.GetLogger(req.NamespacedName)
	logger := r.logger.With("RECONCILE", true)

	var mdb mongodb.Database
	if err := r.Get(ctx, req.NamespacedName, &mdb); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		logger.Errorf("Failed to get mdb resource: %v", err)
		return reconcileResult.FailedE(err)
	}
	logger.Infof("Reconcile started...")
	r.mdb = &mdb

	logger.Infof("MongoDatabase %+v", mdb.Spec.ManagedSvc)
	managedSvc := &crdsv1.ManagedService{}
	if err := r.Get(ctx, types.NamespacedName{Name: mdb.Spec.ManagedSvc, Namespace: mdb.Namespace}, managedSvc); err != nil {
		return r.notifyAndDie(ctx, errors.NewEf(err, "failing to get %s, queing for later", managedSvc.LogRef()))
	}

	// STEP: check if managedsvc is ready
	if ok := meta.IsStatusConditionTrue(managedSvc.Status.Conditions, "Ready"); !ok {
		return r.notifyAndDie(ctx, errors.Newf("%s is not ready", managedSvc.LogRef()))
	}

	msvcSecretName := fmt.Sprintf("msvc-%s", mdb.Spec.ManagedSvc)
	var mSecret corev1.Secret
	if err := r.Get(ctx, types.NamespacedName{Namespace: mdb.Namespace, Name: msvcSecretName}, &mSecret); err != nil {
		return r.notifyAndDie(ctx, errors.NewEf(err, "ManagedSvc secret %s/%s not found, aborting reconcilation", mdb.Namespace, msvcSecretName))
	}

	if mdb.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, &mdb, &mSecret)
	}

	// logger.Debugf("secret data: %+v", mSecret.Data)

	db, err := connectToDB(ctx, string(mSecret.Data["DB_URL"]), "admin")
	if err != nil {
		return r.notifyAndDie(ctx, err)
	}

	sr := db.RunCommand(ctx, bson.D{
		{Key: "usersInfo", Value: mdb.Name},
	})

	var usersInfo UsersInfo
	if err = sr.Decode(&usersInfo); err != nil {
		return r.notifyAndDie(ctx, errors.NewEf(err, "could not decode usersInfo"))
	}

	if len(usersInfo.Users) > 0 {
		r.buildConditions("", metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionTrue,
			Reason:  "MongoAccountAlreadyExists",
			Message: fmt.Sprintf("MongoDB Account with (user=%s, db=%s) already exists", mdb.Name, mdb.Name),
		})
		return r.notify(ctx)
	}

	var user bson.M
	password, err := fn.CleanerNanoid(64)
	if err != nil {
		return r.notifyAndDie(ctx, errors.NewEf(err, "could not generate password using nanoid"))
	}
	// ASSERT user does not exist here
	err = db.RunCommand(ctx, bson.D{
		{Key: "createUser", Value: mdb.Name},
		{Key: "pwd", Value: password},
		{Key: "roles", Value: []bson.M{
			{"role": "dbAdmin", "db": mdb.Name},
			{"role": "readWrite", "db": mdb.Name},
		}},
	}).Decode(&user)
	if err != nil {
		return r.notifyAndDie(ctx, errors.NewEf(err, "could not create user"))
	}
	logger.Info(user)

	resultScrt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: mdb.Namespace,
			Name:      fmt.Sprintf("mres-%s", mdb.Name),
		},
	}

	body := map[string]string{
		DbUser:     mdb.Name,
		DbPassword: password,
		DbHosts:    string(mSecret.Data["HOSTS"]),
		DbUrl:      fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=admin", mdb.Name, password, string(mSecret.Data["HOSTS"]), mdb.Name),
	}

	if _, err = controllerutil.CreateOrUpdate(ctx, r.Client, resultScrt, func() error {
		resultScrt.Immutable = fn.NewBool(true)
		resultScrt.StringData = body
		if err = controllerutil.SetControllerReference(&mdb, resultScrt, r.Scheme); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return r.notifyAndDie(ctx, errors.NewEf(err, "could not create secret %s", resultScrt.Name))
	}

	r.buildConditions("", metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionTrue,
		Reason:  "OutputCreated",
		Message: "Managed Resource output has been created",
	})

	return r.notify(ctx)
}

func (r *DatabaseReconciler) finalize(ctx context.Context, mdb *mongodb.Database, connSecret *corev1.Secret) (ctrl.Result, error) {
	logger := r.logger.With("FINALIZER", "true")
	logger.Debug("finalizing ...")

	if controllerutil.ContainsFinalizer(mdb, finalizers.ManagedResource.String()) {
		logger.Debug("HERE", string(connSecret.Data["DB_URL"]))
		// go to database and delete that user
		db, err := connectToDB(ctx, string(connSecret.Data["DB_URL"]), "admin")
		if err != nil {
			return r.notifyAndDie(ctx, err)
		}

		sr := db.RunCommand(ctx, bson.D{
			{Key: "usersInfo", Value: mdb.Name},
		})

		var usersInfo UsersInfo
		if err = sr.Decode(&usersInfo); err != nil {
			return r.notifyAndDie(ctx, errors.NewEf(err, "could not decode usersInfo"))
		}

		if len(usersInfo.Users) == 1 {
			// then delete the user
			if err = db.RunCommand(ctx, bson.D{
				{Key: "dropUser", Value: mdb.Name},
			}).Err(); err != nil {
				logger.Debug(errors.NewEf(err, "could not drop user"))
				return r.notifyAndDie(ctx, errors.NewEf(err, "could not decode usersInfo"))
			}
		}
		controllerutil.RemoveFinalizer(mdb, finalizers.ManagedResource.String())
		if err := r.Update(ctx, mdb); err != nil {
			return r.notifyAndDie(ctx, errors.NewEf(err, "could not update %s", mdb.String()))
		}
		return reconcileResult.OK()
	}

	if controllerutil.ContainsFinalizer(mdb, finalizers.Foreground.String()) {
		controllerutil.RemoveFinalizer(mdb, finalizers.Foreground.String())
		if err := r.Update(ctx, mdb); err != nil {
			return r.notifyAndDie(ctx, errors.NewEf(err, "could not update %s", mdb.String()))
		}
		return reconcileResult.OK()
	}
	return reconcileResult.OK()
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mongodb.Database{}).
		// Watches(&source.Kind{
		// 	Type: &crdsv1.ManagedService{},
		// }, handler.EnqueueRequestsFromMapFunc(func(c client.Object) []reconcile.Request {
		// })).
		Complete(r)
}
