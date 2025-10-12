package reconciler

import (
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Step[T Resource] struct {
	Name     string
	Title    string
	OnCreate func(check *Check[T], obj T) StepResult
	OnDelete func(check *Check[T], obj T) StepResult
}

func isBeingDeleted(obj client.Object) bool {
	return obj.GetDeletionTimestamp() != nil
}

func ReconcileSteps[T Resource](req *Request[T], steps []Step[T]) (ctrl.Result, error) {
	// Validate steps
	for i, step := range steps {
		if step.Name == "" {
			return ctrl.Result{}, fmt.Errorf("step %d has empty name", i)
		}
		if step.OnCreate == nil {
			return ctrl.Result{}, fmt.Errorf("step %s has nil OnCreate function", step.Name)
		}
	}

	checkList := make([]CheckDefinition, 0, len(steps))
	if isBeingDeleted(req.Object) {
		for i := range steps {
			checkList = append(checkList, CheckDefinition{Name: "delete/" + steps[i].Name, Title: "[Delete] " + steps[i].Title})
		}

		if step := req.EnsureCheckList(checkList); !step.ShouldProceed() {
			return step.ReconcilerResponse()
		}

		for i := len(steps) - 1; i >= 0; i-- {
			check := NewRunningCheck(checkList[i].Name, req)
			if steps[i].OnDelete == nil {
				if sr := check.Passed(); !sr.ShouldProceed() {
					return sr.ReconcilerResponse()
				}
				continue
			}

			if result := steps[i].OnDelete(check, req.Object); !result.ShouldProceed() {
				return result.ReconcilerResponse()
			}
		}

		// Remove all finalizers that were added
		for _, f := range req.Object.GetFinalizers() {
			if f == Finalizer || f == GenericFinalizer {
				controllerutil.RemoveFinalizer(req.Object, f)
			}
		}
		if err := req.client.Update(req.Context(), req.Object); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	for i := range steps {
		checkList = append(checkList, CheckDefinition{Name: "create/" + steps[i].Name, Title: "[Create] " + steps[i].Title})
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureCheckList(checkList); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(Finalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	forcedReconcile := req.Object.GetAnnotations()[AnnotationForceReconcileKey] == AnnotationForceReconcileValue

	for i := range steps {
		checkName := checkList[i].Name

		// Skip if already passed in current generation (resumability)
		if !forcedReconcile {
			if existingCheck, exists := req.Object.GetStatus().Checks[checkName]; exists {
				if existingCheck.State == PassedState && existingCheck.Generation == req.Object.GetGeneration() {
					req.internalLogger.Info("⏭️  skipping passed check", "check.name", checkName)
					continue
				}
			}
		}

		check := NewRunningCheck(checkName, req)
		if result := steps[i].OnCreate(check, req.Object); !result.ShouldProceed() {
			return result.ReconcilerResponse()
		}
	}

	if forcedReconcile {
		ann := req.Object.GetAnnotations()
		delete(ann, AnnotationForceReconcileKey)
		req.Object.SetAnnotations(ann)
	}

	req.Object.GetStatus().IsReady = true
	if err := req.statusUpdate(); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
