package types

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
