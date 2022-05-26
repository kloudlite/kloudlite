package types

import (
	"encoding/json"
)

type ReconReq struct {
	stateData map[string]any
}

func (req *ReconReq) GetStateData(key string) (any, bool) {
	if req.stateData == nil {
		req.stateData = map[string]any{}
	}
	v, ok := req.stateData[key]
	return v, ok
}

func (req *ReconReq) SetStateData(key string, value any) {
	if req.stateData == nil {
		req.stateData = map[string]any{}
	}
	req.stateData[key] = value
}

// +kubebuilder:object:generate=true

type RawJson struct {
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	json.RawMessage `json:",inline"`
}

func (r RawJson) toMap() (map[string]any, error) {
	//func (r RawJson) toMap() (map[string]interface{}, error) {
	m, err := r.RawMessage.MarshalJSON()
	if err != nil {
		return map[string]any{}, err
	}
	v := make(map[string]any, 1)
	if err := json.Unmarshal(m, &v); err != nil {
		return nil, err
	}

	if v == nil {
		v = map[string]any{}
	}

	return v, nil
}

func (r RawJson) ToMap() (map[string]any, error) {
	return r.toMap()
	//m, _ := r.toMap()
	//if m == nil {
	//	m = map[string]any{}
	//}
	//return m, nil
}

func (r RawJson) Get(k string) (any, bool) {
	m, err := r.ToMap()
	if err != nil {
		return nil, false
	}
	v, ok := m[k]
	return v, ok
}
