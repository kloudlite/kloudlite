package lib

import (
	"io/ioutil"
	"kloudlite.io/cmd/internal/lib/server"
)

func CheckLogin() error {
	_, err := server.Me()
	if err != nil {
		return err
	}
	return nil
}

func Login(authToken string) error {
	return ioutil.WriteFile("/tmp/auth.token", []byte(authToken), 0644)
}
