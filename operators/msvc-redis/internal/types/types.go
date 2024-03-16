package types

import (
	"encoding/json"
	"fmt"
)

func GenerateRedisURI(host string, password string, dbNumber int) string {
	return fmt.Sprintf("redis://:%s@%s/%d?allowUsernameInURI=true", password, host, dbNumber)
}

type MsvcOutput struct {
	Host string `json:"HOST"`
	Port string `json:"PORT"`
	Addr string `json:"ADDR"`

	Uri string `json:"URI"`

	RootPassword string `json:"ROOT_PASSWORD"`
}

func (mo *MsvcOutput) ToMap() (map[string]string, error) {
	m := make(map[string]string)
	b, err := json.Marshal(mo)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	return m, nil
}

type MresOutput struct {
	Hosts    string `json:"HOSTS"`
	Password string `json:"PASSWORD"`
	Username string `json:"USERNAME"`
	Prefix   string `json:"PREFIX"`
	Uri      string `json:"URI"`
}

type PrefixCredentialsData struct {
	Host string `json:"HOST"`
	Port string `json:"PORT"`
	Addr string `json:"ADDR"`
	DB   string `json:"DB"`

	Uri string `json:"URI"`

	Password string `json:"PASSWORD"`
	Prefix   string `json:"PREFIX"`
}
