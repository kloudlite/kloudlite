package server

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"os"
	"path"
	"slices"
)

func keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func BoxHashFilePath() string {
	cwd, err := os.Getwd()
	if err != nil {
		fn.PrintError(err)
		return ""
	}
	hash := md5.New()
	hash.Write([]byte(cwd))
	boxHashFilePath := fmt.Sprintf("hash-%x", hash.Sum(nil))
	return boxHashFilePath
}

func generateBoxHashContent() ([]byte, error) {
	config, err := LoadDevboxConfig()
	if err != nil {
		return nil, err
	}
	hash := md5.New()
	hash.Write([]byte(config.KlConfig.Dns))
	mountKeys := keys(config.KlConfig.Mounts)
	slices.Sort(mountKeys)
	for _, v := range mountKeys {
		hash.Write([]byte(v))
		hash.Write([]byte(config.KlConfig.Mounts[v]))
	}
	packages := keys(config.PackageHashes)
	slices.Sort(packages)
	for _, v := range packages {
		hash.Write([]byte(v))
		hash.Write([]byte(config.PackageHashes[v]))
	}
	envKeys := keys(config.Env)
	slices.Sort(envKeys)
	for _, v := range envKeys {
		hash.Write([]byte(v))
		hash.Write([]byte(config.Env[v]))
	}
	marshal, err := json.Marshal(map[string]any{
		"config": config,
		"hash":   hash.Sum(nil),
	})
	if err != nil {
		return nil, err
	}
	return marshal, nil
}

func currentWorkingDir() string {
	if os.Getenv("KL_WORKSPACE") != "" {
		return os.Getenv("KL_WORKSPACE")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}

func SyncBoxHash() error {
	fn.Printf("environments may have been updated. to reflect the changes, do you want to restart the container? Y/n`.")
	if !fn.Confirm("Y", "Y") {
		return nil
	}
	configFolder, err := client.GetConfigFolder()
	cwd := currentWorkingDir()
	hash := md5.New()
	hash.Write([]byte(cwd))
	boxHashFilePath := fmt.Sprintf("hash-%x", hash.Sum(nil))
	if err != nil {
		return err
	}
	content, err := generateBoxHashContent()
	if err != nil {
		return err
	}
	err = os.WriteFile(path.Join(configFolder, "box-hash", boxHashFilePath), content, 0644)
	if err != nil {
		return err
	}
	return nil
}

func EnsureBoxHash() error {
	if err := ensureBoxHashFolder(); err != nil {
		return err
	}
	cwd := currentWorkingDir()
	// check if kl.yml exists in cwd
	klFile := path.Join(cwd, "kl.yml")
	if _, err := os.Stat(klFile); err != nil {
		return err
	}
	err := SyncBoxHash()
	if err != nil {
		return err
	}
	return nil
}

func ensureBoxHashFolder() error {
	configFolder, err := client.GetConfigFolder()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(path.Join(configFolder, "box-hash"), 0755); err != nil {
		return err
	}
	return nil
}
