  ctx := req.Context()
	$objectVar$ := req.Object

	isReady := true
	var cs []metav1.Condition
	var childC []metav1.Condition

	// STEP: 1. sync conditions from CRs of helm/custom controllers
	helmResource, err := rApi.Get(
		ctx, r.Client, fn.NN($objectVar$.Namespace, $objectVar$.Name), fn.NewUnstructured(constants.HelmMongoDBType),
	)

	if err != nil {
		isReady = false
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New(conditions.HelmResourceExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.HelmResourceExists, true, conditions.Found))

		rConditions, err := conditions.ParseFromResource(helmResource, "Helm")
		if err != nil {
			return req.FailWithStatusError(err)
		}
		childC = append(childC, rConditions...)
		rReady := meta.IsStatusConditionTrue(rConditions, "Deployed")
		if !rReady {
			isReady = false
		}
		cs = append(
			cs, conditions.New(conditions.HelmResourceReady, rReady, conditions.Empty),
		)
	}

	// STEP: 2. sync conditions from deployments/statefulsets
  deploymentRes, err := rApi.Get(ctx, r.Client, fn.NN($objectVar$.Namespace, $objectVar$.Name), &appsv1.$appType${})
	if err != nil {
		isReady = false
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New(conditions.DeploymentExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.DeploymentExists, true, conditions.Found))
		rConditions, err := conditions.ParseFromResource(deploymentRes, "Deployment")
		if err != nil {
			return req.FailWithStatusError(err)
		}
		childC = append(childC, rConditions...)
		rReady := meta.IsStatusConditionTrue(rConditions, "Available")
		if !rReady {
			isReady = false
		}
		cs = append(
			cs, conditions.New(conditions.DeploymentReady, rReady, conditions.Empty),
		)
	}

	// STEP: 3. if vars generated ?
  if !$objectVar$.Status.GeneratedVars.Exists(SvcRootPasswordKey) {
		isReady = false
		cs = append(
			cs, conditions.New(
				conditions.GeneratedVars, false, conditions.NotReconciledYet,
			),
		)
	} else {
		cs = append(cs, conditions.New(conditions.GeneratedVars, true, conditions.Found))
	}

	// STEP: if reconciler output exists
  _, err = rApi.Get(
		ctx, r.Client, fn.NN($objectVar$.Namespace, fmt.Sprintf("msvc-%s", $objectVar$.Name)), &corev1.Secret{},
	)
	if err != nil {
		isReady = false
		cs = append(cs, conditions.New(conditions.ReconcilerOutputExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.ReconcilerOutputExists, true, conditions.Found))
	}

	// STEP: patch aggregated conditions
  nConditionsC, hasUpdatedC, err := conditions.Patch($objectVar$.Status.ChildConditions, childC)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	nConditions, hasSUpdated, err := conditions.Patch($objectVar$.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !hasUpdatedC && !hasSUpdated && isReady == $objectVar$.Status.IsReady {
		return req.Next()
	}

	$objectVar$.Status.IsReady = isReady
	$objectVar$.Status.Conditions = nConditions
	$objectVar$.Status.ChildConditions = nConditionsC
	$objectVar$.Status.OpsConditions = []metav1.Condition{}

	return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, $objectVar$))

