package plugin

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/kloudlite/operator/pkg/errors"
	"sigs.k8s.io/yaml"
)

type ValueRef string

// +kubebuilder:object:generate=true
type Export struct {
	ViaSecret string `json:"viaSecret,omitempty"`
	Template  string `json:"template"`
}

type (
	GetSecret    func(secretName string) (map[string]string, error)
	GetConfigMap func(secretName string) (map[string]string, error)
)

func (ex Export) ParseKV(getSecret GetSecret, getConfigMap GetConfigMap, constantsStruct any) (map[string]string, error) {
	secrets := make(map[string]map[string]string)
	configs := make(map[string]map[string]string)

	getSecretWithCache := func(secretName string) (map[string]string, error) {
		secret, ok := secrets[secretName]
		if ok {
			return secret, nil
		}
		secret, err := getSecret(secretName)
		if err != nil {
			return nil, err
		}
		secrets[secretName] = secret
		return secret, nil
	}

	getConfigMapWithCache := func(configName string) (map[string]string, error) {
		config, ok := configs[configName]
		if ok {
			return config, nil
		}

		config, err := getConfigMap(configName)
		if err != nil {
			return nil, err
		}
		configs[configName] = config
		return config, nil
	}

	t := template.New("tag")
	t.Funcs(template.FuncMap{
		"secret": func(secretRef string) (string, error) {
			s := strings.Split(secretRef, "/")
			if len(s) != 2 {
				return "", fmt.Errorf("invalid secretRef value")
			}
			m, err := getSecretWithCache(s[0])
			if err != nil {
				return "", err
			}
			v, ok := m[s[1]]
			if !ok {
				return "", err
			}
			return v, nil
		},

		"config": func(configRef string) (string, error) {
			s := strings.Split(configRef, "/")
			if len(s) != 2 {
				return "", fmt.Errorf("invalid configRef value")
			}
			m, err := getConfigMapWithCache(s[0])
			if err != nil {
				return "", err
			}
			v, ok := m[s[1]]
			if !ok {
				return "", err
			}
			return v, nil
		},
	})

	if _, err := t.Parse(ex.Template); err != nil {
		return nil, err
	}

	out := new(bytes.Buffer)
	if err := t.ExecuteTemplate(out, t.Name(), constantsStruct); err != nil {
		return nil, errors.NewEf(err, "could not execute template")
	}

	var m map[string]string
	if err := yaml.Unmarshal(out.Bytes(), &m); err != nil {
		return nil, err
	}
	return m, nil
}
