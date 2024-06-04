package types

import (
	"fmt"
	"sync"
)

type SyncMap[K comparable, V any] struct {
	mu    sync.RWMutex
	value map[K]V
}

func (m *SyncMap[K, V]) Keys() []K {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]K, 0, len(m.value))
	for k := range m.value {
		keys = append(keys, k)
	}
	return keys
}

func (m *SyncMap[K, V]) Get(key K) V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.value[key]
}

func (m *SyncMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.value[key] = value
}

func (m *SyncMap[K, V]) Debug() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range m.value {
		fmt.Printf("%v: %v\n", k, v)
	}
}

func NewSyncMap[K comparable, V any](m map[K]V) *SyncMap[K, V] {
	return &SyncMap[K, V]{
		mu:    sync.RWMutex{},
		value: m,
	}
}
