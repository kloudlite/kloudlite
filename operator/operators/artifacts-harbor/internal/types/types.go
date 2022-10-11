package types

type UserAccountOutput struct {
	DockerConfigJson string `json:".dockerconfigjson"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	Registry         string `json:"registry"`
	Project          string `json:"project"`
}
