package client

import "os"

func InsideBox() bool {
	s, ok := os.LookupEnv("IN_DEV_BOX")
	if !ok {
		return false
	}

	return s == "true"
}
