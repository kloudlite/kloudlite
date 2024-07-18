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
	Username string `json:"USERNAME"`
	Password string `json:"PASSWORD"`

	Port string `json:"PORT"`

	DbName string `json:"DB_NAME"`

	Host string `json:"HOST"`
	URI  string `json:"URI"`
	DSN  string `json:"DSN"`

	ClusterLocalHost string `json:".CLUSTER_LOCAL_HOST"`
	ClusterLocalURI  string `json:".CLUSTER_LOCAL_URI"`
	ClusterLocalDSN  string `json:".CLUSTER_LOCAL_DSN"`

	GlobalVPNHost string `json:".GLOBAL_VPN_HOST"`
	GlobalVPNURI  string `json:".GLOBAL_VPN_URI"`
	GlobalVPNDSN  string `json:".GLOBAL_VPN_DSN"`
}

func (o *StandaloneServiceOutput) ToMap() (map[string]string, error) {
	b, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

type StandaloneDatabaseOutput struct {
	Username string `json:"USERNAME"`
	Password string `json:"PASSWORD"`

	Port string `json:"PORT"`

	DbName string `json:"DB_NAME"`

	Host string `json:"HOST"`
	URI  string `json:"URI"`
	DSN  string `json:"DSN"`

	ClusterLocalHost string `json:".CLUSTER_LOCAL_HOST"`
	ClusterLocalURI  string `json:".CLUSTER_LOCAL_URI"`
	ClusterLocalDSN  string `json:".CLUSTER_LOCAL_DSN"`

	GlobalVPNHost string `json:".GLOBAL_VPN_HOST"`
	GlobalVPNURI  string `json:".GLOBAL_VPN_URI"`
	GlobalVPNDSN  string `json:".GLOBAL_VPN_DSN"`
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
