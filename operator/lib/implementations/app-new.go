package implementations

// {
// 	_ = log.FromContext(ctx).WithValues("resourceName:", req.Name, "namespace:", req.Namespace)
// 	logger := GetLogger(req.NamespacedName)
// 	r.logger = logger

// 	app := &crdsv1.App{}
// 	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
// 		if apiErrors.IsNotFound(err) {
// 			return reconcileResult.OK()
// 		}
// 		return reconcileResult.Failed()
// 	}

// 	r.logger.Infof("Status %+v\n", app.Status)

// 	if app.HasJob() {
// 		return reconcileResult.Retry(5)
// 	}

// 	if app.Status.Generation != nil && app.Generation > *app.Status.Generation {
// 		return reconcileResult.Retry(5)
// 	}

// 	if app.Status.Job == nil {
// 		job, err := r.JobMgr.Create(ctx, "hotspot", &lib.JobVars{
// 			Name:            "create-job",
// 			Namespace:       "hotspot",
// 			ServiceAccount:  "hotspot-cluster-svc-account",
// 			Image:           "nginx",
// 			ImagePullPolicy: "Always",
// 			Command:         []string{"/bin/sh"},
// 			Args:            []string{"-c", "sleep 5"},
// 		})

// 		if err != nil {
// 			return reconcileResult.Failed()
// 		}

// 		app.Status.Generation = &app.Generation
// 		app.Status.JobDone = false
// 		app.Status.Job = &crdsv1.ReconJob{
// 			Namespace: job.Namespace,
// 			Name:      job.Name,
// 		}

// 		err = r.Status().Update(ctx, app)
// 		if err != nil {
// 			return reconcileResult.RetryE(5, errors.StatusUpdate(err))
// 		}

// 		return reconcileResult.Retry(5)
// 	} else if !app.Status.JobDone {
// 		hasSucceeded, err := r.JobMgr.HasSucceeded(ctx, "hotpsot", app.Name)
// 		if err != nil {
// 			return reconcileResult.Failed()
// 		}
// 		if hasSucceeded {
// 			app.Status.Generation = &app.Generation
// 			app.Status.Job = nil
// 			app.Status.JobDone = true
// 			e := r.Status().Update(ctx, app)
// 			if e != nil {
// 				return reconcileResult.Failed()
// 			}
// 			return reconcileResult.OK()
// 		}
// 		return reconcileResult.Retry(2)
// 	}

// 	return reconcileResult.OK()
// }
