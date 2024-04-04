package client

import (
	"fmt"

	"github.com/kloudlite/operator/apps/multi-cluster/apps/common"
	"github.com/kloudlite/operator/apps/multi-cluster/templates"
)

type Config struct {
	Endpoint   string `json:"endpoint"`
	PrivateKey string `json:"privateKey"`
	IpAddress  string `json:"ip"`

	Peers []common.Peer `json:"peers,omitempty"`
}

func (s *Config) String() string {
	return fmt.Sprintf("%#v", *s)
}

func (s *Config) toConfigBytes() ([]byte, error) {
	b, err := templates.ParseTemplate(templates.ClientConfg, *s)
	if err != nil {
		return nil, err
	}

	return b, nil
}
