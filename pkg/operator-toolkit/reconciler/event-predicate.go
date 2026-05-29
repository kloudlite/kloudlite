package reconciler

import (
	"encoding/json"
	"fmt"
	"os"
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
		Checks  map[string]CheckResult `json:"checks,omitempty"`
		IsReady *bool                  `json:"isReady,omitempty"`
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

// filterAnnotations returns a new map containing only annotations that are not in the exclusion list
// This prevents internal operator annotations from triggering unnecessary reconciliations
func filterAnnotations(annotations map[string]string) map[string]string {
	if annotations == nil {
		return nil
	}

	// Excluded annotations that change frequently during reconciliation
	excludedKeys := map[string]bool{
		LastAppliedKey:                      true,
		"deployment.kubernetes.io/revision": true,
		AnnotationResourceReady:             true,
		AnnotationResourceChecks:            true,
	}

	filtered := make(map[string]string)
	for k, v := range annotations {
		if !excludedKeys[k] {
			filtered[k] = v
		}
	}

	return filtered
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
		if os.Getenv("KLOUDLITE_CONTROLLER_DEBUG_EVENTS") == "true" && recorder != nil {
			recorder.Event(obj, "Normal", reason, message)
		}
	}

	return predicate.Funcs{
		UpdateFunc: func(ev event.UpdateEvent) bool {
			oldObj := ev.ObjectOld
			newObj := ev.ObjectNew

			if newObj.GetGeneration() > oldObj.GetGeneration() {
				fireEvent(newObj, ReasonGenerationUpdated, fmt.Sprintf("generation change from (%d) to (%d)", oldObj.GetGeneration(), newObj.GetGeneration()))
				return true
			}

			if newObj.GetDeletionTimestamp().IsZero() != oldObj.GetDeletionTimestamp().IsZero() {
				fireEvent(newObj, ReasonDeletionTimestamp, "deletion timestamp has been added")
				return true
			}

			if len(oldObj.GetLabels()) != len(newObj.GetLabels()) || !reflect.DeepEqual(oldObj.GetLabels(), newObj.GetLabels()) {
				fireEvent(newObj, ReasonLabelsUpdated, fmt.Sprintf("labels updated from (%+v) to (%+v)", oldObj.GetLabels(), newObj.GetLabels()))
				return true
			}

			// Filter out internal operator annotations before comparison
			// This prevents cascading reconciliations from excluded annotations
			oldAnnFiltered := filterAnnotations(oldObj.GetAnnotations())
			newAnnFiltered := filterAnnotations(newObj.GetAnnotations())

			if !reflect.DeepEqual(oldAnnFiltered, newAnnFiltered) {
				fireEvent(newObj, ReasonAnnotationsUpdated, fmt.Sprintf("annotations updated from (%+v) to (%+v)", oldAnnFiltered, newAnnFiltered))
				return true
			}

			if len(oldObj.GetFinalizers()) != len(newObj.GetFinalizers()) || !reflect.DeepEqual(oldObj.GetFinalizers(), newObj.GetFinalizers()) {
				fireEvent(newObj, ReasonFinalizersUpdated, fmt.Sprintf("finalizers updated from (%+v) to (%+v)", oldObj.GetFinalizers(), newObj.GetFinalizers()))
				return true
			}

			if len(oldObj.GetOwnerReferences()) != len(newObj.GetOwnerReferences()) ||
				!reflect.DeepEqual(oldObj.GetOwnerReferences(), newObj.GetOwnerReferences()) {
				fireEvent(newObj, ReasonOwnerReferencesUpdated, fmt.Sprintf("owner-references updated from (%+v) to (%+v)", oldObj.GetOwnerReferences(), newObj.GetOwnerReferences()))
				return true
			}

			oldRes, newRes := getRes(ev.ObjectOld), getRes(ev.ObjectNew)
			if oldRes.Enabled != newRes.Enabled {
				return true
			}

			// Only check isReady for Kloudlite resources (where both old and new have isReady set)
			// For non-Kloudlite resources (Pods, Jobs, etc.), rely on other predicates above
			if oldRes.Status.IsReady != nil && newRes.Status.IsReady != nil {
				if *oldRes.Status.IsReady != *newRes.Status.IsReady {
					fireEvent(newObj, ReasonStatusIsReadyChanged, fmt.Sprintf("resource isReady changed from (%v) to (%v)", *oldRes.Status.IsReady, *newRes.Status.IsReady))
					return true
				}
			}

			if len(oldRes.Status.Checks) != len(newRes.Status.Checks) {
				fireEvent(newObj, ReasonStatusChecksUpdated, fmt.Sprintf("resource status.checks changed from (%+v) to (%+v)", oldRes.Status.Checks, newRes.Status.Checks))
				return true
			}

			for k, v := range oldRes.Status.Checks {
				if !AreChecksEqual(newRes.Status.Checks[k], v) {
					fireEvent(newObj, ReasonStatusChecksUpdated, fmt.Sprintf("resource status.checks changed from (%+v) to (%+v)", oldRes.Status.Checks, newRes.Status.Checks))
					return true
				}
			}

			return false
		},
	}
}
