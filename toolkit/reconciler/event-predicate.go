package reconciler

import (
	"encoding/json"
	"fmt"
	"reflect"

	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type res struct {
	Enabled bool `json:"enabled,omitempty"`
	// Overrides struct {
	// 	Patches []jsonPatch.PatchOperation `json:"patches,omitempty"`
	// } `json:"overrides,omitempty"`
	Status struct {
		Checks  map[string]Check `json:"checks,omitempty"`
		IsReady *bool            `json:"isReady"`
	} `json:"status"`
}

func getRes(obj client.Object) res {
	b, err := json.Marshal(obj)
	if err != nil {
		return res{}
	}
	var xRes res
	if err := json.Unmarshal(b, &xRes); err != nil {
		return res{}
	}

	return xRes
}

const (
	ReasonGenerationUpdated      string = "GENERATION_UPDATED"
	ReasonDeletionTimestamp      string = "DELETION_TIMESTAMP"
	ReasonLabelsUpdated          string = "LABELS_UPDATED"
	ReasonAnnotationsUpdated     string = "ANNOTATIONS_UPDATED"
	ReasonFinalizersUpdated      string = "FINALIZERS_UPDATED"
	ReasonOwnerReferencesUpdated string = "OWNER_REFERENCES_UPDATED"
	ReasonStatusIsReadyChanged   string = "STATUS_IS_READY_CHANGED"
	ReasonStatusChecksUpdated    string = "CHECKS_UPDATED"
)

func ReconcileFilter(eventRecorder ...record.EventRecorder) predicate.Funcs {
	var recorder record.EventRecorder
	if len(eventRecorder) > 0 {
		recorder = eventRecorder[0]
	}

	fireEvent := func(obj client.Object, reason string, message string) {
		if recorder != nil {
			recorder.Event(obj, "Normal", reason, message)
		}
	}

	return predicate.Funcs{
		UpdateFunc: func(ev event.UpdateEvent) bool {
			oldObj := ev.ObjectOld
			newObj := ev.ObjectNew

			resourceName := oldObj.GetName()

			if newObj.GetGeneration() > oldObj.GetGeneration() {
				fireEvent(newObj, ReasonGenerationUpdated, fmt.Sprintf("generation change from (%d) to (%d)", oldObj.GetGeneration(), newObj.GetGeneration()))
				return true
			}

			if newObj.GetDeletionTimestamp().IsZero() != oldObj.GetDeletionTimestamp().IsZero() {
				fireEvent(newObj, ReasonDeletionTimestamp, "deletion timestamp has been added")
				return true
			}

			if len(oldObj.GetLabels()) != len(newObj.GetLabels()) || !reflect.DeepEqual(oldObj.GetLabels(), newObj.GetLabels()) {
				fireEvent(newObj, ReasonLabelsUpdated, fmt.Sprintf("labels updated from (%+v) to (%+v)", newObj.GetLabels(), oldObj.GetLabels()))
				return true
			}

			oldAnn := oldObj.GetAnnotations()
			newAnn := newObj.GetAnnotations()

			annHasChanged := false
			for k, v := range oldAnn {
				if k != LastAppliedKey {
					if v != newAnn[k] {
						annHasChanged = true
						break
					}
				}
			}

			if len(oldAnn) != len(newAnn) || annHasChanged {
				fireEvent(newObj, ReasonAnnotationsUpdated, fmt.Sprintf("annotations updated from (%+v) to (%+v)", newObj.GetAnnotations(), oldObj.GetAnnotations()))
				return true
			}

			if len(oldObj.GetFinalizers()) != len(newObj.GetFinalizers()) || !reflect.DeepEqual(oldObj.GetFinalizers(), newObj.GetFinalizers()) {
				fireEvent(newObj, ReasonFinalizersUpdated, fmt.Sprintf("finalizers updated from (%+v) to (%+v)", newObj.GetFinalizers(), oldObj.GetFinalizers()))
				return true
			}

			if len(oldObj.GetOwnerReferences()) != len(newObj.GetOwnerReferences()) ||
				!reflect.DeepEqual(oldObj.GetOwnerReferences(), newObj.GetOwnerReferences()) {
				fireEvent(newObj, ReasonOwnerReferencesUpdated, fmt.Sprintf("owner-references updated from (%+v) to (%+v)", newObj.GetOwnerReferences(), oldObj.GetOwnerReferences()))
				return true
			}

			oldRes, newRes := getRes(ev.ObjectOld), getRes(ev.ObjectNew)
			if oldRes.Enabled != newRes.Enabled {
				return true
			}

			if oldRes.Status.IsReady == nil || newRes.Status.IsReady == nil {
				// this is not our object, it is some other k8s resource, just defaulting it to be always watched
				fireEvent(newObj, ReasonStatusIsReadyChanged, fmt.Sprintf("resource isReady changed from (%+v) to (%+v)", newRes.Status.IsReady, oldRes.Status.IsReady))
				return true
			}

			if *oldRes.Status.IsReady != *newRes.Status.IsReady {
				fireEvent(newObj, ReasonStatusIsReadyChanged, fmt.Sprintf("resource isReady changed from (%+v) to (%+v)", newRes.Status.IsReady, oldRes.Status.IsReady))
				return true
			}

			if len(oldRes.Status.Checks) != len(newRes.Status.Checks) {
				fireEvent(newObj, ReasonStatusChecksUpdated, fmt.Sprintf("resource status.checks changed from (%+v) to (%+v)", newRes.Status.Checks, oldRes.Status.Checks))
				return true
			}

			for k, v := range oldRes.Status.Checks {
				if !AreChecksEqual(newRes.Status.Checks[k], v) {
					fireEvent(newObj, ReasonStatusChecksUpdated, fmt.Sprintf("resource status.checks changed from (%+v) to (%+v)", newRes.Status.Checks, oldRes.Status.Checks))
					return true
				}
			}

			_ = resourceName
			return false
		},
	}
}
