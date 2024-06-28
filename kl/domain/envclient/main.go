package envclient

import (
	"os"

	"github.com/kloudlite/kl/pkg/functions"
)

func GetWorkspacePath() (string, error) {
	if InsideBox() {
		s, ok := os.LookupEnv("KL_WORKSPACE")
		if !ok {
			return "", functions.Error("KL_WORKSPACE is not set")
		}

		return s, nil
	}

	return os.Getwd()
}

func InsideBox() bool {
	s, ok := os.LookupEnv("IN_DEV_BOX")
	if !ok {
		return false
	}

	return s == "true"
}
