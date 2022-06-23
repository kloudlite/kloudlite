ctx := req.Context()
$objectVar$ := req.Object

isReady := true
var cs []metav1.Condition

// -- cut to top
type MsvcOutputRef struct {
	// TODO: (user)
}

func parseMsvcOutput(s *corev1.Secret) *MsvcOutputRef {
	return &MsvcOutputRef{}
}
// -- cut to top

// STEP: 1. check managed service is ready
msvc, err := rApi.Get(
	ctx, r.Client, fn.NN($objectVar$.Namespace, $objectVar$.Spec.ManagedSvcName),
	&$managedSvcPackage$.Service{},
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
	if !msvc.Status.IsReady {
		isReady = false
		msvc = nil
	}
}

// STEP: 2. retrieve managed svc output (usually secret)
if msvc != nil {
	msvcRef, err2 := func () (*MsvcOutputRef, error) {
		msvcOutput, err := rApi.Get(
			ctx, r.Client, fn.NN(msvc.Namespace, fmt.Sprintf("msvc-%s", msvc.Name)),
			&corev1.Secret{},
		)
		if err != nil {
			isReady = false
			cs = append(cs, conditions.New(conditions.ManagedSvcOutputExists, false, conditions.NotFound, err.Error()))
			return nil, err
		}
		cs = append(cs, conditions.New(conditions.ManagedSvcOutputExists, true, conditions.Found))
		outputRef := parseMsvcOutput(msvcOutput)
		rApi.SetLocal(req, "msvc-output-ref", outputRef)
		return outputRef, nil
	}()
	if err2 != nil {
		return req.FailWithStatusError(err2)
	}

	if err2 := func () error {
		// STEP: 3. check reconciler (child components e.g. mongo account, s3 bucket, redis ACL user) exists
		// TODO: (user) use msvcRef values

		return nil
	}(); err2 != nil {
		return req.FailWithStatusError(err2)
	}
}


// STEP: 4. check generated vars
if msvc != nil && !$objectVar$.Status.GeneratedVars.Exists(...) {
	cs = append(cs, conditions.New(conditions.GeneratedVars, false, conditions.NotReconciledYet))
} else {
	cs = append(cs, conditions.New(conditions.GeneratedVars, true, conditions.Found))
}

	// STEP: 5. reconciler output exists?
_, err5 := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, fmt.Sprintf("mres-%s", obj.Name)), &corev1.Secret{})
if err5 != nil {
	cs = append(cs, conditions.New(conditions.ReconcilerOutputExists, false, conditions.NotFound, err.Error()))
} else {
	cs = append(cs, conditions.New(conditions.ReconcilerOutputExists, true, conditions.Found))
}


// STEP: 6. patch conditions
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
