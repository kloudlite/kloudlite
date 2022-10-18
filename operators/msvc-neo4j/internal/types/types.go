package types

type MsvcOutput struct {
	RootPassword string `json:"ROOT_PASSWORD"`
	Hosts        string `json:"HOSTS"`
	AdminHosts   string `json:"ADMIN_HOSTS"`
	PortBolt     string `json:"PORT_BOLT"`
	PortHttp     string `json:"PORT_HTTP"`
	PortHttps    string `json:"PORT_HTTPS"`
	PortBackup   string `json:"PORT_BACKUP"`
	// BoltUri      string `json:"BOLT_URI"`
	// Neo4jUri     string `json:"NEO4J_URI"`
}
