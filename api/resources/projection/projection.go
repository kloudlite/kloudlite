package projection

import (
	"github.com/kloudlite/kloudlite/api/resources/events"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const HashLabel = "kloudlite.io/hash"

type Selector struct {
	Labels map[string]string
	Hash   string
}

type ObjectState struct {
	Object client.Object
	Dirty  *events.DirtyObject
}

func ScopedNamespace(clusterScoped bool, namespace string) string {
	if clusterScoped {
		return ""
	}
	return namespace
}

func MatchesObject(object client.Object, selector Selector) bool {
	return MatchesLabels(object.GetLabels(), selector)
}

func MatchesLabels(labels map[string]string, selector Selector) bool {
	if selector.Hash != "" && labels[HashLabel] != selector.Hash {
		return false
	}
	for key, value := range selector.Labels {
		if labels[key] != value {
			return false
		}
	}
	return true
}

func NewDirtyObject(alias string, namespace string, name string, reason events.DirtyReason, labels map[string]string) events.DirtyObject {
	return events.DirtyObject{Resource: alias, Namespace: namespace, Name: name, Reason: reason, Labels: CloneLabels(labels)}
}

func CleanEvent(alias string, namespace string, name string, dirty events.DirtyObject) events.Event {
	return events.Event{Type: events.EventClean, Resource: alias, Namespace: namespace, Name: name, Dirty: &dirty}
}

func ObjectInitialState(object client.Object, dirty *events.DirtyObject) ObjectState {
	return ObjectState{Object: object, Dirty: dirty}
}

func CloneLabels(labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return nil
	}
	clone := make(map[string]string, len(labels))
	for key, value := range labels {
		clone[key] = value
	}
	return clone
}
