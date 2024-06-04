package client

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/kloudlite/kl/klbox-docker/devboxfile"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type ResEnvType struct {
	Name   string `json:"name" yaml:"name"`
	Key    string `json:"key"`
	RefKey string `json:"refKey" yaml:"refKey"`
}

type EnvType struct {
	Key       string  `json:"key" yaml:"key"`
	Value     *string `json:"value,omitempty" yaml:"value,omitempty"`
	ConfigRef *string `json:"configRef,omitempty" yaml:"configRef,omitempty"`
	SecretRef *string `json:"secretRef,omitempty" yaml:"secretRef,omitempty"`
	MresRef   *string `json:"mresRef,omitempty" yaml:"mresRef,omitempty"`
}

type ResType struct {
	Name string       `json:"name"`
	Env  []ResEnvType `json:"env"`
}

type EnvVars []EnvType

type NormalEnv struct {
	Key   string
	Value string
}

func (e *EnvVars) GetEnvs() []NormalEnv {
	resp := make([]NormalEnv, 0)
	if e == nil {
		return resp
	}

	for _, r := range *e {
		if r.Value != nil {
			resp = append(resp, NormalEnv{
				Key:   r.Key,
				Value: *r.Value,
			})
		}
	}

	return resp
}

type resType string

const (
	Res_config resType = "config"
	Res_secret resType = "secret"
	Res_mres   resType = "mres"
)

func (e *EnvVars) getReses(res resType) []ResType {

	resp := make([]ResType, 0)
	if e == nil {
		return resp
	}

	hist := map[string]int{}

	for _, r := range *e {
		var ref *string

		switch res {
		case Res_config:
			ref = r.ConfigRef
		case Res_secret:
			ref = r.SecretRef
		case Res_mres:
			ref = r.MresRef
		default:
			continue
		}

		if ref == nil {
			continue
		}

		s := strings.Split(*ref, "/")
		if len(s) != 2 {
			continue
		}

		mName, mKey := s[0], s[1]

		j, ok := hist[mName]
		if !ok {
			hist[mName] = len(resp)
			resp = append(resp, ResType{
				Name: mName,
				Env: []ResEnvType{
					{
						Key:    r.Key,
						RefKey: mKey,
					},
				},
			})
			continue
		}

		resp[j].Env = append(resp[j].Env, ResEnvType{
			Key:    r.Key,
			RefKey: mKey,
		})
	}

	return resp
}

func (e *EnvVars) GetMreses() []ResType {
	return e.getReses(Res_mres)
}

func (e *EnvVars) GetConfigs() []ResType {
	return e.getReses(Res_config)
}

func (e *EnvVars) GetSecrets() []ResType {
	return e.getReses(Res_secret)
}

func (e *EnvVars) AddResTypes(rt []ResType, rtype resType) {

	if e == nil {
		e = &EnvVars{}
	}

	keys := map[string]bool{}

	getEnvKey := func(r EnvType) string {
		return fmt.Sprint(r.Key, func() string {
			if r.SecretRef != nil {
				return *r.SecretRef
			}
			if r.MresRef != nil {
				return *r.MresRef
			}
			if r.SecretRef != nil {
				return *r.SecretRef
			}
			if r.Value != nil {
				return *r.Value
			}

			return ""
		}())
	}

	getRtKey := func(name, key, refKey string) string {
		return fmt.Sprint(key, name, "/", refKey)
	}

	for _, r := range *e {
		ek := getEnvKey(r)

		if !keys[ek] {
			keys[ek] = true
		}
	}

	appendEnv := func(key, name, refKey string) {
		*e = append(*e, EnvType{
			Key:   key,
			Value: nil,
			ConfigRef: func() *string {
				if rtype != Res_config {
					return nil
				}

				return fn.Ptr(fmt.Sprint(name, "/", refKey))
			}(),
			SecretRef: func() *string {
				if rtype != Res_secret {
					return nil
				}

				return fn.Ptr(fmt.Sprint(name, "/", refKey))
			}(),
			MresRef: func() *string {
				if rtype != Res_mres {
					return nil
				}

				return fn.Ptr(fmt.Sprint(name, "/", refKey))
			}(),
		})
	}

	for _, r := range rt {
		for _, ret := range r.Env {
			ek := getRtKey(r.Name, ret.Key, ret.RefKey)
			if !keys[ek] {
				keys[ek] = true
				appendEnv(ret.Key, r.Name, ret.RefKey)
			}
		}
	}
}

func SyncDevboxShellEnvFile() error {
	if !InsideBox() {
		return nil
	}

	devBoxDir := filepath.Dir(DEVBOX_JSON_PATH)

	command := exec.Command("devbox", "shellenv")
	command.Dir = devBoxDir

	out, err := command.Output()
	if err != nil {
		return err
	}

	return os.WriteFile(path.Join(devBoxDir, "devbox-env.sh"), out, os.ModePerm)
}

/*
Steps performed in ExecPackageCommand:
1. Sync devbox.lock with kl.lock
2. Update devbox.json by combining kl.json
3. Run devbox command
4. Copy devbox.json to workspace/kl.json
5. Update workspace/kl.lock
*/

func ExecPackageCommand(cmd string) error {
	if !InsideBox() {
		return nil
	}
	defer syncDevboxLock()()

	devboxContext := devboxfile.DevboxConfig{}

	devboxJsonConfig, err := os.ReadFile(DEVBOX_JSON_PATH)
	if err == nil {
		if err := devboxContext.ParseJson(devboxJsonConfig); err != nil {
			return err
		}
	}

	klContext, err := GetKlFile("")
	if err != nil {
		return err
	}

	devboxContext.Packages = klContext.Packages

	devboxConfig, err := devboxContext.ToJson()
	if err != nil {
		return err
	}

	if err := os.WriteFile(DEVBOX_JSON_PATH, devboxConfig, os.ModePerm); err != nil {
		return err
	}

	r := csv.NewReader(strings.NewReader(cmd))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return err
	}

	command := exec.Command(cmdArr[0], cmdArr[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Dir = filepath.Dir(DEVBOX_JSON_PATH)

	if err = command.Run(); err != nil {
		return err
	}

	b, err := os.ReadFile(DEVBOX_JSON_PATH)
	if err != nil {
		return err
	}

	if err := devboxContext.ParseJson(b); err != nil {
		return err
	}

	klContext.Packages = devboxContext.Packages

	if err := WriteKLFile(*klContext); err != nil {
		return err
	}

	return SyncDevboxShellEnvFile()
}

func syncDevboxLock() func() {
	if err := fn.CopyFile(KL_LOCK_PATH, DEVBOX_LOCK_PATH); err != nil {
		fn.Warn(err)
	}

	return func() {
		if err := fn.CopyFile(DEVBOX_LOCK_PATH, KL_LOCK_PATH); err != nil {
			fn.Warn(err)
		}
	}
}
