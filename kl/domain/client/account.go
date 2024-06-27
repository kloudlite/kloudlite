package client

import (
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func CurrentAccountName() (string, error) {

	kt, err := GetKlFile("")
	if err != nil {
		return "", functions.NewE(err)
	}

	if kt.AccountName == "" {
		return "", fn.Error("no account selected")
	}

	return kt.AccountName, nil
}
