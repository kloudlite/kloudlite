package raw_json

import (
	"encoding/json"

	"github.com/kloudlite/operator/pkg/errors"
)

// +kubebuilder:pruning:PreserveUnknownFields
// +kubebuilder:validation:Schemaless
// +kubebuilder:validation:Type=object
// +kubebuilder:object:generate=true

type RawJson struct {
	items           map[string]any
	json.RawMessage `json:",omitempty,inline"`
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

func (s *RawJson) EnsureRawJson() *RawJson {
  if s == nil { 
    return &RawJson{}
  }
  return s
}

// suppressing error
func (s *RawJson) fillMap() {
  s = s.EnsureRawJson()
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

func (s *RawJson) complete() error {
  s = s.EnsureRawJson()
	b, err := json.Marshal(s.items)
	if err != nil {
		return err
	}
	s.RawMessage = b
	return nil
}

func (s *RawJson) Len() int {
  s = s.EnsureRawJson()
	s.fillMap()
	return len(s.items)
}

func (s *RawJson) Set(key string, value any) error {
  s = s.EnsureRawJson()
	s.fillMap()
	s.items[key] = value
	return s.complete()
}

func (s *RawJson) SetFromMap(m map[string]any) error {
  s = s.EnsureRawJson()
	s.fillMap()
	for k, v := range m {
		s.items[k] = v
	}
	return s.complete()
}

func (s *RawJson) Exists(keys ...string) bool {
  s = s.EnsureRawJson()
	s.fillMap()
	for i := range keys {
		if _, ok := s.items[keys[i]]; ok {
			return true
		}
	}
	return false
}

func (s *RawJson) Delete(key string) error {
  s = s.EnsureRawJson()
	s.fillMap()
	c := len(s.items)
	delete(s.items, key)
	if c != len(s.items) {
		return s.complete()
	}
	return nil
}

func (s *RawJson) Get(key string, fillInto any) error {
  s = s.EnsureRawJson()
	s.fillMap()
	value, ok := s.items[key]
	if !ok {
		fillInto = nil
		// return nil
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
  s = s.EnsureRawJson()
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

func (s *RawJson) ToString() string {
  s = s.EnsureRawJson()
	if s.RawMessage == nil {
		return ""
	}
	return string(s.RawMessage)
}
