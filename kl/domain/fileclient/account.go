package fileclient

import (
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func (f *fclient) CurrentTeamName() (string, error) {
	kt, err := f.getKlFile("")
	if err != nil {
		return "", functions.NewE(err)
	}

	if kt.TeamName == "" {
		return "", fn.Error("no team selected")
	}

	return kt.TeamName, nil
}
