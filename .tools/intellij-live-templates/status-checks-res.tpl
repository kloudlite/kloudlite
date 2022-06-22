	ctx := req.Context()
	$objectVar$ := req.Object

	isReady := true
	var cs []metav1.Condition

	// STEP: check managed service is ready
	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN($objectVar$.Namespace, $objectVar$.Spec.ManagedSvcName),
		&mongodbStandalone.Service{},
	)

	if err != nil {
		isReady = false
		msvc = nil
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New(conditions.ManagedSvcExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.ManagedSvcExists, true, conditions.Found))
		cs = append(cs, conditions.New(conditions.ManagedSvcReady, msvc.Status.IsReady, conditions.Empty))
	}

	// STEP: retrieve managed svc output (usually secret)
	if msvc != nil {
		msvcOutput, err := rApi.Get(
			ctx, r.Client, fn.NN(msvc.Namespace, fmt.Sprintf("msvc-%s", msvc.Name)),
			&corev1.Secret{},
		)
		if err != nil {
			isReady = false
			if !apiErrors.IsNotFound(err) {
				return req.FailWithStatusError(err)
			}
			cs = append(cs, conditions.New(conditions.ManagedSvcOutputExists, false, conditions.NotFound, err.Error()))
		} else {
			cs = append(cs, conditions.New(conditions.ManagedSvcOutputExists, true, conditions.Found))
			rApi.SetLocal(req, MsvcOutputKey, msvcOutput)
		}

    // STEP: check reconciler (child components e.g. mongo account, s3 bucket, redis ACL user) exists
    // TODO: (user)
	}


	// STEP: check generated vars
	if msvc != nil && !$objectVar$.Status.GeneratedVars.Exists() {
		cs = append(cs, conditions.New(conditions.GeneratedVars, false, conditions.NotReconciledYet))
	} else {
		cs = append(cs, conditions.New(conditions.GeneratedVars, true, conditions.Found))
	}

	// STEP: patch conditions
	newConditions, updated, err := conditions.Patch($objectVar$.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !updated && isReady == $objectVar$.Status.IsReady {
		return req.Next()
	}

	$objectVar$.Status.IsReady = isReady
	$objectVar$.Status.Conditions = newConditions
	return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, $objectVar$))
