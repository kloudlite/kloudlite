package client

import (
	"errors"
	"strings"
)

func CurrentProjectName() (string, error) {
	returnErr :=
		errors.New("can't get current project from you kl file. please initialize your project using \"kl init\" first.")

	kfile, err := GetKlFile(nil)
	if err != nil {
		return "", returnErr
	}

	if kfile.Project == "" {
		return "", returnErr
	}

	s := strings.Split(kfile.Project, "/")

	if len(s) != 2 {
		return "", returnErr
	}

	return s[1], nil
}
