package types

import "encoding/json"

type StandaloneSvcOutput struct {
	RootUsername string `json:"ROOT_USERNAME"`
	RootPassword string `json:"ROOT_PASSWORD"`

	ClusterLocalHosts string `json:"CLUSTER_LOCAL_HOSTS"`
	ClusterLocalURI   string `json:"CLUSTER_LOCAL_URI"`

	GlobalVPNHosts string `json:"GLOBAL_VPN_HOSTS"`
	GlobalVpnURI   string `json:"GLOBAL_VPN_URI"`

	AuthSource string `json:"AUTH_SOURCE"`
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

	ClusterLocalHosts string `json:"CLUSTER_LOCAL_HOSTS"`
	ClusterLocalURI   string `json:"CLUSTER_LOCAL_URI"`

	GlobalVpnHosts string `json:"GLOBAL_VPN_HOSTS"`
	GlobalVpnURI   string `json:"GLOBAL_VPN_URI"`

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

type DatabaseOutput struct {
	Username string `json:"USERNAME"`
	Password string `json:"PASSWORD"`

	DbName string `json:"DB_NAME"`

	ClusterLocalHosts string `json:"CLUSTER_LOCAL_HOSTS"`
	ClusterLocalURI   string `json:"CLUSTER_LOCAL_URI"`

	GlobalVPNHosts string `json:"GLOBAL_VPN_HOSTS"`
	GlobalVpnURI   string `json:"GLOBAL_VPN_URI"`
}

func ExtractPVCLabelsFromStatefulSetLabels(m map[string]string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/component": m["app.kubernetes.io/name"],
		"app.kubernetes.io/instance":  m["app.kubernetes.io/instance"],
		"app.kubernetes.io/name":      m["app.kubernetes.io/name"],
	}
}
