package mongodbstandalonemsvc

//
//// DatabaseReconciler reconciles a Database object
//type DatabaseReconciler struct {
//	client.Client
//	Scheme *runtime.Scheme
//	lt     metav1.Time
//}
//
//type DatabaseReconReq struct {
//	ctrl.Request
//	stateData map[string]string
//	logger    *zap.SugaredLogger
//	database  *mongodbStandalone.Database
//}
//
//const (
//	DbUser     string = "DB_USER"
//	DbPassword string = "DB_PASSWORD"
//	DbHosts    string = "DB_HOSTS"
//	DbUrl      string = "DB_URL"
//)
//
//const (
//	DbPasswordKey string = "db-password"
//)
//
//// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
//// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
//// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases/finalizers,verbs=update
//
//func (r *DatabaseReconciler) Reconcile(ctx context.Context, orgReq ctrl.Request) (ctrl.Result, error) {
//	req := &DatabaseReconReq{
//		logger:   crds.GetLogger(orgReq.NamespacedName),
//		Request:  orgReq,
//		database: new(mongodbStandalone.Database),
//	}
//
//	req.logger.Infof("Reconciling Database %s", req.database.Name)
//	if err := r.Client.Get(ctx, req.NamespacedName, req.database); err != nil {
//		return ctrl.Result{}, client.IgnoreNotFound(err)
//	}
//
//	if req.database.HasLabels() {
//		req.database.EnsureLabels()
//		if err := r.Update(ctx, req.database); err != nil {
//			return reconcileResult.FailedE(err)
//		}
//	}
//
//	if req.database.GetDeletionTimestamp() != nil {
//		return r.finalize(ctx, req)
//	}
//
//	reconResult, err := r.reconcileStatus(ctx, req)
//	if err != nil {
//		return r.failWithErr(ctx, req, err)
//	}
//	if reconResult != nil {
//		return *reconResult, nil
//	}
//
//	return r.reconcileOperations(ctx, req)
//}
//
//func (r *DatabaseReconciler) finalize(ctx context.Context, req *DatabaseReconReq) (ctrl.Result, error) {
//	return reconcileResult.OK()
//}
//
//func (r *DatabaseReconciler) failWithErr(ctx context.Context, req *DatabaseReconReq, err error) (ctrl.Result, error) {
//	fn.Conditions2.MarkNotReady(&req.database.Status.OpsConditions, err, "ReconFailedWithErr")
//	if err2 := r.Status().Update(ctx, req.database); err2 != nil {
//		return ctrl.Result{}, err2
//	}
//	return reconcileResult.FailedE(err)
//}
//
//func (r *DatabaseReconciler) reconcileStatus(ctx context.Context, req *DatabaseReconReq) (*ctrl.Result, error) {
//	cMap := map[string]metav1.Condition{}
//
//}
//
//func (r *DatabaseReconciler) reconcileStatus2(ctx context.Context, req *DatabaseReconReq) (*ctrl.Result, error) {
//	cMap := map[string]metav1.Condition{}
//
//	// check output exists
//	output := new(corev1.Secret)
//	if err := r.Get(
//		ctx, types.NamespacedName{Namespace: req.database.Namespace, Name: req.database.Name}, output,
//	); err != nil {
//		if !apiErrors.IsNotFound(err) {
//			return nil, err
//		}
//		fn.Conditions2.Build(
//			&conditions, "Output", metav1.Condition{
//				Type:    "Exists",
//				Status:  metav1.ConditionFalse,
//				Reason:  "SecretNotFound",
//				Message: err.Error(),
//			},
//		)
//		output = nil
//	}
//
//	if output != nil {
//		fn.Conditions2.Build(
//			&conditions, "Output", metav1.Condition{
//				Type:    "Exists",
//				Status:  metav1.ConditionTrue,
//				Reason:  "SecretFound",
//				Message: fmt.Sprintf("Secret %s exists", output.Name),
//			},
//		)
//	}
//
//	var usersInfo struct {
//		Users []interface{} `json:"users" bson:"users"`
//	}
//
//	msvcSecret := new(corev1.Secret)
//	if err := r.Client.Get(
//		ctx,
//		types.NamespacedName{
//			Namespace: req.database.Namespace, Name: fmt.Sprintf("msvc-%s", req.database.Spec.ManagedSvcName),
//		},
//		msvcSecret,
//	); err != nil {
//		if !apiErrors.IsNotFound(err) {
//			return nil, err
//		}
//		msvcSecret = nil
//	}
//
//	if msvcSecret != nil {
//		uri := string(msvcSecret.Data["DB_URL"])
//		db, err := connectToDB(ctx, uri, "admin")
//		if err != nil {
//			return nil, err
//		}
//
//		sr := db.RunCommand(
//			ctx, bson.D{
//				{Key: "usersInfo", Value: req.database.Name},
//			},
//		)
//		if err = sr.Decode(&usersInfo); err != nil {
//			return nil, errors.NewEf(err, "could not decode usersInfo")
//		}
//
//		if len(usersInfo.Users) > 0 {
//			fn.Conditions2.Build(&conditions, .MarkReady(
//				fmt.Sprintf(
//					"MongoDB account with (user=%s,db=%s) already exists",
//					req.database.Name,
//					req.database.Name,
//				), "MongoAccountAlreadyExists",
//			)
//			req.condBuilder.MarkReady("Db and User already exists", "DbAndUserAlreadyExists")
//
//			if output == nil {
//				req.condBuilder.MarkReady(
//					"Db and User already exists, but output secret does not exist, could not reconcile further, aborting...",
//					"IrrReconcilable",
//				)
//			}
//		}
//
//	}
//
//	// check if conditions match or not
//	if req.condBuilder.Equal(prevStatus.Conditions) {
//		req.logger.Infof("conditions match, proceeding with reconcile ops")
//		return nil, nil
//	}
//
//	if err := r.Status().Update(ctx, req.database); err != nil {
//		return nil, err
//	}
//	return &ctrl.Result{}, nil
//}
//
//func connectToDB(ctx context.Context, uri, dbName string) (*mongo.Database, error) {
//	cli, err := mongo.NewClient(options.Client().ApplyURI(uri))
//	if err != nil {
//		return nil, errors.NewEf(err, "could not create mongodb client")
//	}
//
//	if err := cli.Connect(ctx); err != nil {
//		return nil, errors.NewEf(err, "could not connect to specified mongodb service")
//	}
//	db := cli.Database(dbName)
//	return db, nil
//}
//
//func (r *DatabaseReconciler) preOps(ctx context.Context, req *DatabaseReconReq) (map[string]string, error) {
//	var opVars map[string]string
//
//	dbPasswd, err := fn.JsonGet[string](req.database.Status.GeneratedVars, DbPasswordKey)
//	if err != nil {
//		m, err := req.database.Status.GeneratedVars.ToMap()
//		if err != nil {
//			return nil, err
//		}
//		m[DbPasswordKey] = fn.CleanerNanoid(40)
//		if err := req.database.Status.GeneratedVars.FillFrom(m); err != nil {
//			return nil, err
//		}
//		return nil, r.Status().Update(ctx, req.database)
//	}
//	opVars["db-password"] = dbPasswd
//
//	opVars["hosts"] = string(msvcSecret.Data["HOSTS"])
//	opVars["root-uri"] = string(msvcSecret.Data["DB_URL"])
//	return opVars, nil
//}
//
//func (r *DatabaseReconciler) reconcileOperations(ctx context.Context, req *DatabaseReconReq) (ctrl.Result, error) {
//	opVars, err := r.preOps(ctx, req)
//	if err != nil {
//		return r.failWithErr(ctx, req, err)
//	}
//	req.logger.Infof("Reconciling Operations")
//	db, err := connectToDB(ctx, opVars["root-uri"], "admin")
//	if err != nil {
//		return r.failWithErr(ctx, req, err)
//	}
//	req.logger.Info("Connected to DB")
//
//	sr := db.RunCommand(
//		ctx, bson.D{
//			{Key: "usersInfo", Value: req.database.Name},
//		},
//	)
//
//	var usersInfo struct {
//		Users []interface{} `json:"users" bson:"users"`
//	}
//
//	if err = sr.Decode(&usersInfo); err != nil {
//		return r.failWithErr(ctx, req, errors.NewEf(err, "could not decode usersInfo"))
//	}
//
//	if len(usersInfo.Users) > 0 {
//		return r.failWithErr(
//			ctx,
//			req,
//			errors.Newf("MongoDB account with (user=%s,db=%s) already exists", req.database.Name, req.database.Name),
//		)
//	}
//
//	var user bson.M
//	if err != nil {
//		return r.failWithErr(ctx, req, errors.NewEf(err, "could not generate password using nanoid"))
//	}
//
//	// ASSERT user does not exist here
//	err = db.RunCommand(
//		ctx, bson.D{
//			{Key: "createUser", Value: req.database.Name},
//			{Key: "pwd", Value: opVars["db-password"]},
//			{
//				Key: "roles", Value: []bson.M{
//					{"role": "dbAdmin", "db": req.database.Name},
//					{"role": "readWrite", "db": req.database.Name},
//				},
//			},
//		},
//	).Decode(&user)
//	if err != nil {
//		return r.failWithErr(ctx, req, errors.NewEf(err, "could not create user"))
//	}
//	req.logger.Info(user)
//
//	outScrt := &corev1.Secret{
//		ObjectMeta: metav1.ObjectMeta{
//			Namespace: req.database.Namespace,
//			Name:      fmt.Sprintf("mres-%s", req.database.Name),
//			OwnerReferences: []metav1.OwnerReference{
//				fn.AsOwner(req.database, true),
//			},
//			Labels: req.database.GetLabels(),
//		},
//		StringData: map[string]string{
//			"PASSWORD": opVars["db-password"],
//			"USERNAME": req.database.Name,
//			"HOSTS":    opVars["hosts"],
//			"URI": fmt.Sprintf(
//				"mongodb://%s:%s@%s/%s",
//				req.database.Name, opVars["db-password"], opVars["hosts"], req.database.Name,
//			),
//		},
//	}
//
//	if err := fn.KubectlApply(ctx, r.Client, outScrt); err != nil {
//		return r.failWithErr(ctx, req, err)
//	}
//
//	return reconcileResult.OK()
//}
//
//// SetupWithManager sets up the controller with the Manager.
//func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
//	r.lt = metav1.Time{Time: time.Now()}
//	return ctrl.NewControllerManagedBy(mgr).
//		For(&mongodbStandalone.Database{}).
//		Complete(r)
//}
