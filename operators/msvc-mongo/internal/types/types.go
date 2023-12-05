package types

import "encoding/json"

type MsvcOutput struct {
	RootPassword string `json:"ROOT_PASSWORD"`
	Hosts        string `json:"HOSTS"`
	URI          string `json:"URI"`
}

type ClusterSvcOutput struct {
	RootUsername    string `json:"ROOT_USERNAME"`
	RootPassword    string `json:"ROOT_PASSWORD"`
	Hosts           string `json:"HOSTS"`
	URI             string `json:"URI"`
	AuthSource      string `json:"AUTH_SOURCE"`
	ReplicasSetName string `json:"REPLICASET_NAME"`
	ReplicaSetKey   string `json:"REPLICASET_KEY"`
}

func (cso ClusterSvcOutput) ToMap() (map[string]string, error) {
	b, err := json.Marshal(cso)
	if err != nil {
		return nil, err
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

type MresOutput struct {
	Username string `json:"USERNAME"`
	Password string `json:"PASSWORD"`
	Hosts    string `json:"HOSTS"`
	DbName   string `json:"DB_NAME"`
	URI      string `json:"URI"`
}
