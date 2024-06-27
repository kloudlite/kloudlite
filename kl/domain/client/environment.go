package client

import (
	"fmt"
	"net"
	"os"

	"github.com/kloudlite/kl/pkg/functions"
)

func CheckPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}

func SelectEnv(ev Env) error {
	k, err := GetExtraData()
	if err != nil {
		return functions.NewE(err)
	}

	dir, err := os.Getwd()
	if err != nil {
		return functions.NewE(err)
	}

	if InsideBox() {
		dir = os.Getenv("KL_WORKSPACE")
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
		return functions.NewE(err)
	}

	if k.SelectedEnvs == nil {
		k.SelectedEnvs = map[string]*Env{}
	}

	k.SelectedEnvs[pth] = &ev

	return SaveExtraData(k)
}

func EnvOfPath(pth string) (*Env, error) {
	c, err := GetExtraData()
	if err != nil {
		return nil, functions.NewE(err)
	}

	if c.SelectedEnvs == nil {
		return nil, functions.Error("no selected environment")
	}

	if c.SelectedEnvs[pth] == nil {
		return nil, functions.Error("no selected environment")
	}

	return c.SelectedEnvs[pth], nil
}

func CurrentEnv() (*Env, error) {

	c, err := GetExtraData()
	if err != nil {
		return nil, functions.NewE(err)
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, functions.NewE(err)
	}

	if InsideBox() {
		dir = os.Getenv("KL_WORKSPACE")
	}

	if c.SelectedEnvs == nil {
		return nil, functions.Error("no selected environment")
	}

	if c.SelectedEnvs[dir] == nil {
		return nil, functions.Error("no selected environment")
	}

	return c.SelectedEnvs[dir], nil
}
