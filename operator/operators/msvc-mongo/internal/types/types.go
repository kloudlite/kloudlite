package types

import "encoding/json"

type StandaloneSvcOutput struct {
	RootUsername string `json:"ROOT_USERNAME"`
	RootPassword string `json:"ROOT_PASSWORD"`

	DBName     string `json:"DB_NAME"`
	AuthSource string `json:"AUTH_SOURCE"`

	Port string `json:"PORT"`

	Host string `json:"HOST"`
	Addr string `json:"ADDR"`
	URI  string `json:"URI"`

	ClusterLocalHost string `json:".CLUSTER_LOCAL_HOST"`
	ClusterLocalAddr string `json:".CLUSTER_LOCAL_ADDR"`
	ClusterLocalURI  string `json:".CLUSTER_LOCAL_URI"`

	GlobalVpnHost string `json:".GLOBAL_VPN_HOST,omitempty"`
	GlobalVpnAddr string `json:".GLOBAL_VPN_ADDR,omitempty"`
	GlobalVpnURI  string `json:".GLOBAL_VPN_URI,omitempty"`
}

func (sso StandaloneSvcOutput) ToMap() (map[string]string, error) {
	b, err := json.Marshal(sso)
	if err != nil {
		return nil, err
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

type ClusterSvcOutput struct {
	RootUsername string `json:"ROOT_USERNAME"`
	RootPassword string `json:"ROOT_PASSWORD"`
	AuthSource   string `json:"AUTH_SOURCE"`

	ClusterLocalHosts string `json:".CLUSTER_LOCAL_HOSTS"`
	ClusterLocalURI   string `json:".CLUSTER_LOCAL_URI"`

	GlobalVpnHosts string `json:"GLOBAL_VPN_HOSTS,omitempty"`
	GlobalVpnURI   string `json:"GLOBAL_VPN_URI,omitempty"`

	ReplicasSetName string `json:"REPLICASET_NAME,omitempty"`
	ReplicaSetKey   string `json:"REPLICASET_KEY,omitempty"`
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

type DatabaseOutput struct {
	Username string `json:"USERNAME"`
	Password string `json:"PASSWORD"`

	DbName string `json:"DB_NAME"`

	ClusterLocalHosts string `json:"CLUSTER_LOCAL_HOSTS"`
	ClusterLocalURI   string `json:".CLUSTER_LOCAL_URI"`

	// just an alias to ClusterLocalURI/GlobalVpnURI
	URI string `json:"URI"`

	GlobalVPNHosts string `json:"GLOBAL_VPN_HOSTS,omitempty"`
	GlobalVpnURI   string `json:"GLOBAL_VPN_URI,omitempty"`
}

// func ExtractPVCLabelsFromStatefulSetLabels(m map[string]string) map[string]string {
// 	return map[string]string{
// 		"app.kubernetes.io/component": m["app.kubernetes.io/name"],
// 		"app.kubernetes.io/instance":  m["app.kubernetes.io/instance"],
// 		"app.kubernetes.io/name":      m["app.kubernetes.io/name"],
// 	}
// }

type StandaloneDatabaseOutput struct {
	Username string `json:"USERNAME"`
	Password string `json:"PASSWORD"`

	DbName string `json:"DB_NAME"`

	Port string `json:"PORT"`

	Host string `json:"HOST"`
	URI  string `json:"URI"`

	ClusterLocalHost string `json:".CLUSTER_LOCAL_HOST"`
	ClusterLocalURI  string `json:".CLUSTER_LOCAL_URI"`

	GlobalVpnHost string `json:".GLOBAL_VPN_HOST"`
	GlobalVpnURI  string `json:".GLOBAL_VPN_URI"`
}

func (sdo StandaloneDatabaseOutput) ToMap() (map[string]string, error) {
	b, err := json.Marshal(sdo)
	if err != nil {
		return nil, err
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}
