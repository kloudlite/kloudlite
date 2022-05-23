package types

type ReconReq struct {
	stateData map[string]string
}

func (req *ReconReq) GetStateData(key string) string {
	if req.stateData == nil {
		req.stateData = map[string]string{}
	}
	return req.stateData[key]
}

func (req *ReconReq) SetStateData(key, value string) {
	if req.stateData == nil {
		req.stateData = map[string]string{}
	}
	req.stateData[key] = value
}
