package types

import "encoding/json"

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

func (r *RawJson) toMap() (map[string]any, error) {
	m, err := r.RawMessage.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var v map[string]any
	if err := json.Unmarshal(m, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func (r *RawJson) ToMap() (map[string]any, error) {
	return r.toMap()
}

func (r *RawJson) Get(k string) (any, bool) {
	m, err := r.ToMap()
	if err != nil {
		return nil, false
	}
	v, ok := m[k]
	return v, ok
}

func (r *RawJson) FillFrom(m map[string]any, upsert ...bool) error {
	canUpsert := true
	if len(upsert) > 0 {
		canUpsert = upsert[0]
	}

	if !canUpsert {
		b, err := json.Marshal(m)
		if err != nil {
			return err
		}
		r.RawMessage = b
		return nil
	}

	currMap, err := r.toMap()
	if err != nil {
		return err
	}
	for k, v := range m {
		currMap[k] = v
	}
	b, err := json.Marshal(currMap)
	if err != nil {
		return err
	}
	r.RawMessage = b
	return nil
}
