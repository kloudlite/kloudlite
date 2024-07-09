package types

import "encoding/json"

type StandaloneOutput struct {
	Username string `json:"USERNAME"`
	Password string `json:"PASSWORD"`

	DbName string `json:"DB_NAME"`

	Port string `json:"PORT"`

	Host string `json:"HOST"`
	URI  string `json:"URI"`

	ClusterLocalHost string `json:"CLUSTER_LOCAL_HOST"`
	ClusterLocalURI  string `json:"CLUSTER_LOCAL_URI"`
}

func (so *StandaloneOutput) ToMap() (map[string]string, error) {
	m := make(map[string]string)
	b, err := json.Marshal(so)
	if err != nil {
		return nil, err
	}

  err = json.Unmarshal(b, &m)
  if err != nil {
    return nil, err
  }

  return m, nil
}

type StandaloneDatabaseOutput struct {
	Username string `json:"USERNAME"`
	Password string `json:"PASSWORD"`

	DbName string `json:"DB_NAME"`

	Port string `json:"PORT"`

	Host string `json:"HOST"`
	URI  string `json:"URI"`

	ClusterLocalHost string `json:"CLUSTER_LOCAL_HOST"`
	ClusterLocalURI  string `json:"CLUSTER_LOCAL_URI"`
}

func (do *StandaloneDatabaseOutput) ToMap() (map[string]string, error) {
	m := make(map[string]string)
	b, err := json.Marshal(do)
	if err != nil {
		return nil, err
	}

  err = json.Unmarshal(b, &m)
  if err != nil {
    return nil, err
  }

  return m, nil
}
