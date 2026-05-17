package store

import (
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const hashLabel = "kloudlite.io/hash"

type Selector struct {
	Labels map[string]string
	Hash   string
}

type Store struct {
	mu      sync.RWMutex
	objects map[string]map[string]client.Object
	ready   map[string]bool
}

func New() *Store {
	return &Store{objects: map[string]map[string]client.Object{}, ready: map[string]bool{}}
}

func (s *Store) SetReady(alias string, namespace string, ready bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ready[scopeKey(alias, namespace)] = ready
}

func (s *Store) Ready(alias string, namespace string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ready[scopeKey(alias, namespace)]
}

func (s *Store) Upsert(alias string, object client.Object) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := scopeKey(alias, object.GetNamespace())
	if s.objects[key] == nil {
		s.objects[key] = map[string]client.Object{}
	}
	s.objects[key][object.GetName()] = copyObject(object)
}

func (s *Store) ReplaceScope(alias string, namespace string, objects []client.Object) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := scopeKey(alias, namespace)
	s.objects[key] = map[string]client.Object{}
	for _, object := range objects {
		s.objects[key][object.GetName()] = copyObject(object)
	}
	s.ready[key] = true
}

func (s *Store) Get(alias string, namespace string, name string) (client.Object, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	object, ok := s.objects[scopeKey(alias, namespace)][name]
	if !ok {
		return nil, false
	}
	return copyObject(object), true
}

func (s *Store) List(alias string, namespace string, selector Selector) []client.Object {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []client.Object{}
	for _, object := range s.objects[scopeKey(alias, namespace)] {
		if matches(object, selector) {
			items = append(items, copyObject(object))
		}
	}
	return items
}

func (s *Store) Delete(alias string, namespace string, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects[scopeKey(alias, namespace)], name)
}

func scopeKey(alias string, namespace string) string {
	return alias + "/" + namespace
}

func copyObject(object client.Object) client.Object {
	return object.DeepCopyObject().(client.Object)
}

func matches(object client.Object, selector Selector) bool {
	labels := object.GetLabels()
	if selector.Hash != "" && labels[hashLabel] != selector.Hash {
		return false
	}
	for key, value := range selector.Labels {
		if labels[key] != value {
			return false
		}
	}
	return true
}
