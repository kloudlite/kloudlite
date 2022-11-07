package domain

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"kloudlite.io/pkg/kubeapi"

	"kloudlite.io/apps/console/internal/domain/entities"
	op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

func (d *domain) OnUpdateProject(ctx context.Context, response *op_crds.StatusUpdate) error {
	one, err := d.projectRepo.FindById(ctx, repos.ID(response.Metadata.ResourceId))
	if err = mongoError(err, "managed resource not found"); err != nil {
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

	if err := d.checkProjectAccess(ctx, projectId, ReadProject); err != nil {
		return nil, err
	}

	id, err := d.projectRepo.FindById(ctx, projectId)
	return id, err
}

func (d *domain) GetAccountProjects(ctx context.Context, acountId repos.ID) ([]*entities.Project, error) {
	if err := d.checkAccountAccess(ctx, acountId, ReadProject); err != nil {
		return nil, err
	}

	res, err := d.projectRepo.Find(
		ctx, repos.Query{
			Filter: repos.Filter{
				"account_id": acountId,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (d *domain) InviteProjectMember(ctx context.Context, projectID repos.ID, email string, role string) (bool, error) {

	var err error
	switch role {
	case "project-owner":
		err = d.checkProjectAccess(ctx, projectID, "invite_proj_owner")
	case "project-admin":
		err = d.checkProjectAccess(ctx, projectID, "invite_proj_admin")
	case "project-member":
		err = d.checkProjectAccess(ctx, projectID, "invite_proj_member")
	}
	if err != nil {
		return false, err
	}

	byEmail, err := d.authClient.EnsureUserByEmail(ctx, &auth.GetUserByEmailRequest{Email: email})
	if err != nil {
		return false, err
	}

	if byEmail == nil {
		return false, errors.New("user not found")
	}
	_, err = d.iamClient.InviteMembership(
		ctx, &iam.InAddMembership{
			UserId:       byEmail.UserId,
			ResourceType: "project",
			ResourceId:   string(projectID),
			Role:         role,
		},
	)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (d *domain) RemoveProjectMember(ctx context.Context, projectId repos.ID, userId repos.ID) error {

	if err := d.checkProjectAccess(ctx, projectId, "cancel_proj_invite"); err != nil {
		return err
	}

	_, err := d.iamClient.RemoveMembership(
		ctx, &iam.InRemoveMembership{
			UserId:     string(userId),
			ResourceId: string(projectId),
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func (d *domain) GetProjectMemberships(ctx context.Context, projectID repos.ID) ([]*entities.ProjectMembership, error) {

	if err := d.checkProjectAccess(ctx, projectID, ReadProject); err != nil {
		return nil, err
	}

	rbs, err := d.iamClient.ListResourceMemberships(
		ctx, &iam.InResourceMemberships{
			ResourceId:   string(projectID),
			ResourceType: string(common.ResourceProject),
		},
	)
	if err != nil {
		return nil, err
	}

	var memberships []*entities.ProjectMembership
	for _, rb := range rbs.RoleBindings {
		memberships = append(
			memberships, &entities.ProjectMembership{
				ProjectId: repos.ID(rb.ResourceId),
				UserId:    repos.ID(rb.UserId),
				Role:      common.Role(rb.Role),
			},
		)
	}
	if err != nil {
		return nil, err
	}

	return memberships, nil
}

func (d *domain) CreateProject(ctx context.Context, ownerId repos.ID, accountId repos.ID, projectName string, displayName string, logo *string, regionId *repos.ID, description *string) (*entities.Project, error) {
	if err := d.checkAccountAccess(ctx, accountId, "create_project"); err != nil {
		return nil, err
	}

	project, err := d.projectRepo.Create(
		ctx, &entities.Project{
			Name:        projectName,
			AccountId:   accountId,
			ReadableId:  repos.ID(generateReadable(projectName)),
			DisplayName: displayName,
			Logo:        logo,
			Description: description,
			RegionId:    regionId,
			Status:      entities.ProjectStateSyncing,
		},
	)
	if err != nil {
		return nil, err
	}

	_, err = d.iamClient.AddMembership(
		ctx, &iam.InAddMembership{
			UserId:       string(ownerId),
			ResourceType: "project",
			ResourceId:   string(project.Id),
			Role:         "project-admin",
		},
	)
	if err != nil {
		return nil, err
	}

	clusterId, err := d.getClusterForAccount(ctx, accountId)
	if err != nil {
		return nil, err
	}

	err = d.workloadMessenger.SendAction(
		"apply", d.getDispatchKafkaTopic(clusterId), string(project.Id), &op_crds.Project{
			APIVersion: op_crds.APIVersion,
			Kind:       op_crds.ProjectKind,
			Metadata: op_crds.ProjectMetadata{
				Name: project.Name,
				Labels: map[string]string{
					"kloudlite.io/account-ref":  string(project.AccountId),
					"kloudlite.io/resource-ref": string(project.Id),
				},
				Annotations: map[string]string{
					"kloudlite.io/account-ref":  string(project.AccountId),
					"kloudlite.io/resource-ref": string(project.Id),
				},
			},
			Spec: op_crds.ProjectSpec{
				DisplayName: displayName,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return project, err
}

func (d *domain) UpdateProject(ctx context.Context, projectID repos.ID, displayName *string, cluster *string, logo *string, description *string) (bool, error) {
	proj, err := d.projectRepo.FindById(ctx, projectID)
	if err = mongoError(err, "project not found"); err != nil {
		return false, err
	}

	if err = d.checkAccountAccess(ctx, proj.AccountId, "create_project"); err != nil {
		return false, err
	}
	if displayName != nil {
		proj.DisplayName = *displayName
	}
	if logo != nil {
		proj.Logo = logo
	}

	if description != nil {
		proj.Description = description
	}

	_, err = d.projectRepo.UpdateById(ctx, projectID, proj)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (d *domain) DeleteProject(ctx context.Context, id repos.ID) (bool, error) {
	proj, err := d.projectRepo.FindById(ctx, id)
	if err = mongoError(err, "project not found"); err != nil {
		return false, err
	}

	if err = d.checkAccountAccess(ctx, proj.AccountId, "delete_project"); err != nil {
		return false, err
	}

	clusterId, err := d.getClusterForAccount(ctx, proj.AccountId)
	if err != nil {
		return false, err
	}

	proj.IsDeleting = true
	_, err = d.projectRepo.UpdateById(ctx, id, proj)
	if err != nil {
		return false, err
	}
	err = d.workloadMessenger.SendAction(
		"delete", d.getDispatchKafkaTopic(clusterId), string(id), &op_crds.Project{
			APIVersion: op_crds.APIVersion,
			Kind:       op_crds.ProjectKind,
			Metadata: op_crds.ProjectMetadata{
				Name: proj.Name,
			},
		},
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) OnDeleteProject(ctx context.Context, response *op_crds.StatusUpdate) error {
	return d.projectRepo.DeleteById(ctx, repos.ID(response.Metadata.ResourceId))
}

func (d *domain) getProjectRegionDetails(ctx context.Context, proj *entities.Project) (cloudProvider string, region string, err error) {

	if err = d.checkProjectAccess(ctx, proj.Id, UpdateProject); err != nil {
		return "", "", err
	}

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

	// return projectCloudProvider.Provider, projectRegion.Region, nil
	return projectCloudProvider.Provider, string(projectRegion.Id), nil
}

func (d *domain) GetDockerCredentials(ctx context.Context, projectId repos.ID) (username string, password string, err error) {

	if err = d.checkProjectAccess(ctx, projectId, "get_docker_credentials"); err != nil {
		return "", "", err
	}

	project, err := d.projectRepo.FindById(ctx, projectId)
	if err != nil {
		return "", "", err
	}

	kubecli := kubeapi.NewClientWithConfigPath(fmt.Sprintf("%s/kl-01", d.clusterConfigsPath))
	secret, err := kubecli.GetSecret(ctx, project.Name, "kloudlite-docker-registry")
	if err != nil {
		return "", "", err
	}

	var data struct {
		Auths map[string]struct {
			Auth string `json:"auth"`
		} `json:"auths"`
	}

	if err = json.Unmarshal(secret.Data[".dockerconfigjson"], &data); err != nil {
		return "", "", nil
	}

	connectionStr := data.Auths["registry.kloudlite.io"].Auth
	decodeString, err := base64.StdEncoding.DecodeString(connectionStr)
	if err != nil {
		return "", "", err
	}

	splits := strings.Split(string(decodeString), ":")
	return splits[0], splits[1], nil
}

func (d *domain) checkProjectAccess(ctx context.Context, projectId repos.ID, action string) error {

	userId, err := GetUser(ctx)
	if err != nil {
		return err
	}

	project, err := d.projectRepo.FindById(ctx, projectId)
	if err = mongoError(err, "project not found"); err != nil {
		return err
	}

	can, err := d.iamClient.Can(
		ctx, &iam.InCan{
			UserId:      userId,
			ResourceIds: []string{string(projectId), string(project.AccountId)},
			Action:      action,
		},
	)
	if err != nil {
		return err
	}

	if !can.Status {
		return fmt.Errorf("you don't have permission to perform this operation")
	}

	return nil
}

func (d *domain) checkAccountAccess(ctx context.Context, accountId repos.ID, action string) error {
	userId, err := GetUser(ctx)
	if err != nil {
		return err
	}

	// TODO: This is backdoor
	if userId == "usr-jlnueeicdfbl7elqw-bzwczt5zmo" {
		return nil
	}

	can, err := d.iamClient.Can(
		ctx, &iam.InCan{
			UserId:      userId,
			ResourceIds: []string{string(accountId)},
			Action:      action,
		},
	)
	if err != nil {
		return err
	}

	if !can.Status {
		return fmt.Errorf("you don't have permission to perform this operation")
	}

	return nil
}
