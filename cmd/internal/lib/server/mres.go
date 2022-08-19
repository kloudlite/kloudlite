package server

import (
	"encoding/json"
)

type ResourceType struct {
	Id   string
	Name string
	// Outputs      map[string]string
	ResourceType string
}

type Mres struct {
	Id        string
	Name      string
	Source    string
	Resources []ResourceType
}

type Outputs []struct {
	Label string
	Name  string
}

type MresMarketItem struct {
	Active      bool
	DisplayName string `json:"display_name"`
	Name        string
	Resources   []struct {
		DisplayName string
		Name        string
		Outputs     Outputs
	}
}

type mCategory struct {
	Category    string
	DisplayName string `json:"display_name"`
	List        []MresMarketItem
}

type MresMarketCategories struct {
	Categories []mCategory
}

func GetMreses() ([]*Mres, []mCategory, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, nil, err
	}

	projectId, err := currentProjectId()
	if err != nil {
		return nil, nil, err
	}

	respData, err := gql(`
		query ManagedSvc_listInstallations($projectId: ID!) {
			managedSvc_listInstallations(projectId: $projectId) {
				id
				name
				source
				resources {
					id
					name
					resourceType
					status
				}
			}

	   managedSvc_marketList
		}
	`, map[string]any{
		"projectId": projectId,
	}, &cookie)

	if err != nil {
		return nil, nil, err
	}

	// cmd := exec.Command("jq", fmt.Sprintf("%q", string(respData)))
	// cmd.Stdout = os.Stdout
	// cmd.Run()
	// fmt.Println(string(respData))

	type Response struct {
		Data struct {
			ManagedSvc_listInstallations []*Mres               `json:"managedSvc_listInstallations"`
			ManagedSvc_marketList        *MresMarketCategories `json:"managedSvc_marketList"`
		} `json:"data"`
	}
	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data.ManagedSvc_listInstallations,
		resp.Data.ManagedSvc_marketList.Categories, nil
}
