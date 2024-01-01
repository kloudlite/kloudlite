package server

import (
	"encoding/json"
	"errors"

	"github.com/kloudlite/kl/lib"
	"github.com/kloudlite/kl/lib/util"
)

type Project struct {
	Id          string `json:"id"`
	ReadableId  string `json:"readableId"`
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func CurrentProjectId() (string, error) {

	file, err := util.GetContextFile()

	if err != nil {
		return "", err
	}

	if file.ProjectId == "" {
		return "",
			errors.New("no project is selected yet. please select one using \"kl use project\"")
	}

	return file.ProjectId, nil
}

func ListProjects(options ...lib.Option) ([]Project, error) {
	accountName := common.GetOption(options, "accountName")

	cookie, err := getCookie()

	if err != nil {
		return nil, err
	}

	if accountName == "" {
		accountName, err = util.CurrentAccountName()

		if err != nil {
			return nil, err
		}
	}

	respData, err := klFetch("cli_listProjects", map[string]any{
		"cId": accountName,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	type Response struct {
		FinanceAccount struct {
			Projects []Project `json:"projects"`
		} `json:"data"`
	}
	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}
	return resp.FinanceAccount.Projects, nil
}
