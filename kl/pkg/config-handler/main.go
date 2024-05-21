package confighandler

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/kloudlite/kl/pkg/ui/text"
	yaml "gopkg.in/yaml.v2"
)

var KlFileNotExists = fmt.Errorf(text.Colored("please ensure kl.yaml file by running \"kl init\" command in your workspace root.", 0))

func ReadConfig[T any](path string) (*T, error) {
	var v T
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s, %s", err.Error(), text.Colored("please ensure kl.yaml file by running \"kl init\" command in your workspace root.", 0))
	}
	if err := yaml.Unmarshal(b, &v); err != nil {
		return nil, err
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
				return err
			}

			return os.WriteFile(path, b, perm)
		}

		return err
	}

	var config yaml.MapSlice
	if err := yaml.Unmarshal(b, &config); err != nil {
		return err
	}

	if err := fillConfig(&config, v); err != nil {
		return err
	}

	b, err = yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, perm)
}

// FillConfig updates src with values from dest
func fillConfig(src *yaml.MapSlice, dest interface{}) error {
	// Marshal dest to YAML
	destBytes, err := yaml.Marshal(dest)
	if err != nil {
		return err
	}

	// Unmarshal dest YAML into a MapSlice
	var destMapSlice yaml.MapSlice
	if err := yaml.Unmarshal(destBytes, &destMapSlice); err != nil {
		return err
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
