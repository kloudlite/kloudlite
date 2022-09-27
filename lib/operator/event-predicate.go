package operator

import (
	"encoding/json"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type resStatus struct {
	Status Status `json:"status"`
}

func getStatus(obj client.Object) *resStatus {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	var res resStatus
	if err := json.Unmarshal(b, &res); err != nil {
		return nil
	}
	return &res
}

func ReconcileFilter() predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(ev event.UpdateEvent) bool {
			if ev.ObjectNew.GetGeneration() > ev.ObjectOld.GetGeneration() {
				return true
			}

			oldObj, newObj := getStatus(ev.ObjectOld), getStatus(ev.ObjectNew)

			if !reflect.DeepEqual(oldObj.Status.Checks, newObj.Status.Checks) {
				return true
			}

			if oldObj.Status.IsReady != newObj.Status.IsReady {
				return true
			}

			return false
		},
	}
}
