package mres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mresv1 "operators.kloudlite.io/apis/mres/v1"
	controllers "operators.kloudlite.io/controllers"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
)

// MongoDatabaseReconciler reconciles a MongoDatabase object
type MongoDatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger *zap.SugaredLogger
}

//INFO: mongo commands referenced from [https://www.mongodb.com/docs/manual/reference/command/]
type UsersInfo struct {
	Users []interface{} `json:"users" bson:"users"`
}

//+kubebuilder:rbac:groups=mres.kloudlite.io,resources=mongodatabases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mres.kloudlite.io,resources=mongodatabases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mres.kloudlite.io,resources=mongodatabases/finalizers,verbs=update

func (r *MongoDatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	r.logger = controllers.GetLogger(req.NamespacedName)
	logger := r.logger.With("RECONCILE", true)

	// get mdb resource
	mdb := &mresv1.MongoDatabase{}
	if err := r.Get(ctx, req.NamespacedName, mdb); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		logger.Errorf("Failed to get mdb resource: %v", err)
		return reconcileResult.FailedE(err)
	}
	logger.Infof("Reconcile started...")

	//TODO: check if mongodatabase is not owned by .Spec.ManagedSvc.Name, then make it so, otherwise peace

	secret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Name: mdb.Spec.ManagedSvc.SecretRef.Name, Namespace: req.Namespace}, secret); err != nil {
		return reconcileResult.FailedE(err)
	}

	if mdb.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, mdb, secret)
	}

	logger.Debugf("secret data: %+v", secret.Data)

	client, err := mongo.NewClient(options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s.%s", "root", string(secret.Data["mongodb-root-password"]), mdb.Spec.ManagedSvc.Service, req.Namespace)))
	if err != nil {
		logger.Infof("could not create mongodb client")
		return reconcileResult.FailedE(err)
	}

	if err = client.Connect(ctx); err != nil {
		logger.Infof("could not connect to specified mongodb service")
		return reconcileResult.FailedE(err)
	}

	// source: https://www.mongodb.com/docs/manual/reference/command/createUser/
	password, err := fn.CleanerNanoid(64)
	if err != nil {
		logger.Infof("could not generate password")
		return reconcileResult.FailedE(err)
	}

	db, err := mdb.ConnectToDB(ctx, "root", string(secret.Data["mongodb-root-password"]), "admin")
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

	if len(usersInfo.Users) == 0 {
		// means user does not exist
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
			e := errors.NewEf(err, "could not create user")
			logger.Info(e)
			return reconcileResult.FailedE(e)
		}
		logger.Info(user)

		resultScrt := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: mdb.Namespace,
				Name:      fmt.Sprintf("%s-%s", mdb.Name, strings.ToLower(mdb.Kind)),
			},
		}

		body := map[string]string{
			"USERNAME": mdb.Name,
			"PASSWORD": password,
			"URI":      fmt.Sprintf("mongodb://%s:%s@%s.%s.svc.cluster.local", mdb.Name, password, mdb.Spec.ManagedSvc.Service, mdb.Namespace),
		}

		jsonB, err := json.Marshal(body)
		if err != nil {
			return reconcileResult.FailedE(errors.NewEf(err, "could not marshal secret body"))
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
			return reconcileResult.FailedE(err)
		}

		mdb.Status.SecretRef.Name = resultScrt.Name
		meta.SetStatusCondition(&mdb.Status.Conditions, metav1.Condition{
			Type:    "Ready",
			Status:  "True",
			Message: "Mongo User Account created",
			Reason:  "MongoAccountCreated",
		})
		if err := r.Status().Update(ctx, mdb); err != nil {
			return reconcileResult.FailedE(err)
		}
		logger.Info("mdb status updated")
		return reconcileResult.OK()
	}

	logger.Infof("Reconcile complete...")
	return ctrl.Result{}, nil
}

func (r *MongoDatabaseReconciler) finalize(ctx context.Context, mdb *mresv1.MongoDatabase, connSecret *corev1.Secret) (ctrl.Result, error) {
	logger := r.logger.With("FINALIZER", "true")
	logger.Debug("finalizing ...")

	if controllerutil.ContainsFinalizer(mdb, finalizers.ManagedResource.String()) {
		// go to database and delete that user
		db, err := mdb.ConnectToDB(ctx, "root", string(connSecret.Data["mongodb-root-password"]), "admin")
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

		if len(usersInfo.Users) == 1 {
			// then delete the user
			if err = db.RunCommand(ctx, bson.D{
				{Key: "dropUser", Value: mdb.Name},
			}).Err(); err != nil {
				logger.Debug(errors.NewEf(err, "could not drop user"))
				return reconcileResult.FailedE(err)
			}
		}
		controllerutil.RemoveFinalizer(mdb, finalizers.ManagedResource.String())
		if err := r.Update(ctx, mdb); err != nil {
			return reconcileResult.FailedE(err)
		}
		return reconcileResult.OK()
	}

	if controllerutil.ContainsFinalizer(mdb, finalizers.Foreground.String()) {
		controllerutil.RemoveFinalizer(mdb, finalizers.Foreground.String())
		if err := r.Update(ctx, mdb); err != nil {
			return reconcileResult.FailedE(err)
		}
		return reconcileResult.OK()
	}
	return reconcileResult.OK()
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongoDatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mresv1.MongoDatabase{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
