package devinfo

import (
	"encoding/base64"
	"encoding/json"
)

type DeviceInfo struct {
	Name        string `json:"name"`
	AccountName string `json:"accountName"`
	ClusterName string `json:"clusterName"`
}

func (d *DeviceInfo) ToBase64() (*string, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}

	s := base64.StdEncoding.EncodeToString(b)
	return &s, nil
}

func (d *DeviceInfo) FromBase64(str string) error {
	b, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, d)
}

func (d *DeviceInfo) String() string {
	b, _ := json.Marshal(d)
	return string(b)
}
