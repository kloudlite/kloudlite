package operator

import (
	"encoding/json"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type klResource struct {
	Status *Status `json:"status"`
}

func getStatus(obj client.Object) *Status {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	var res klResource
	if err := json.Unmarshal(b, &res); err != nil {
		return nil
	}

	return res.Status
}

func ReconcileFilter() predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(ev event.UpdateEvent) bool {
			oldObj := ev.ObjectOld
			newObj := ev.ObjectNew

			if oldObj.GetGeneration() > newObj.GetGeneration() {
				return true
			}

			if len(oldObj.GetLabels()) != len(newObj.GetLabels()) || !reflect.DeepEqual(oldObj.GetLabels(), newObj.GetLabels()) {
				return true
			}

			if len(oldObj.GetAnnotations()) != len(newObj.GetAnnotations()) ||
				!reflect.DeepEqual(oldObj.GetAnnotations(), newObj.GetAnnotations()) {
				return true
			}

			if len(oldObj.GetFinalizers()) != len(newObj.GetFinalizers()) || !reflect.DeepEqual(oldObj.GetFinalizers(), newObj.GetFinalizers()) {
				return true
			}

			if len(oldObj.GetOwnerReferences()) != len(newObj.GetOwnerReferences()) ||
				!reflect.DeepEqual(oldObj.GetOwnerReferences(), newObj.GetOwnerReferences()) {
				return true
			}

			oldStatus, newStatus := getStatus(ev.ObjectOld), getStatus(ev.ObjectNew)
			if oldStatus == nil || newStatus == nil {
				// this is not our object, it is some other k8s resource, just defaulting it to be always watched
				return true
			}

			if oldStatus.IsReady != newStatus.IsReady {
				return true
			}

			if len(oldStatus.Checks) != len(newStatus.Checks) {
				return true
			}
			for k, v := range oldStatus.Checks {
				if newStatus.Checks[k] != v {
					return true
				}
			}
			return false
		},
	}
}
