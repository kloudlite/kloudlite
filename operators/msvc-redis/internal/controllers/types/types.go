package types

type MsvcOutput struct {
	RootPassword string `json:"ROOT_PASSWORD"`
	Hosts        string `json:"HOSTS"`
	Uri          string `json:"URI"`
}

type MresOutput struct {
	Hosts    string `json:"HOSTS"`
	Password string `json:"PASSWORD"`
	Username string `json:"USERNAME"`
	Prefix   string `json:"PREFIX"`
	Uri      string `json:"URI"`
}
