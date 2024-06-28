package fileclient

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/domain/envclient"
	fn "github.com/kloudlite/kl/pkg/functions"
)

var NoEnvSelected = fmt.Errorf("no selected environment")

func SelectEnv(ev Env) error {
	k, err := GetExtraData()
	if err != nil {
		return fn.NewE(err)
	}

	dir, err := os.Getwd()
	if err != nil {
		return fn.NewE(err)
	}

	if envclient.InsideBox() {
		s, err := envclient.GetWorkspacePath()
		if err != nil {
			return fn.NewE(err)
		}

		dir = s
	}

	if k.SelectedEnvs == nil {
		k.SelectedEnvs = map[string]*Env{}
	}

	k.SelectedEnvs[dir] = &ev

	return SaveExtraData(k)
}

func SelectEnvOnPath(pth string, ev Env) error {
	k, err := GetExtraData()
	if err != nil {
		return fn.NewE(err)
	}

	if k.SelectedEnvs == nil {
		k.SelectedEnvs = map[string]*Env{}
	}

	k.SelectedEnvs[pth] = &ev

	return SaveExtraData(k)
}

func EnvOfPath(pth string) (*Env, error) {
	if envclient.InsideBox() {
		s, err := envclient.GetWorkspacePath()
		if err != nil {
			return nil, fn.NewE(err)
		}

		pth = s
	}

	c, err := GetExtraData()
	if err != nil {
		return nil, fn.NewE(err)
	}

	if c.SelectedEnvs == nil || c.SelectedEnvs[pth] == nil {
		return nil, fn.NewE(NoEnvSelected)
	}

	return c.SelectedEnvs[pth], nil
}

func CurrentEnv() (*Env, error) {
	c, err := GetExtraData()
	if err != nil {
		return nil, fn.NewE(err)
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, fn.NewE(err)
	}

	if envclient.InsideBox() {
		s, err := envclient.GetWorkspacePath()
		if err != nil {
			return nil, fn.NewE(err)
		}

		dir = s
	}

	if c.SelectedEnvs == nil {
		return nil, fn.Error("no selected environment")
	}

	if c.SelectedEnvs[dir] == nil {
		return nil, fn.Error("no selected environment")
	}

	return c.SelectedEnvs[dir], nil
}
