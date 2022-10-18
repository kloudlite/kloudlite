package types

type MsvcOutput struct {
	RootPassword      string `json:"ROOT_PASSWORD"`
	MysqlUserPassword string `json:"MYSQL_USER_PASSWORD"`
	Hosts             string `json:"HOSTS"`
	DSN               string `json:"DSN,omitempty"`
	URI               string `json:"URI"`
}

type MresOutput struct {
	Username string `json:"USERNAME"`
	Password string `json:"PASSWORD"`
	Hosts    string `json:"HOSTS"`
	DbName   string `json:"DB_NAME"`
	DSN      string `json:"DSN"`
	URI      string `json:"URI"`
}
