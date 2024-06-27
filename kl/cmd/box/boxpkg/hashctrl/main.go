package hashctrl

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/kloudlite/kl/cmd/box/boxpkg/packagectrl"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
)

func keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func generateBoxHashContent(envName string, path string, klFile *client.KLFileType) ([]byte, error) {

	persistedConfig, err := generatePersistedEnv(klFile, envName, path)
	if err != nil {
		return nil, fn.NewE(err)
	}

	hash := md5.New()
	mountKeys := keys(persistedConfig.Mounts)
	slices.Sort(mountKeys)
	for _, v := range mountKeys {
		hash.Write([]byte(v))
		hash.Write([]byte(persistedConfig.Mounts[v]))
	}

	packages := keys(persistedConfig.PackageHashes)
	slices.Sort(packages)
	for _, v := range packages {
		hash.Write([]byte(v))
		hash.Write([]byte(persistedConfig.PackageHashes[v]))
	}

	envKeys := keys(persistedConfig.Env)
	slices.Sort(envKeys)
	for _, v := range envKeys {
		hash.Write([]byte(v))
		hash.Write([]byte(persistedConfig.Env[v]))
	}

	marshal, err := json.Marshal(map[string]any{
		"config": persistedConfig,
		"hash":   hash.Sum(nil),
	})
	if err != nil {
		return nil, functions.NewE(err)
	}

	return marshal, nil
}

func BoxHashFile(workspacePath string) (*PersistedEnv, error) {
	fileName, err := BoxHashFileName(workspacePath)
	if err != nil {
		return nil, functions.NewE(err)
	}
	configFolder, err := client.GetConfigFolder()
	if err != nil {
		return nil, functions.NewE(err)
	}
	filePath := path.Join(configFolder, "box-hash", fileName)
	data, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, functions.NewE(err)
	}
	if os.IsNotExist(err) {
		return nil, functions.NewE(err)
		// env, err := server.EnvAtPath(workspacePath)
		// if err != nil {
		// 	return nil, functions.Error(err)
		// }
		// if err = SyncBoxHash(env.Name, workspacePath); err != nil {
		// 	return nil, functions.Error(err)
		// }
		// return BoxHashFile(workspacePath)
	}
	var r struct {
		Config PersistedEnv `json:"config"`
		Hash   string       `json:"hash"`
	}

	if err = json.Unmarshal(data, &r); err != nil {
		return nil, fn.NewE(err)
	}
	return &r.Config, nil
}

func BoxHashFileName(path string) (string, error) {
	if os.Getenv("IN_DEV_BOX") == "true" {
		path = os.Getenv("KL_WORKSPACE")
	}

	hash := md5.New()
	if _, err := hash.Write([]byte(path)); err != nil {
		return "", nil
	}

	return fmt.Sprintf("hash-%x", hash.Sum(nil)), nil
}

func SyncBoxHash(fpath string) error {
	defer spinner.Client.UpdateMessage("updating lockfile")()

	klFile, err := client.GetKlFile(path.Join(fpath, "kl.yml"))
	if err != nil {
		return fn.NewE(err)
	}
	envName := ""
	e, err := client.EnvOfPath(fpath)
	if err != nil && err.Error() == "no selected environment" {
		envName = klFile.DefaultEnv
	} else if err != nil {
		return fn.NewE(err)
	} else {
		envName = e.Name
	}
	if envName == "" {
		return fn.Error("envName is required")
	}

	configFolder, err := client.GetConfigFolder()
	if err != nil {
		return fn.NewE(err)
	}

	boxHashFilePath, err := BoxHashFileName(fpath)
	if err != nil {
		return fn.NewE(err)
	}

	content, err := generateBoxHashContent(envName, fpath, klFile)
	if err != nil {
		return fn.NewE(err)
	}

	if err = os.MkdirAll(path.Join(configFolder, "box-hash"), 0755); err != nil {
		return fn.NewE(err)
	}

	if err = os.WriteFile(path.Join(configFolder, "box-hash", boxHashFilePath), content, 0644); err != nil {
		return fn.NewE(err)
	}

	return nil
}

func GenerateKLConfigHash(kf *client.KLFileType) (string, error) {
	klConfhash := md5.New()
	slices.SortFunc(kf.EnvVars, func(a, b client.EnvType) int {
		return strings.Compare(a.Key, b.Key)
	})
	for _, v := range kf.EnvVars {
		klConfhash.Write([]byte(v.Key))
		klConfhash.Write([]byte(func() string {
			if v.Value != nil {
				return *v.Value
			}
			return ""
		}()))
		klConfhash.Write([]byte(func() string {
			if v.ConfigRef != nil {
				return *v.ConfigRef
			}
			return ""
		}()))
		klConfhash.Write([]byte(func() string {
			if v.SecretRef != nil {
				return *v.SecretRef
			}
			return ""
		}()))
		klConfhash.Write([]byte(func() string {
			if v.MresRef != nil {
				return *v.MresRef
			}
			return ""
		}()))
	}
	slices.Sort(kf.Packages)
	for _, v := range kf.Packages {
		klConfhash.Write([]byte(v))
	}
	for _, v := range kf.Mounts {
		klConfhash.Write([]byte(v.Path))
		klConfhash.Write([]byte(func() string {
			if v.ConfigRef != nil {
				return *v.ConfigRef
			}
			return ""
		}()))
		klConfhash.Write([]byte(func() string {
			if v.SecretRef != nil {
				return *v.SecretRef
			}
			return ""
		}()))
		klConfhash.Write([]byte(v.Path))
	}
	return fmt.Sprintf("%x", klConfhash.Sum(nil)), nil
}

func generatePersistedEnv(kf *client.KLFileType, envName string, path string) (*PersistedEnv, error) {

	envs, mm, err := server.GetLoadMaps()
	if err != nil {
		return nil, functions.NewE(err)
	}

	realPkgs, err := packagectrl.SyncLockfileWithNewConfig(*kf)
	if err != nil {
		return nil, functions.NewE(err)
	}

	hashConfig := PersistedEnv{
		Packages:      kf.Packages,
		PackageHashes: realPkgs,
	}

	fm := map[string]string{}
	for _, fe := range kf.Mounts.GetMounts() {
		pth := fe.Path
		if pth == "" {
			pth = fe.Key
		}

		fm[pth] = mm[pth]
	}

	ev := map[string]string{}
	for k, v := range envs {
		ev[k] = v
	}

	for _, ne := range kf.EnvVars.GetEnvs() {
		ev[ne.Key] = ne.Value
	}

	e, err := client.EnvOfPath(path)
	if err != nil {
		return nil, functions.NewE(err)
	}

	ev["PURE_PROMPT_SYMBOL"] = fmt.Sprintf("(%s) %s", envName, "❯")
	ev["KL_SEARCH_DOMAIN"] = fmt.Sprintf("%s.svc.%s.local", e.TargetNs, e.ClusterName)

	klConfhash, err := GenerateKLConfigHash(kf)
	if err != nil {
		return nil, functions.NewE(err)
	}

	hashConfig.Env = ev
	hashConfig.Mounts = fm
	hashConfig.KLConfHash = klConfhash
	return &hashConfig, nil
}
