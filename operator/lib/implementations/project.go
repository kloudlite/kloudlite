package implementations

func (r *ProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("resourceName:", req.Name, "namespace:", req.Namespace)

	project := &crdsv1.Project{}
	if err := r.Get(ctx, req.NamespacedName, project); err != nil {
		if apiErrors.IsNotFound(err) {
			logrus.Debugf("resource (name=%s, namespace=%s) could not be found, must have been deleted after reconcile", req.Name, req.Namespace)
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	// STEP: is object marked for deletion
	isMarkedForDel := project.GetDeletionTimestamp() != nil
	if isMarkedForDel {
		logrus.Debug("RESOURCE is marked for deletion")
		containsFinalizer := controllerutil.ContainsFinalizer(project, projectFinalizer)
		// STEP: 2: add finalizers if not present
		if !containsFinalizer {
			controllerutil.AddFinalizer(project, projectFinalizer)
			logrus.Debugf("ADDING resource finalizer")
			err := r.Update(ctx, project)
			if err != nil {
				return ctrl.Result{}, errors.Newf("could not add finalizer to resource (name=%s, namespace=%s)", req.Name, req.Namespace)
			}
			return ctrl.Result{}, nil
		}

		// STEP: it will container finalizer for sure
		logrus.Debugf("EXECUTING resource finalizer")
		err := finalizeProject(project, logger)
		if err != nil {
			return ctrl.Result{}, err
		}
		controllerutil.RemoveFinalizer(project, projectFinalizer)
		err = r.Update(ctx, project)
		if err != nil {
			return ctrl.Result{}, errors.NewEf(err, "could not remove finalizer from project(%s)", project.Name)
		}
		return ctrl.Result{}, nil
	}

	// FIX: FIX this shit sooner
	if !canReconcile(project) {
		logrus.Infof("returning as job already aborted or is running")
		return ctrl.Result{}, nil
	}

	r.setReady(ctx, project, false)
	r.setInProgress(ctx, project, true)
	err := r.updateStatus(ctx, project)
	if err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 1}, errors.ConditionUpdate(err)
	}

	// DO the operations below
	_, ok := r.CheckIfNSExists(project.Name)
	if !ok {

		// WARN: extract `hotspot` it into env variable
		job, err := r.JobMgr.Create(ctx, "hotspot", &lib.JobVars{
			Name:            "job-create",
			Namespace:       "hotspot",
			ServiceAccount:  "hotspot-cluster-svc-account",
			Image:           "harbor.dev.madhouselabs.io/kloudlite/jobs/project:latest",
			ImagePullPolicy: "Always",
			Args: []string{
				"create",
				"--name", project.Name,
			},
		})

		if err != nil {
			return ctrl.Result{}, errors.NewEf(err, "could not create job from jobMgr")
		}

		status, err := r.JobMgr.Watch(ctx, job.Namespace, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("metadata.name=%s", job.Name),
		})

		if err != nil {
			return ctrl.Result{}, errors.NewEf(err, "watching job failed")
		}

		if !status {
			r.setAborted(ctx, project, true, "job failed")
			err = r.updateStatus(ctx, project)
			if err != nil {
				return ctrl.Result{}, errors.ConditionUpdate(err)
			}
			return ctrl.Result{}, nil
		}

		r.setReady(ctx, project, Bool(status))
		err = r.updateStatus(ctx, project)
		if err != nil {
			return ctrl.Result{}, errors.ConditionUpdate(err)
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func finalizeProject(project *crdsv1.Project, logger logr.Logger) error {
	logger.Info("project finalizer triggerred")
	return nil
}

func canReconcile(project *crdsv1.Project) bool {
	can := true

	failedC := meta.FindStatusCondition(project.Status.Conditions, TypeFailed)
	if failedC == nil {
		fmt.Println("failedC == nil")
		can = true
	}

	if failedC != nil && failedC.Status == metav1.ConditionFalse && failedC.ObservedGeneration == project.Generation {
		fmt.Println("failedC.Status == metav1.ConditionFalse")
		return false
	}

	successC := meta.FindStatusCondition(project.Status.Conditions, TypeSucceeded)
	if successC == nil {
		fmt.Println("successC == nil")
		can = true
	}
	if successC != nil && successC.Status == metav1.ConditionFalse && successC.ObservedGeneration == project.Generation {
		fmt.Println("successC.Status == metav1.ConditionFalse")
		return false
	}

	return can
}

func isAborted(project *crdsv1.Project) bool {
	cond := meta.FindStatusCondition(project.Status.Conditions, TypeFailed)
	if cond == nil {
		return false
	}
	return cond.Status == metav1.ConditionTrue && cond.ObservedGeneration == project.Generation
}

func (r *ProjectReconciler) setReady(ctx context.Context, project *crdsv1.Project, status Bool) {
	meta.RemoveStatusCondition(&project.Status.Conditions, TypeFailed)
	meta.SetStatusCondition(&project.Status.Conditions, metav1.Condition{
		Type:               TypeSucceeded,
		ObservedGeneration: project.GetObjectMeta().GetGeneration(),
		Status:             status.Condition(),
		Reason:             "initialized",
		Message:            "",
	})
}

func (r *ProjectReconciler) setInProgress(ctx context.Context, project *crdsv1.Project, status Bool) {
	meta.RemoveStatusCondition(&project.Status.Conditions, TypeFailed)
	meta.SetStatusCondition(&project.Status.Conditions, metav1.Condition{
		Type:               TypeInProgress,
		ObservedGeneration: project.Generation,
		Status:             status.Condition(),
		Reason:             "queued",
		Message:            "",
	})
}

func (r *ProjectReconciler) setAborted(ctx context.Context, project *crdsv1.Project, status Bool, msg string) {
	meta.RemoveStatusCondition(&project.Status.Conditions, TypeSucceeded)
	meta.SetStatusCondition(&project.Status.Conditions, metav1.Condition{
		Type:               TypeFailed,
		Status:             status.Condition(),
		ObservedGeneration: project.ObjectMeta.GetGeneration(),
		Reason:             "final",
		Message:            msg,
	})
}

func (r *ProjectReconciler) updateStatus(ctx context.Context, project *crdsv1.Project) error {
	err := r.Status().Update(ctx, project)
	if err != nil {
		return errors.Newf("could not update conditions on project (%s) as %w", project.Name, err)
	}
	return nil
}
