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

// STEP: 3. apply CRs of helm/custom controller
if errP := func() error {
  b1, err := templates.Parse(
    templates.$templateName$, map[string]any{
      "object": $objectVar$,
      // TODO: storage-class
      "storage-class": constants.$storageClass$,
      "owner-refs": []metav1.OwnerReference{
        fn.AsOwner($objectVar$, true),
      },
    },
  )

  if err != nil {
    return err
  }

  // STEP: 4. create output
  // TODO:(user)

  b2, err := templates.Parse(
    templates.Secret, &corev1.Secret{
      ObjectMeta: metav1.ObjectMeta{
        Name:      fmt.Sprintf("msvc-%s", $objectVar$.Name),
        Namespace: $objectVar$.Namespace,
        OwnerReferences: []metav1.OwnerReference{
          fn.AsOwner($objectVar$, true),
        },
      },
      StringData: map[string]string{
        // TODO:(user)
      },
    },
  )
  if err != nil {
    return err
  }

  if _, err := fn.KubectlApplyExec(b1, b2); err != nil {
    return err
  }
  return nil
}(); errP != nil {
  req.FailWithOpError(errP)
}

return req.Done()

