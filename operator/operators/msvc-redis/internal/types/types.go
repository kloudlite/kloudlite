package types

import (
	"encoding/json"
)

type MsvcOutput struct {
	RootUsername string `json:"ROOT_USERNAME"`
	RootPassword string `json:"ROOT_PASSWORD"`
	Hosts        string `json:"HOSTS"`
	Uri          string `json:"URI"`
}

func (mo *MsvcOutput) ToMap() (map[string]string, error) {
	m := make(map[string]string)
	b, err := json.Marshal(mo)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	return m, nil
}

type MresOutput struct {
	Hosts    string `json:"HOSTS"`
	Password string `json:"PASSWORD"`
	Username string `json:"USERNAME"`
	Prefix   string `json:"PREFIX"`
	Uri      string `json:"URI"`
}

type PrefixCredentialsData struct {
	Hosts    string `json:"HOSTS"`
	Password string `json:"PASSWORD"`
	Username string `json:"USERNAME"`
	Prefix   string `json:"PREFIX"`
	Uri      string `json:"URI"`
}
