ctx := req.Context()
$objectVar$ := req.Object

// STEP: 1. add finalizers if needed
if !controllerutil.ContainsFinalizer($objectVar$, constants.CommonFinalizer) {
	controllerutil.AddFinalizer($objectVar$, constants.CommonFinalizer)
	controllerutil.AddFinalizer($objectVar$, constants.ForegroundFinalizer)

	return rApi.NewStepResult(&ctrl.Result{}, r.Update(ctx, $objectVar$))
}

// STEP: 2. generate vars if needed to
if meta.IsStatusConditionFalse($objectVar$.Status.Conditions, conditions.GeneratedVars.String()) {
	if err := $objectVar$.Status.GeneratedVars.Set($key1$, fn.CleanerNanoid(40)); err != nil {
		return req.FailWithStatusError(err)
	}
	return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, $objectVar$))
}

// STEP: 3. retrieve msvc output, need it in creating reconciler output
msvcOutputRef, ok := rApi.GetLocal[*MsvcOutputRef](req, "msvc-output-ref")
if !ok {
	return req.FailWithOpError(errors.Newf("err=%s key not found in req locals", "msvc-output-ref"))
}

// STEP: 4. create child components like mongo-user, redis-acl etc.
_, err4 := func() (..._t, error) {
	// TODO:(user)
}(); 
if err4 != nil {
	// TODO:(user) might need to reconcile with retry with timeout error
	return req.FailWithOpError(err)
}

// STEP: 5. create reconciler output (eg. secret)
// TODO:(user)

if errt := func () error {
	b, err := templates.Parse(
		templates.Secret, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("mres-%s", $objectVar$.Name),
				Namespace: $objectVar$.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					fn.AsOwner($objectVar$, true),
				},
			},
			StringData: map[string]string{
				// TODO: (user)
			},
		},
	)
	if err != nil {
		return err
	}

	if _, err := fn.KubectlApplyExec(b); err != nil {
		return err
	}
	return nil
}(); errt != nil {
	return req.FailWithOpError(errt)
}

return req.Done()
