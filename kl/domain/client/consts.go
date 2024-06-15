package client

import (
	"fmt"
	"os"
)

const (
	KL_LOCK_PATH = "/home/kl/workspace/kl.lock"
	// DEVBOX_LOCK_PATH = "/kl-tmp/devbox/devbox.lock"
	// DEVBOX_JSON_PATH = "/kl-tmp/devbox/devbox.json"
)

func DevBoxLockPath() string {
	s, b := os.LookupEnv("KL_TMP_PATH")
	if b {
		return fmt.Sprintf("%s/devbox/devbox.lock", s)
	}

	return KL_LOCK_PATH
}

func DevBoxJsonPath() string {
	s, b := os.LookupEnv("KL_TMP_PATH")
	if b {
		return fmt.Sprintf("%s/devbox/devbox.json", s)
	}

	return KL_LOCK_PATH
}
