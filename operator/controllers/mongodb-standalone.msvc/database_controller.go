package mongodbstandalonemsvc

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	mongodbStandalone "operators.kloudlite.io/apis/mongodb-standalone.msvc/v1"
	"operators.kloudlite.io/controllers/crds"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	t "operators.kloudlite.io/lib/types"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	lt     metav1.Time
}

type DatabaseReconReq struct {
	t.ReconReq
	ctrl.Request
	logger      *zap.SugaredLogger
	condBuilder fn.StatusConditions
	database    *mongodbStandalone.Database
}

const (
	DbUser     string = "DB_USER"
	DbPassword string = "DB_PASSWORD"
	DbHosts    string = "DB_HOSTS"
	DbUrl      string = "DB_URL"
)

// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *DatabaseReconciler) Reconcile(ctx context.Context, orgReq ctrl.Request) (ctrl.Result, error) {
	req := &DatabaseReconReq{
		logger:   crds.GetLogger(orgReq.NamespacedName),
		Request:  orgReq,
		database: new(mongodbStandalone.Database),
	}

	req.logger.Infof("Reconciling Database %s", req.database.Name)
	if err := r.Client.Get(ctx, req.NamespacedName, req.database); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.condBuilder = fn.Conditions.From(req.database.Status.Conditions)

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
	panic("ASDFASdf")
	req.database.Status.Conditions = []metav1.Condition{}
	return ctrl.Result{}, r.Status().Update(ctx, req.database)
}

func (r *DatabaseReconciler) failWithErr(ctx context.Context, req *DatabaseReconReq, err error) (ctrl.Result, error) {
	req.logger.Error(err)
	req.condBuilder.MarkNotReady(err)
	return ctrl.Result{}, r.updateStatus(ctx, req)
}

func (r *DatabaseReconciler) reconcileStatus(ctx context.Context, req *DatabaseReconReq) (*ctrl.Result, error) {
	msvcSecret := new(corev1.Secret)
	if err := r.Client.Get(
		ctx,
		types.NamespacedName{
			Namespace: req.database.Namespace, Name: fmt.Sprintf("msvc-%s", req.database.Spec.ManagedSvcName),
		},
		msvcSecret,
	); err != nil {
		return nil, err
	}

	req.SetStateData(DbUrl, string(msvcSecret.Data["DB_URL"]))
	req.SetStateData(DbHosts, string(msvcSecret.Data["HOSTS"]))

	return nil, nil
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

func (r *DatabaseReconciler) reconcileOperations(ctx context.Context, req *DatabaseReconReq) (ctrl.Result, error) {
	req.logger.Infof("Reconciling Operations")
	db, err := connectToDB(ctx, req.GetStateData(DbUrl), "admin")
	if err != nil {
		req.logger.Error(err)
		return r.failWithErr(ctx, req, err)
	}
	req.logger.Info("\tConnected to DB")

	sr := db.RunCommand(
		ctx, bson.D{
			{Key: "usersInfo", Value: req.database.Name},
		},
	)

	var usersInfo struct {
		Users []interface{} `json:"users" bson:"users"`
	}

	if err = sr.Decode(&usersInfo); err != nil {
		return r.failWithErr(ctx, req, errors.NewEf(err, "could not decode usersInfo"))
	}

	if len(usersInfo.Users) > 0 {
		req.condBuilder.MarkReady(
			fmt.Sprintf(
				"MongoDB account with (user=%s,db=%s) already exists",
				req.database.Name,
				req.database.Name,
			), "MongoAccountAlreadyExists",
		)
		return reconcileResult.OK()
	}

	var user bson.M
	password := fn.CleanerNanoid(64)
	if err != nil {
		return r.failWithErr(ctx, req, errors.NewEf(err, "could not generate password using nanoid"))
	}

	// ASSERT user does not exist here
	err = db.RunCommand(
		ctx, bson.D{
			{Key: "createUser", Value: req.database.Name},
			{Key: "pwd", Value: password},
			{
				Key: "roles", Value: []bson.M{
					{"role": "dbAdmin", "db": req.database.Name},
					{"role": "readWrite", "db": req.database.Name},
				},
			},
		},
	).Decode(&user)
	if err != nil {
		return r.failWithErr(ctx, req, errors.NewEf(err, "could not create user"))
	}
	req.logger.Info(user)

	resultScrt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: req.database.Namespace,
			Name:      fmt.Sprintf("mres-%s", req.database.Name),
		},
	}

	body := map[string]string{
		DbUser:     req.database.Name,
		DbPassword: password,
		DbHosts:    req.GetStateData("dbHosts"),
		DbUrl: fmt.Sprintf(
			"mongodb://%s:%s@%s/%s?authSource=admin",
			req.database.Name,
			password,
			req.GetStateData("dbHosts"),
			req.database.Name,
		),
	}

	if _, err = controllerutil.CreateOrUpdate(
		ctx, r.Client, resultScrt, func() error {
			resultScrt.Immutable = fn.NewBool(true)
			resultScrt.StringData = body
			if err = controllerutil.SetControllerReference(req.database, resultScrt, r.Scheme); err != nil {
				return err
			}
			return nil
		},
	); err != nil {
		return r.failWithErr(ctx, req, errors.NewEf(err, "could not create secret %s", resultScrt.Name))
	}

	req.condBuilder.SetReady(
		metav1.ConditionTrue,
		"OutputCreated",
		"managed resource output has been created",
	)
	return reconcileResult.OK()
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.lt = metav1.Time{Time: time.Now()}
	return ctrl.NewControllerManagedBy(mgr).
		For(&mongodbStandalone.Database{}).
		Complete(r)
}

func (r *DatabaseReconciler) updateStatus(ctx context.Context, req *DatabaseReconReq) error {
	req.database.Status.Conditions = req.condBuilder.GetAll()
	return r.Status().Update(ctx, req.database)
}
