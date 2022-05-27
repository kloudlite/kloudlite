package raw_json

import (
	"encoding/json"
)

// +kubebuilder:object:generate=false

// +kubebuilder:pruning:PreserveUnknownFields
// +kubebuilder:validation:Schemaless
// +kubebuilder:validation:Type=object

type KubeRawJson struct {
	RawJson[string, any] `json:",inline"`
}

func (k *KubeRawJson) DeepCopyInto(out *KubeRawJson) {
	*out = *k
	k.RawJson.DeepCopyInto(&out.RawJson)
}

func (k *KubeRawJson) DeepCopy() *KubeRawJson {
	if k == nil {
		return nil
	}
	out := new(KubeRawJson)
	k.DeepCopyInto(out)
	return out
}

func (k *KubeRawJson) UnmarshalJSON(data []byte) error {
	return k.UnmarshalJSON(data)
}

func (k *KubeRawJson) MarshalJSON() ([]byte, error) {
	return k.MarshalJSON()
}

type RawJson[K ~string, V any] struct {
	json.RawMessage `json:",inline"`
}

func (k *RawJson[K, V]) DeepCopyInto(out *RawJson[K, V]) {
	*out = *k
	k.DeepCopyInto(out)
}

func (k *RawJson[K, V]) DeepCopy() *RawJson[K, V] {
	if k == nil {
		return nil
	}
	out := new(RawJson[K, V])
	k.DeepCopyInto(out)
	return out
}

func (k *RawJson[K, V]) UnmarshalJSON(data []byte) error {
	return k.UnmarshalJSON(data)
}

func (k *RawJson[K, V]) MarshalJSON() ([]byte, error) {
	return k.MarshalJSON()
}

func (s RawJson[K, V]) toMap() (map[K]V, error) {
	m, err := s.RawMessage.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var v map[K]V
	if err := json.Unmarshal(m, &v); err != nil {
		return nil, err
	}
	if v == nil {
		v = map[K]V{}
	}
	return v, nil
}

func (s RawJson[K, V]) ToMap() (map[K]V, error) {
	return s.toMap()
}

func (s *RawJson[K, V]) Set(key K, value V) error {
	return s.Merge(map[K]V{key: value})
}

func (s *RawJson[K, V]) Merge(val map[K]V) error {
	m, err := s.toMap()
	if err != nil {
		return nil
	}

	for k, v := range val {
		m[k] = v
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	s.RawMessage = b
	return nil
}

func (s *RawJson[K, V]) Get(key K) (V, bool) {
	m, err := s.toMap()
	if err != nil {
		return *new(V), false
	}

	value, ok := m[key]
	if !ok {
		return *new(V), false
	}
	return value, true
}

func (s *RawJson[K, V]) GetString(key K) (string, bool) {
	x, ok := s.Get(key)
	if !ok {
		return "", false
	}
	s2, ok := any(x).(string)
	if !ok {
		return "", false
	}
	return s2, true
}
