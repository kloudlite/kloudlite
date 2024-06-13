package client

import (
	"errors"
	"fmt"
	"net"
	"os"
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

func GetAvailablePort() (int, error) {
	for i := 12300; i <= 12400; i++ {
		if CheckPortAvailable(i) {
			return i, nil
		}
	}
	// 61100, 61300
	return 0, fmt.Errorf("no ports available to use")
}

func SelectEnv(ev Env) error {
	k, err := GetExtraData()
	if err != nil {
		return err
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
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

func SelectEnvOnPath(ev Env, pth string) error {
	k, err := GetExtraData()
	if err != nil {
		return err
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
		return nil, err
	}

	if c.SelectedEnvs == nil {
		return nil, errors.New("no selected environment")
	}

	if c.SelectedEnvs[pth] == nil {
		return nil, errors.New("no selected environment")
	}

	return c.SelectedEnvs[pth], nil
}

func CurrentEnv() (*Env, error) {

	c, err := GetExtraData()
	if err != nil {
		return nil, err
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if InsideBox() {
		dir = os.Getenv("KL_WORKSPACE")
	}

	if c.SelectedEnvs == nil {
		return nil, errors.New("no selected environment")
	}

	if c.SelectedEnvs[dir] == nil {
		return nil, errors.New("no selected environment")
	}

	return c.SelectedEnvs[dir], nil
}
