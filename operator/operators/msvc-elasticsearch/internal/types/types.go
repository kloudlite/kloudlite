package types

type MsvcOutput struct {
	Username string `json:"USERNAME"`
	Password string `json:"PASSWORD"`
	Hosts    string `json:"HOSTS"`
	Port     uint16 `json:"PORT"`
	Uri      string `json:"URI"`
}
