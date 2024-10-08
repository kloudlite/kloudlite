package fileclient

import (
	"errors"
	confighandler "github.com/kloudlite/kl/pkg/config-handler"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func (f *fclient) CurrentTeamName() (string, error) {
	kt, err := f.getKlFile("")
	if err != nil {
		if errors.Is(err, confighandler.ErrKlFileNotExists) {
			extraData, err := GetExtraData()
			if err != nil {
				return "", fn.NewE(err)
			}
			return extraData.SelectedTeam, nil
		}
		return "", fn.NewE(err)
	}

	if kt.TeamName == "" {
		return "", fn.Error("no team selected")
	}

	return kt.TeamName, nil
}
