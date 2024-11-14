package plugin

import (
	"fmt"
	"io"
	"strings"

	"github.com/valyala/fasttemplate"
)

type ValueRef string

// +kubebuilder:object:generate=true
type Export struct {
	ViaSecret string              `json:"viaSecret,omitempty"`
	KV        map[string]ValueRef `json:"kv"`
}

type (
	GetSecret    func(secretName string) (map[string]string, error)
	GetConfigMap func(secretName string) (map[string]string, error)
)

func (ex Export) ParseKV(getSecret GetSecret, getConfigMap GetConfigMap) (map[string]string, error) {
	result := make(map[string]string, len(ex.KV))

	secrets := make(map[string]map[string]string)
	configs := make(map[string]map[string]string)

	for k, v := range ex.KV {
		value, err2 := fasttemplate.ExecuteFuncStringWithErr(string(v), "{", "}", func(w io.Writer, tag string) (int, error) {
			var err error
			switch {
			case strings.HasPrefix(tag, "secret/"):
				{
					sp := strings.SplitN(tag, "/", 3)
					if len(sp) != 3 {
						return 0, fmt.Errorf("invalid secret template, must be of format `secret/<secret-name>/<secret-key>`")
					}
					secret, ok := secrets[sp[1]]
					if !ok {
						secret, err = getSecret(sp[1])
						if err != nil {
							return 0, err
						}
						secrets[sp[1]] = secret
					}
					if v, ok := secret[sp[2]]; ok {
						return w.Write([]byte(v))
					}
					return 0, fmt.Errorf("secret: `%s` does not have key: `%s`", sp[1], sp[2])
				}
			case strings.HasPrefix(tag, "config/"):
				{
					sp := strings.SplitN(tag, "/", 3)
					if len(sp) != 3 {
						return 0, fmt.Errorf("invalid secret template, must be of format `config/<config-name>/<config-key>`")
					}
					config, ok := configs[sp[1]]
					if !ok {
						config, err = getConfigMap(sp[1])
						if err != nil {
							return 0, err
						}
						configs[sp[1]] = config
					}
					if v, ok := config[sp[2]]; ok {
						return w.Write([]byte(v))
					}
					return 0, fmt.Errorf("config: `%s` does not have key: `%s`", sp[1], sp[2])
				}
			default:
				{
					return 0, fmt.Errorf("incorrect template")
				}
			}
		})

		if err2 != nil {
			return nil, err2
		}

		result[k] = value
	}

	return result, nil
}
