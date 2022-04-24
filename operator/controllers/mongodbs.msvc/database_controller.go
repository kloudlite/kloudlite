package mongodbsmsvc

import (
	"context"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crdsv1 "operators.kloudlite.io/api/v1"
	mongodb "operators.kloudlite.io/apis/mongodbs.msvc/v1"
	"operators.kloudlite.io/controllers"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger *zap.SugaredLogger
	mres   *crdsv1.ManagedResource
}

type UsersInfo struct {
	Users []interface{} `json:"users" bson:"users"`
}

func connectToDB(ctx context.Context, uri, dbName string) (*mongo.Database, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, errors.NewEf(err, "could not create mongodb client")
	}

	if err := client.Connect(ctx); err != nil {
		return nil, errors.NewEf(err, "could not connect to specified mongodb service")
	}
	db := client.Database(dbName)
	return db, nil
}

func (r *DatabaseReconciler) updateManagedResource(ctx context.Context, mdb *mongodb.Database, condition metav1.Condition) error {
	meta.SetStatusCondition(&r.mres.Status.Conditions, condition)
	if err := r.Status().Update(ctx, r.mres); err != nil {
		r.logger.Infof("could not update mres status")
		return err
	}
	return nil
}

func (r *DatabaseReconciler) notifyAndDie(ctx context.Context, err error) (ctrl.Result, error) {
	meta.SetStatusCondition(&r.mres.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  "False",
		Reason:  "ErrUnknown",
		Message: errors.NewEf(err, "failed to create mongodb Database").Error(),
	})

	if err2 := r.Status().Update(ctx, r.mres); err2 != nil {
		r.logger.Infof("could not update mres status")
		return reconcileResult.FailedE(errors.New("could not update managed resource"))
	}
	return reconcileResult.FailedE(err)
}

//+kubebuilder:rbac:groups=mongodbs.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mongodbs.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mongodbs.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = controllers.GetLogger(req.NamespacedName)
	logger := r.logger.With("RECONCILE", true)
	mdb := &mongodb.Database{}
	if err := r.Get(ctx, req.NamespacedName, mdb); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		logger.Errorf("Failed to get mdb resource: %v", err)
		return reconcileResult.FailedE(err)
	}
	logger.Infof("Reconcile started...")

	var mres crdsv1.ManagedResource
	if err := r.Get(ctx, types.NamespacedName{Namespace: mdb.Namespace, Name: mdb.Name}, &mres); err != nil {
		return reconcileResult.FailedE(err)
	}

	r.mres = &mres

	meta.SetStatusCondition(&r.mres.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  "False",
		Reason:  "Initialized",
		Message: "starting to reconcile resource",
	})
	if err := r.Status().Update(ctx, r.mres); err != nil {
		reconcileResult.FailedE(err)
	}

	logger.Infof("MongoDatabase %+v", mdb.Spec.ManagedSvc)
	managedSvc := &crdsv1.ManagedService{}
	if err := r.Get(ctx, types.NamespacedName{Name: mdb.Spec.ManagedSvc, Namespace: mdb.Namespace}, managedSvc); err != nil {
		logger.Info(err.Error())
		logger.Infof("failing to get %s, queing for later", managedSvc.String())
		return reconcileResult.FailedE(err)
	}

	//STEP: check if managedsvc is ready
	if ok := meta.IsStatusConditionTrue(managedSvc.Status.Conditions, "Ready"); !ok {
		return reconcileResult.FailedE(errors.Newf("%s is not ready", managedSvc.String()))
	}

	msvcSecretName := fmt.Sprintf("msvc-%s", mdb.Spec.ManagedSvc)
	var mSecret corev1.Secret
	if err := r.Get(ctx, types.NamespacedName{Namespace: mdb.Namespace, Name: msvcSecretName}, &mSecret); err != nil {
		logger.Errorf("ManagedSvc secret %s/%s not found, aborting reconcilation", mdb.Namespace, msvcSecretName)
		return reconcileResult.Failed()
	}

	if mdb.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, mdb, &mSecret)
	}

	logger.Debugf("secret data: %+v", mSecret.Data)

	client, err := mongo.NewClient(options.Client().ApplyURI(string(mSecret.Data["URI"])))
	if err != nil {
		logger.Infof("could not create mongodb client")
		return reconcileResult.FailedE(err)
	}
	if err = client.Connect(ctx); err != nil {
		logger.Infof("could not connect to specified mongodb service")
		return reconcileResult.FailedE(err)
	}

	password, err := fn.CleanerNanoid(64)
	if err != nil {
		logger.Infof("could not generate password")
		return reconcileResult.FailedE(err)
	}

	db, err := connectToDB(ctx, string(mSecret.Data["URI"]), "admin")
	if err != nil {
		return reconcileResult.FailedE(err)
	}

	sr := db.RunCommand(ctx, bson.D{
		{Key: "usersInfo", Value: mdb.Name},
	})

	var usersInfo UsersInfo
	if err = sr.Decode(&usersInfo); err != nil {
		logger.Debug(errors.NewEf(err, "could not decode usersInfo"))
		return reconcileResult.FailedE(err)
	}

	logger.Debugf("UserInfo.Users: %+v", usersInfo.Users)

	if len(usersInfo.Users) > 0 {
		logger.Infof("MongoDB Account with (user=%s, db=%s) already exists", mdb.Name, mdb.Name)
		return reconcileResult.Failed()
	}

	// ASSERT user does not exist here
	var user bson.M
	err = db.RunCommand(ctx, bson.D{
		{Key: "createUser", Value: mdb.Name},
		{Key: "pwd", Value: password},
		{Key: "roles", Value: []bson.M{
			{"role": "dbAdmin", "db": mdb.Name},
			{"role": "readWrite", "db": mdb.Name},
		}},
	}).Decode(&user)
	if err != nil {
		return r.notifyAndDie(ctx, err)
	}
	logger.Info(user)

	resultScrt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: mdb.Namespace,
			Name:      fmt.Sprintf("mres-%s", mdb.Name),
		},
	}

	body := map[string]string{
		"USERNAME": mdb.Name,
		"PASSWORD": password,
		"URI":      fmt.Sprintf("mongodb://%s:%s@%s/%s", mdb.Name, password, mSecret.Data["HOST"], mdb.Name),
	}

	jsonB, err := json.Marshal(body)
	if err != nil {
		return r.notifyAndDie(ctx, errors.NewEf(err, "could not marshal"))
	}
	body["JSON"] = string(jsonB)

	if _, err = controllerutil.CreateOrUpdate(ctx, r.Client, resultScrt, func() error {
		resultScrt.Immutable = fn.NewBool(true)
		resultScrt.StringData = body
		if err = controllerutil.SetControllerReference(mdb, resultScrt, r.Scheme); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return r.notifyAndDie(ctx, errors.NewEf(err, "could not create result secret"))
	}

	// Updating Conditions for managed resource
	meta.SetStatusCondition(&r.mres.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  "True",
		Reason:  "MongoAccountCreated",
		Message: fmt.Sprintf("mongodb (db=%s, user=%s) created", mdb.Name, mdb.Name),
	})
	if err := r.Status().Update(ctx, r.mres); err != nil {
		reconcileResult.FailedE(err)
	}

	logger.Info("Reconcile Completed")
	return reconcileResult.OK()
}

func (r *DatabaseReconciler) finalize(ctx context.Context, mdb *mongodb.Database, connSecret *corev1.Secret) (ctrl.Result, error) {
	logger := r.logger.With("FINALIZER", "true")
	logger.Debug("finalizing ...")

	if controllerutil.ContainsFinalizer(mdb, finalizers.ManagedResource.String()) {
		// go to database and delete that user
		db, err := connectToDB(ctx, string(connSecret.Data["URI"]), "admin")
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
		Complete(r)
}
