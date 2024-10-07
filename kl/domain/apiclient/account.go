package apiclient

import (
	"github.com/kloudlite/kl/pkg/functions"
)

type Team struct {
	Metadata    Metadata `json:"metadata"`
	DisplayName string   `json:"displayName"`
	Status      Status   `json:"status"`
}

func (apic *apiClient) ListTeams() ([]Team, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, functions.NewE(err)
	}

	respData, err := klFetch("cli_listAccounts", map[string]any{}, &cookie)
	if err != nil {
		return nil, functions.NewE(err)
	}

	type AccList []Team
	if fromResp, err := GetFromResp[AccList](respData); err != nil {
		return nil, functions.NewE(err)
	} else {
		return *fromResp, nil
	}
}

// func SelectTeam(teamName string) (*Team, error) {

// 	teams, err := ListTeams()
// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}

// 	if teamName != "" {
// 		for _, a := range teams {
// 			if a.Metadata.Name == teamName {
// 				return &a, nil
// 			}
// 		}
// 		return nil, functions.Error("you don't have access to this team")
// 	}

// 	team, err := fzf.FindOne(
// 		teams,
// 		func(team Team) string {
// 			return team.DisplayName
// 		},
// 		fzf.WithPrompt("Select Team > "),
// 	)

// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}

// 	return team, nil
// }

// func EnsureTeam(options ...fn.Option) (string, error) {
// 	teamName := fn.GetOption(options, "teamName")

// 	if teamName != "" {
// 		return teamName, nil
// 	}

// 	fc, err := fileclient.New()
// 	if err != nil {
// 		return "", functions.NewE(err)
// 	}

// 	s, _ := fc.CurrentTeamName()
// 	if s == "" {
// 		a, err := SelectTeam("")
// 		if err != nil {
// 			return "", functions.NewE(err)
// 		}

// 		return a.Metadata.Name, nil
// 	}

// 	return s, nil
// }
