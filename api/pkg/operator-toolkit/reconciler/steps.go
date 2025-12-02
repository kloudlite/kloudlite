package reconciler

import (
	"fmt"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Step[T Resource] struct {
	Name      string
	Title     string
	ShouldRun func(obj T) bool // Optional: If provided, step only runs when this returns true
	OnCreate  func(check *Check[T], obj T) StepResult
	OnDelete  func(check *Check[T], obj T) StepResult
}

func isBeingDeleted(obj client.Object) bool {
	return obj.GetDeletionTimestamp() != nil
}

func ReconcileSteps[T Resource](req *Request[T], steps []Step[T]) (ctrl.Result, error) {
	req.PreReconcile()
	defer req.PostReconcile()

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
		// Build checkList for ALL steps first (don't filter here - filter during execution)
		for i := range steps {
			checkList = append(checkList, CheckDefinition{Name: "delete/" + steps[i].Name, Title: "[Delete] " + steps[i].Title})
		}

		if step := req.EnsureCheckList(checkList); !step.ShouldProceed() {
			return step.ReconcilerResponse()
		}

		for i := len(steps) - 1; i >= 0; i-- {
			// Skip step if ShouldRun condition is not met
			if steps[i].ShouldRun != nil && !steps[i].ShouldRun(req.Object) {
				continue
			}

			checkName := checkList[i].Name

			// Skip creating a new running check if it already passed with the same generation
			if existingCheck, exists := req.Object.GetStatus().Checks[checkName]; exists {
				if existingCheck.State == PassedState && existingCheck.Generation == req.Object.GetGeneration() {
					// Only skip if the check completed recently (within last 60 seconds)
					// This allows the check to run again if resources were deleted externally
					if existingCheck.CompletedAt != nil && time.Since(existingCheck.CompletedAt.Time) < 60*time.Second {
						continue
					}
				}
			}

			check := NewRunningCheck(checkName, req)
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

		// Skip step if ShouldRun condition is not met
		if steps[i].ShouldRun != nil && !steps[i].ShouldRun(req.Object) {
			continue
		}
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

	for i := range steps {
		// Skip step if ShouldRun condition is not met
		if steps[i].ShouldRun != nil && !steps[i].ShouldRun(req.Object) {
			continue
		}

		checkName := checkList[i].Name

		// Skip creating a new running check if it already passed with the same generation
		// This prevents infinite reconciliation loops where the check state toggles between running/passed
		if existingCheck, exists := req.Object.GetStatus().Checks[checkName]; exists {
			if existingCheck.State == PassedState && existingCheck.Generation == req.Object.GetGeneration() {
				// Only skip if the check completed recently (within last 60 seconds)
				// This allows the check to run again if resources were deleted externally
				if existingCheck.CompletedAt != nil && time.Since(existingCheck.CompletedAt.Time) < 60*time.Second {
					continue
				}
			}
		}

		check := NewRunningCheck(checkName, req)
		if result := steps[i].OnCreate(check, req.Object); !result.ShouldProceed() {
			return result.ReconcilerResponse()
		}
	}

	req.Object.GetStatus().IsReady = true
	return ctrl.Result{}, nil
}
