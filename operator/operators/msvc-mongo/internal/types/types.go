package types

type MsvcOutput struct {
	RootPassword string `json:"ROOT_PASSWORD"`
	Hosts        string `json:"HOSTS"`
	URI          string `json:"URI"`
}

type MresOutput struct {
	Username string `json:"USERNAME"`
	Password string `json:"PASSWORD"`
	Hosts    string `json:"HOSTS"`
	DbName   string `json:"DB_NAME"`
	URI      string `json:"uri"`
}
