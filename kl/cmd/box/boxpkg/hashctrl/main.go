package hashctrl

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/kloudlite/kl/cmd/box/boxpkg/packagectrl"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/envclient"
	"github.com/kloudlite/kl/domain/fileclient"
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

func generateBoxHashContent(envName string, fpath string, klFile *fileclient.KLFileType) ([]byte, error) {

	persistedConfig, err := generatePersistedEnv(klFile, envName, fpath)
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

	hsh := fmt.Sprintf("%x", hash.Sum(nil))
	marshal, err := json.Marshal(map[string]any{
		"config": persistedConfig,
		"hash":   hsh,
	})

	if err != nil {
		return nil, fn.NewE(err)
	}

	if envclient.InsideBox() {
		if err := os.WriteFile("/tmp/hash", []byte(hsh), os.ModePerm); err != nil {
			return nil, fn.NewE(err)
		}
	}

	return marshal, nil
}

func BoxHashFile(workspacePath string) (*PersistedEnv, error) {
	fileName, err := BoxHashFileName(workspacePath)
	if err != nil {
		return nil, fn.NewE(err)
	}
	configFolder, err := fileclient.GetConfigFolder()
	if err != nil {
		return nil, fn.NewE(err)
	}
	filePath := path.Join(configFolder, "box-hash", fileName)
	data, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fn.NewE(err)
	}
	if os.IsNotExist(err) {
		return nil, fn.NewE(err)
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
	hash := md5.New()
	if _, err := hash.Write([]byte(path)); err != nil {
		return "", nil
	}

	return fmt.Sprintf("hash-%x", hash.Sum(nil)), nil
}

func SyncBoxHash(fpath string) error {
	defer spinner.Client.UpdateMessage("validating kl.yml and kl.lock")()

	fc, err := fileclient.New()
	if err != nil {
		return fn.NewE(err)
	}

	klFile, err := fc.GetKlFile(path.Join(fpath, "kl.yml"))
	if err != nil {
		return fn.NewE(err)
	}
	envName := ""

	pathKey := fpath

	e, err := fileclient.EnvOfPath(pathKey)
	if err != nil && errors.Is(err, fileclient.NoEnvSelected) {
		envName = klFile.DefaultEnv
	} else if err != nil {
		return fn.NewE(err)
	} else {
		envName = e.Name
	}
	if envName == "" {
		return fn.Error("envName is required")
	}

	configFolder, err := fileclient.GetConfigFolder()
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

func GenerateKLConfigHash(kf *fileclient.KLFileType) (string, error) {
	defer spinner.Client.UpdateMessage("validating kl.yml and parsing environment variables")()

	klConfhash := md5.New()
	slices.SortFunc(kf.EnvVars, func(a, b fileclient.EnvType) int {
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

func generatePersistedEnv(kf *fileclient.KLFileType, envName string, path string) (*PersistedEnv, error) {
	envs, mm, err := apiclient.GetLoadMaps()
	if err != nil {
		return nil, fn.NewE(err)
	}

	realPkgs, err := packagectrl.SyncLockfileWithNewConfig(*kf)
	if err != nil {
		return nil, fn.NewE(err)
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

	e, err := fileclient.EnvOfPath(path)
	if err != nil {
		return nil, fn.NewE(err)
	}
	ev["PURE_PROMPT_SYMBOL"] = fmt.Sprintf("(%s) %s", envName, "❯")
	ev["KL_SEARCH_DOMAIN"] = fmt.Sprintf("%s.svc.%s.local", e.TargetNs, e.ClusterName)

	klConfhash, err := GenerateKLConfigHash(kf)
	if err != nil {
		return nil, fn.NewE(err)
	}

	hashConfig.Env = ev
	hashConfig.Mounts = fm
	hashConfig.KLConfHash = klConfhash
	return &hashConfig, nil
}
