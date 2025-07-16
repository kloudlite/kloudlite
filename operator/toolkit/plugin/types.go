package plugin

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/kloudlite/kloudlite/operator/toolkit/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// +kubebuilder:object:generate=true
type Export struct {
	ViaSecret string `json:"viaSecret,omitempty"`
	Template  string `json:"template,omitempty"`
}

type (
	getSecret    func(secretName string) (map[string]string, error)
	getConfigMap func(secretName string) (map[string]string, error)
)

func (ex Export) ParseKV(ctx context.Context, client client.Client, namespace string, constantsStruct any) (map[string]string, error) {
	getSecret := func(secretName string) (map[string]string, error) {
		secret := &corev1.Secret{}
		if err := client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: secretName}, secret); err != nil {
			return nil, err
		}
		m := make(map[string]string, len(secret.Data))
		for k, v := range secret.Data {
			m[k] = string(v)
		}
		return m, nil
	}

	getConfigMap := func(secretName string) (map[string]string, error) {
		cfgmap := &corev1.ConfigMap{}
		if err := client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: secretName}, cfgmap); err != nil {
			return nil, err
		}
		m := make(map[string]string, len(cfgmap.Data)+len(cfgmap.BinaryData))
		for k, v := range cfgmap.Data {
			m[k] = string(v)
		}
		for k, v := range cfgmap.BinaryData {
			m[k] = string(v)
		}
		return m, nil
	}

	return parseKV(&ex, getSecret, getConfigMap, constantsStruct)
}

func parseKV(ex *Export, getSecret getSecret, getConfigMap getConfigMap, constantsStruct any) (map[string]string, error) {
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
