package server

import (
	"encoding/json"
	"fmt"
)

type ResourceType struct {
	Id           string
	Name         string
	Outputs      map[string]string
	ResourceType string
}

type Mres struct {
	Id        string
	Name      string
	Source    string
	Resources []ResourceType
}

func GetMreses() ([]*Mres, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	projectId, err := currentProjectId()
	if err != nil {
		return nil, err
	}

	respData, err := gql(`
		query ManagedSvc_listInstallations($projectId: ID!) {
			managedSvc_listInstallations(projectId: $projectId) {
				id
				name
				outputs
				source
				status
				updatedAt
				values
				createdAt
				resources {
					createdAt
					id
					name
					outputs
					resourceType
					status
					updatedAt
					values
				}
			}
		}
	`, map[string]any{
		"projectId": projectId,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	fmt.Println(string(respData))

	type Response struct {
		Data struct {
			ManagedSvc_listInstallations []*Mres `json:"managedSvc_listInstallations"`
		} `json:"data"`
	}
	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Data.ManagedSvc_listInstallations, nil
}
