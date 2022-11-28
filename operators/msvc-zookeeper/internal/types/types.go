package types

type MsvcOutput struct {
	RootUsername string `json:"ROOT_USERNAME"`
	RootPassword string `json:"ROOT_PASSWORD"`
	Host         string `json:"HOST"`
	Port         string `json:"PORT"`
}
