package fileclient

import (
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func (f *fclient) CurrentAccountName() (string, error) {
	kt, err := f.getKlFile("")
	if err != nil {
		return "", functions.NewE(err)
	}

	if kt.AccountName == "" {
		return "", fn.Error("no account selected")
	}

	return kt.AccountName, nil
}
