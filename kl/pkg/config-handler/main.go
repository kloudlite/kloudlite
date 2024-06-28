package confighandler

import (
	"io/fs"
	"os"

	fn "github.com/kloudlite/kl/pkg/functions"
	yaml "gopkg.in/yaml.v2"
)

var ErrKlFileNotExists = fn.Error("please ensure kl.yaml file by running \"kl init\" command in your workspace root.")

func ReadConfig[T any](path string) (*T, error) {
	var v T
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, ErrKlFileNotExists
	}
	if err := yaml.Unmarshal(b, &v); err != nil {
		return nil, fn.NewE(err)
	}

	return &v, nil
}

func WriteConfig(path string, v interface{}, perm fs.FileMode) error {

	b, err := os.ReadFile(path)
	if err != nil {
		// If file doesn't exist, create it
		if os.IsNotExist(err) {
			b, err = yaml.Marshal(v)
			if err != nil {
				return fn.NewE(err)
			}

			return os.WriteFile(path, b, perm)
		}

		return fn.NewE(err)
	}

	var config yaml.MapSlice
	if err := yaml.Unmarshal(b, &config); err != nil {
		return fn.NewE(err)
	}

	if err := fillConfig(&config, v); err != nil {
		return fn.NewE(err)
	}

	b, err = yaml.Marshal(config)
	if err != nil {
		return fn.NewE(err)
	}

	return os.WriteFile(path, b, perm)
}

// FillConfig updates src with values from dest
func fillConfig(src *yaml.MapSlice, dest interface{}) error {
	// Marshal dest to YAML
	destBytes, err := yaml.Marshal(dest)
	if err != nil {
		return fn.NewE(err)
	}

	// Unmarshal dest YAML into a MapSlice
	var destMapSlice yaml.MapSlice
	if err := yaml.Unmarshal(destBytes, &destMapSlice); err != nil {
		return fn.NewE(err)
	}

	// Iterate over destMapSlice and update src
	for _, item := range destMapSlice {
		key := item.Key
		value := item.Value

		// Check if key already exists in src, and update it if it does
		found := false
		for i, srcItem := range *src {
			if srcItem.Key == key {
				(*src)[i] = yaml.MapItem{Key: key, Value: value}
				found = true
				break
			}
		}

		// If key doesn't exist in src, add it
		if !found {
			*src = append(*src, yaml.MapItem{Key: key, Value: value})
		}
	}

	return nil
}
