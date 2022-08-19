package domain

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"kloudlite.io/apps/console/internal/domain/entities"
	op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
	"strings"
)

func (d *domain) OnUpdateProject(ctx context.Context, response *op_crds.StatusUpdate) error {
	one, err := d.projectRepo.FindOne(ctx, repos.Filter{
		"id": response.Metadata.ResourceId,
		//"cluster_id": response.ClusterId,
	})
	if err != nil {
		return err
	}
	if one == nil {
		// Ignore unknown project
		return nil
	}
	if response.IsReady {
		one.Status = entities.ProjectStateLive
	} else {
		one.Status = entities.ProjectStateSyncing
	}
	one.Conditions = response.ChildConditions
	_, err = d.projectRepo.UpdateById(ctx, one.Id, one)
	return err
}

func (d *domain) GetProjectWithID(ctx context.Context, projectId repos.ID) (*entities.Project, error) {
	id, err := d.projectRepo.FindById(ctx, projectId)
	return id, err
}
func (d *domain) GetAccountProjects(ctx context.Context, acountId repos.ID) ([]*entities.Project, error) {
	res, err := d.projectRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"account_id": acountId,
		},
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *domain) InviteProjectMember(ctx context.Context, projectID repos.ID, email string, role string) (bool, error) {
	byEmail, err := d.authClient.EnsureUserByEmail(ctx, &auth.GetUserByEmailRequest{Email: email})
	if err != nil {
		return false, err
	}
	if byEmail == nil {
		return false, errors.New("user not found")
	}
	_, err = d.iamClient.InviteMembership(ctx, &iam.InAddMembership{
		UserId:       byEmail.UserId,
		ResourceType: "project",
		ResourceId:   string(projectID),
		Role:         role,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}
func (d *domain) RemoveProjectMember(ctx context.Context, projectId repos.ID, userId repos.ID) error {
	_, err := d.iamClient.RemoveMembership(ctx, &iam.InRemoveMembership{
		UserId:     string(userId),
		ResourceId: string(projectId),
	})
	if err != nil {
		return err
	}
	return nil
}
func (d *domain) GetProjectMemberships(ctx context.Context, projectID repos.ID) ([]*entities.ProjectMembership, error) {
	rbs, err := d.iamClient.ListResourceMemberships(ctx, &iam.InResourceMemberships{
		ResourceId:   string(projectID),
		ResourceType: string(common.ResourceProject),
	})
	if err != nil {
		return nil, err
	}
	var memberships []*entities.ProjectMembership
	for _, rb := range rbs.RoleBindings {
		memberships = append(memberships, &entities.ProjectMembership{
			ProjectId: repos.ID(rb.ResourceId),
			UserId:    repos.ID(rb.UserId),
			Role:      common.Role(rb.Role),
		})
	}
	if err != nil {
		return nil, err
	}
	return memberships, nil
}

func (d *domain) CreateProject(ctx context.Context, ownerId repos.ID, accountId repos.ID, projectName string, displayName string, logo *string, regionId *repos.ID, description *string) (*entities.Project, error) {
	create, err := d.projectRepo.Create(ctx, &entities.Project{
		Name:        projectName,
		AccountId:   accountId,
		ReadableId:  repos.ID(generateReadable(projectName)),
		DisplayName: displayName,
		Logo:        logo,
		Description: description,
		RegionId:    regionId,
		Status:      entities.ProjectStateSyncing,
	})
	if err != nil {
		return nil, err
	}
	_, err = d.iamClient.AddMembership(ctx, &iam.InAddMembership{
		UserId:       string(ownerId),
		ResourceType: "project",
		ResourceId:   string(create.Id),
		Role:         "owner",
	})
	if err != nil {
		return nil, err
	}
	err = d.workloadMessenger.SendAction("apply", string(create.Id), &op_crds.Project{
		APIVersion: op_crds.APIVersion,
		Kind:       op_crds.ProjectKind,
		Metadata: op_crds.ProjectMetadata{
			Name: create.Name,
			Labels: map[string]string{
				"kloudlite.io/account-ref":  string(create.AccountId),
				"kloudlite.io/resource-ref": string(create.Id),
			},
			Annotations: map[string]string{
				"kloudlite.io/account-ref":  string(create.AccountId),
				"kloudlite.io/resource-ref": string(create.Id),
			},
		},
		Spec: op_crds.ProjectSpec{
			DisplayName: displayName,
		},
	})
	if err != nil {
		return nil, err
	}
	return create, err
}

func (d *domain) DeleteProject(ctx context.Context, id repos.ID) (bool, error) {
	proj, err := d.projectRepo.FindById(ctx, id)
	if err != nil {
		return false, err
	}
	proj.IsDeleting = true
	_, err = d.projectRepo.UpdateById(ctx, id, proj)
	if err != nil {
		return false, err
	}
	err = d.workloadMessenger.SendAction("delete", string(id), &op_crds.Project{
		APIVersion: op_crds.APIVersion,
		Kind:       op_crds.ProjectKind,
		Metadata: op_crds.ProjectMetadata{
			Name: proj.Name,
		},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) OnDeleteProject(ctx context.Context, response *op_crds.StatusUpdate) error {
	return d.projectRepo.DeleteById(ctx, repos.ID(response.Metadata.ResourceId))
}

func (d *domain) getProjectRegionDetails(ctx context.Context, proj *entities.Project) (cloudProvider string, region string, err error) {
	var projectRegion *entities.EdgeRegion
	var projectCloudProvider *entities.CloudProvider
	if proj.RegionId != nil {
		projectRegion, err = d.regionRepo.FindById(ctx, *proj.RegionId)
		if err != nil {
			return "", "", err
		}
		projectCloudProvider, err = d.providerRepo.FindById(ctx, projectRegion.ProviderId)
		if err != nil {
			return "", "", err
		}
	}
	return projectCloudProvider.Provider, projectRegion.Region, nil
}

func (d *domain) GetDockerCredentials(ctx context.Context, projectId repos.ID) (username string, password string, err error) {
	project, err := d.projectRepo.FindById(ctx, projectId)
	if err != nil {
		return "", "", err
	}
	secret, err := d.kubeCli.GetSecret(ctx, project.Name, "kloudlite-docker-registry")
	if err != nil {
		return "", "", err
	}
	var data struct {
		Auths map[string]struct {
			Auth string `json:"auth"`
		} `json:"auths"`
	}
	err = json.Unmarshal(secret.Data[".dockerconfigjson"], &data)
	if err != nil {
		return "", "", nil
	}
	connectionStr := data.Auths["harbor.dev.madhouselabs.io"].Auth
	decodeString, err := base64.StdEncoding.DecodeString(connectionStr)
	if err != nil {
		return "", "", err
	}
	splits := strings.Split(string(decodeString), ":")
	return splits[0], splits[1], nil
}
