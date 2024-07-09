package types

import "encoding/json"

type MsvcOutput struct {
	RootPassword        string `json:"ROOT_PASSWORD,omitempty"`
	ReplicationPassword string `json:"REPLICATION_PASSWORD,omitempty"`
	Hosts               string `json:"HOSTS,omitempty"`
	DSN                 string `json:"DSN,omitempty"`
	URI                 string `json:"URI,omitempty"`
	ExternalHost        string `json:"EXTERNAL_HOST,omitempty"`
}

type MresOutput struct {
	Username    string `json:"USERNAME"`
	Password    string `json:"PASSWORD"`
	Hosts       string `json:"HOSTS"`
	DbName      string `json:"DB_NAME"`
	DSN         string `json:"DSN"`
	URI         string `json:"URI"`
	ExternalDSN string `json:"EXTERNAL_DSN,omitempty"`
}

type StandaloneServiceOutput struct {
  StandaloneDatabaseOutput `json:",inline"`
}

func (o *StandaloneServiceOutput) ToSecretData() map[string][]byte {
  m := make(map[string][]byte)
  m["USERNAME"] = []byte(o.Username)
  m["PASSWORD"] = []byte(o.Password)

  m["PORT"] = []byte(o.Port)

  m["DB_NAME"] = []byte(o.DbName)

  m["HOST"] = []byte(o.Host)
  m["URI"] = []byte(o.URI)
  m["DSN"] = []byte(o.DSN)

  m["CLUSTER_LOCAL_HOST"] = []byte(o.ClusterLocalHost)
  m["CLUSTER_LOCAL_URI"] = []byte(o.ClusterLocalURI)
  m["CLUSTER_LOCAL_DSN"] = []byte(o.ClusterLocalDSN)

  return m
}

type StandaloneDatabaseOutput struct {
  Username    string `json:"USERNAME"`
  Password    string `json:"PASSWORD"`

  Port       string `json:"PORT"`

  DbName      string `json:"DB_NAME"`

  Host       string `json:"HOST"`
  URI         string `json:"URI"`
  DSN         string `json:"DSN"`

  ClusterLocalHost string `json:"CLUSTER_LOCAL_HOST"`
  ClusterLocalURI string `json:"CLUSTER_LOCAL_URI"`
  ClusterLocalDSN string `json:"CLUSTER_LOCAL_DSN"`
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
