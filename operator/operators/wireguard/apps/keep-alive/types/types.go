package types

import "sigs.k8s.io/yaml"

type Conf struct {
	Cidrs    []string `yaml:"cidrs"`
	Interval int      `yaml:"interval"`
}

func (d *Conf) ToYaml() ([]byte, error) {
	return yaml.Marshal(d)
}

func (d *Conf) FromYaml(data []byte) error {
	return yaml.Unmarshal(data, d)
}
