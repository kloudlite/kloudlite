package reconciler

import (
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
	req.PreReconcile()
	defer req.PostReconcile()

	checkList := make([]CheckDefinition, 0, len(steps))
	if isBeingDeleted(req.Object) {
		for i := range steps {
			checkList = append(checkList, CheckDefinition{Name: "delete/" + steps[i].Name, Title: "Delete " + steps[i].Title})
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

		controllerutil.RemoveFinalizer(req.Object, Finalizer)
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

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureCheckList(checkList); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(Finalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	for i := range steps {
		check := NewRunningCheck(checkList[i].Name, req)
		if result := steps[i].OnCreate(check, req.Object); !result.ShouldProceed() {
			return result.ReconcilerResponse()
		}
	}

	req.Object.GetStatus().IsReady = true
	if err := req.statusUpdate(); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
