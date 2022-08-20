package raw_json

import (
	"encoding/json"

	"operators.kloudlite.io/lib/errors"
)

// +kubebuilder:pruning:PreserveUnknownFields
// +kubebuilder:validation:Schemaless
// +kubebuilder:validation:Type=object

type RawJson struct {
	items map[string]any
	// RawJson[string, json.RawMessage] `json:",inline"`
	json.RawMessage `json:",inline"`
}

func (k *RawJson) DeepCopyInto(out *RawJson) {
	*out = *k
}

func (k *RawJson) DeepCopy() *RawJson {
	if k == nil {
		return nil
	}
	out := new(RawJson)
	k.DeepCopyInto(out)
	return out
}

// old set
// type RawJson[K ~string, V any] struct {
// 	json.RawMessage `json:",inline"`
// }

// suppressing error
func (s *RawJson) fillMap() {
	if s.RawMessage != nil {
		s.items = map[string]any{}
		m, err := s.RawMessage.MarshalJSON()
		if err != nil {
			panic(err)
			return
		}
		if err := json.Unmarshal(m, &s.items); err != nil {
			panic(err)
			return
		}
	}

	if s.items == nil {
		s.items = map[string]any{}
	}
}

// func (s *RawJson) toMap() (map[K]V, error) {
// 	m, err := s.RawMessage.MarshalJSON()
// 	if err != nil {
// 		return nil, err
// 	}
// 	var v map[K]V
// 	if err := json.Unmarshal(m, &v); err != nil {
// 		return nil, err
// 	}
// 	if v == nil {
// 		v = map[K]V{}
// 	}
// 	return v, nil
// }

// func (s *RawJson) ToMap() (map[K]V, error) {
// 	return s.toMap()
// }

func (s *RawJson) Set(key string, value any) error {
	s.fillMap()
	s.items[key] = value
	b, err := json.Marshal(s.items)
	if err != nil {
		return err
	}
	s.RawMessage = b
	return nil
}

func (s *RawJson) Exists(keys ...string) bool {
	s.fillMap()
	for _, key := range keys {
		if _, ok := s.items[key]; !ok {
			return false
		}
	}
	return true
}

func (s *RawJson) Delete(key string) error {
	s.fillMap()
	c := len(s.items)
	delete(s.items, key)
	if c != len(s.items) {
		b, err := json.Marshal(s.items)
		if err != nil {
			return err
		}
		s.RawMessage = b
	}
	return nil
}

func (s *RawJson) Get(key string, fillInto any) error {
	s.fillMap()
	// m, err := s.toMap()
	// if err != nil {
	// 	return *new(V), false
	// }
	//
	value, ok := s.items[key]
	if !ok {
		return errors.Newf("key %s does not exist", key)
	}
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, fillInto); err != nil {
		return err
	}
	return nil
}

func (s *RawJson) GetString(key string) (string, bool) {
	s.fillMap()
	x, ok := s.items[key]
	if !ok {
		return "", false
	}
	s2, ok := (x).(string)
	if !ok {
		return "", false
	}
	return s2, true
}

//
//
// func (s *RawJson) GetInt(key string) (int, bool) {
// 	s.fillMap()
// 	x, ok := s.Get(key)
// 	if !ok {
// 		return 0, false
// 	}
// 	s2, ok := (x).(float64)
// 	if !ok {
// 		return 0, false
// 	}
// 	return int(s2), true
// }
//
// func (s *RawJson) GetInt64(key string) (int64, bool) {
// 	s.fillMap()
// 	x, ok := s.Get(key)
// 	if !ok {
// 		return 0, false
// 	}
// 	s2, ok := (x).(float64)
// 	if !ok {
// 		return 0, false
// 	}
// 	return int64(s2), true
// }
// --- old set
