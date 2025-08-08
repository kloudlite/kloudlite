package domain

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/accounts/internal/entities"
	fc "github.com/kloudlite/api/apps/accounts/internal/entities/field-constants"
	fn "github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type teamDomain struct {
	teamRepo       repos.DbRepo[*entities.Team]
	membershipRepo repos.DbRepo[*entities.TeamMembership]
	platformDomain PlatformService
}

func (d *teamDomain) CreateTeam(ctx UserContext, team entities.Team) (*entities.Team, error) {
	if team.Slug == "" {
		return nil, errors.New("team slug is required")
	}

	if team.DisplayName == "" {
		return nil, errors.New("team display name is required")
	}

	if team.Region == "" {
		return nil, errors.New("team region is required")
	}

	// Get platform settings
	settings, err := d.platformDomain.GetPlatformSettings(ctx.Context)
	if err != nil {
		return nil, err
	}

	// Check if team creation requires approval
	if settings.TeamSettings.RequireApproval {
		// TODO: Check if user has permission to create teams directly using auth service
		// For now, all users need approval
		canCreate := false
		
		if !canCreate {
			// Check if this is user's first team and auto-approve is enabled
			userTeams, err := d.ListTeams(ctx)
			if err != nil {
				return nil, err
			}
			
			if len(userTeams) == 0 && settings.TeamSettings.AutoApproveFirstTeam {
				// Allow first team creation
			} else {
				return nil, errors.New("unauthorized: team creation requires approval. Please submit a team creation request")
			}
		}
	}

	// Check if slug is available
	existingTeam, err := d.teamRepo.FindOne(ctx.Context, repos.Filter{
		fc.TeamSlug: team.Slug,
	})
	if err != nil && !errors.Is(err, repos.ErrNoDocuments) {
		return nil, errors.NewE(err)
	}
	if existingTeam != nil {
		return nil, errors.New("team slug already exists")
	}

	// Set defaults
	team.OwnerId = ctx.UserId
	team.ContactEmail = ctx.UserEmail
	if team.IsActive == nil {
		active := true
		team.IsActive = &active
	}
	if team.ObjectMeta.Name == "" {
		team.ObjectMeta.Name = team.Slug
	}

	team.IncrementRecordVersion()

	// Create team
	createdTeam, err := d.teamRepo.Create(ctx.Context, &team)
	if err != nil {
		return nil, errors.NewE(err)
	}

	// Create owner membership
	membership := &entities.TeamMembership{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", createdTeam.Id, ctx.UserId),
		},
		UserId: ctx.UserId,
		TeamId: createdTeam.Id,
		Role:   iamT.RoleAccountOwner,
	}
	membership.IncrementRecordVersion()

	if _, err := d.membershipRepo.Create(ctx.Context, membership); err != nil {
		// Rollback team creation
		d.teamRepo.DeleteById(ctx.Context, createdTeam.Id)
		return nil, errors.NewE(err)
	}

	return createdTeam, nil
}

func (d *teamDomain) GetTeam(ctx UserContext, teamId repos.ID) (*entities.Team, error) {
	// Check if user has access to team
	membership, err := d.membershipRepo.FindOne(ctx.Context, repos.Filter{
		fc.TeamMembershipUserId: ctx.UserId,
		fc.TeamMembershipTeamId: teamId,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if membership == nil {
		return nil, errors.New("unauthorized: user is not a member of this team")
	}

	team, err := d.teamRepo.FindById(ctx.Context, teamId)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return team, nil
}

func (d *teamDomain) ListTeams(ctx UserContext) ([]*entities.Team, error) {
	// Get all team memberships for user
	memberships, err := d.membershipRepo.Find(ctx.Context, repos.Query{
		Filter: repos.Filter{
			fc.TeamMembershipUserId: ctx.UserId,
		},
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if len(memberships) == 0 {
		return []*entities.Team{}, nil
	}

	// Get team IDs
	teamIds := make([]repos.ID, len(memberships))
	for i, m := range memberships {
		teamIds[i] = m.TeamId
	}

	// Get teams
	teams, err := d.teamRepo.Find(ctx.Context, repos.Query{
		Filter: repos.Filter{
			fc.Id: repos.Filter{"$in": teamIds},
		},
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	return teams, nil
}

func (d *teamDomain) SearchTeams(ctx context.Context, query string, limit, offset int) ([]*entities.Team, int, error) {
	filter := repos.Filter{}
	if query != "" {
		filter = repos.Filter{
			"$or": []repos.Filter{
				{fc.TeamSlug: repos.Filter{"$regex": query, "$options": "i"}},
				{fn.DisplayName: repos.Filter{"$regex": query, "$options": "i"}},
			},
		}
	}

	lim := int64(limit)
	teams, err := d.teamRepo.Find(ctx, repos.Query{
		Filter: filter,
		Limit:  &lim,
		Sort:   map[string]interface{}{fc.TeamSlug: 1},
	})
	if err != nil {
		return nil, 0, errors.NewE(err)
	}

	count, err := d.teamRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, errors.NewE(err)
	}

	return teams, int(count), nil
}

func (d *teamDomain) UpdateTeam(ctx UserContext, teamId repos.ID, displayName string) (*entities.Team, error) {
	// Check if user has owner/admin access
	membership, err := d.membershipRepo.FindOne(ctx.Context, repos.Filter{
		fc.TeamMembershipUserId: ctx.UserId,
		fc.TeamMembershipTeamId: teamId,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if membership == nil {
		return nil, errors.New("unauthorized: user is not a member of this team")
	}
	if membership.Role != iamT.RoleAccountOwner && membership.Role != iamT.RoleAccountAdmin {
		return nil, errors.New("unauthorized: insufficient permissions")
	}

	team, err := d.teamRepo.FindById(ctx.Context, teamId)
	if err != nil {
		return nil, errors.NewE(err)
	}

	team.DisplayName = displayName
	team.IncrementRecordVersion()

	updatedTeam, err := d.teamRepo.UpdateById(ctx.Context, teamId, team)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return updatedTeam, nil
}

func (d *teamDomain) CheckTeamSlugAvailability(ctx context.Context, slug string) (*CheckNameAvailabilityOutput, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	
	if len(slug) < 3 {
		return &CheckNameAvailabilityOutput{
			Result: false,
			SuggestedNames: []string{},
		}, nil
	}

	// Check existing teams
	existingTeam, err := d.teamRepo.FindOne(ctx, repos.Filter{
		fc.TeamSlug: slug,
	})
	if err != nil && !errors.Is(err, repos.ErrNoDocuments) {
		return nil, errors.NewE(err)
	}

	// Check pending requests through platform domain
	isAvailable := existingTeam == nil
	if isAvailable && d.platformDomain != nil {
		// Check if slug is taken by a pending request
		isAvailable, err = d.platformDomain.IsTeamSlugAvailableForRequest(ctx, slug)
		if err != nil {
			return nil, errors.NewE(err)
		}
	}

	if !isAvailable {
		// Generate suggestions that check both teams and pending requests
		suggestions := []string{}
		for i := 1; i <= 5; i++ {
			suggestion := fmt.Sprintf("%s-%d", slug, i)
			
			// Check if team exists
			exists, err := d.teamRepo.FindOne(ctx, repos.Filter{
				fc.TeamSlug: suggestion,
			})
			if err != nil && !errors.Is(err, repos.ErrNoDocuments) {
				return nil, errors.NewE(err)
			}
			
			if exists == nil {
				// Also check pending requests
				available := true
				if d.platformDomain != nil {
					available, err = d.platformDomain.IsTeamSlugAvailableForRequest(ctx, suggestion)
					if err != nil {
						return nil, errors.NewE(err)
					}
				}
				
				if available {
					suggestions = append(suggestions, suggestion)
					if len(suggestions) >= 3 {
						break
					}
				}
			}
		}

		return &CheckNameAvailabilityOutput{
			Result: false,
			SuggestedNames: suggestions,
		}, nil
	}

	return &CheckNameAvailabilityOutput{
		Result: true,
		SuggestedNames: []string{},
	}, nil
}

func (d *teamDomain) GenerateTeamSlugSuggestions(ctx context.Context, displayName string) []string {
	if displayName == "" {
		return []string{}
	}

	// Convert display name to slug format
	reg := regexp.MustCompile(`[^a-z0-9\s-]`)
	baseSlug := strings.ToLower(displayName)
	baseSlug = reg.ReplaceAllString(baseSlug, "")
	baseSlug = strings.ReplaceAll(baseSlug, " ", "-")
	baseSlug = regexp.MustCompile(`-+`).ReplaceAllString(baseSlug, "-")
	baseSlug = strings.Trim(baseSlug, "-")
	
	// Limit length
	if len(baseSlug) > 30 {
		baseSlug = baseSlug[:30]
		// Make sure we don't end with a hyphen after truncation
		baseSlug = strings.TrimRight(baseSlug, "-")
	}
	
	// If baseSlug is empty after cleaning, generate a default
	if baseSlug == "" {
		baseSlug = "team"
	}

	suggestions := []string{}
	
	// First try the base slug
	exists, err := d.teamRepo.FindOne(ctx, repos.Filter{
		fc.TeamSlug: baseSlug,
	})
	if err == nil && exists == nil {
		// Also check pending requests
		available := true
		if d.platformDomain != nil {
			available, err = d.platformDomain.IsTeamSlugAvailableForRequest(ctx, baseSlug)
			if err != nil {
				available = false
			}
		}
		if available {
			suggestions = append(suggestions, baseSlug)
		}
	}
	
	// Generate variations
	variations := []string{
		fmt.Sprintf("%s-team", baseSlug),
		fmt.Sprintf("team-%s", baseSlug),
		fmt.Sprintf("%s-org", baseSlug),
	}
	
	// Add number suffixes
	for i := 1; i <= 3; i++ {
		variations = append(variations, fmt.Sprintf("%s-%d", baseSlug, i))
	}
	
	// Check each variation
	for _, slug := range variations {
		// Ensure slug doesn't exceed 30 characters
		if len(slug) > 30 {
			continue
		}
		
		exists, err := d.teamRepo.FindOne(ctx, repos.Filter{
			fc.TeamSlug: slug,
		})
		if err == nil && exists == nil {
			// Also check pending requests
			available := true
			if d.platformDomain != nil {
				available, err = d.platformDomain.IsTeamSlugAvailableForRequest(ctx, slug)
				if err != nil {
					available = false
				}
			}
			if available {
				suggestions = append(suggestions, slug)
				if len(suggestions) >= 5 {
					break
				}
			}
		}
	}
	
	return suggestions
}

func (d *teamDomain) GetUserRoleInTeam(ctx context.Context, userId repos.ID, teamId repos.ID) (iamT.Role, error) {
	membership, err := d.membershipRepo.FindOne(ctx, repos.Filter{
		fc.TeamMembershipUserId: userId,
		fc.TeamMembershipTeamId: teamId,
	})
	if err != nil {
		return "", errors.NewE(err)
	}
	if membership == nil {
		return "", errors.New("user is not a member of this team")
	}
	return membership.Role, nil
}