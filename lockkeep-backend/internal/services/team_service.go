package services

import (
	"context"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TeamService struct {
	teamRepo       ports.TeamRepository
	membershipRepo ports.MembershipRepository
}

func NewTeamService(teamRepo ports.TeamRepository, membershipRepo ports.MembershipRepository) *TeamService {
	return &TeamService{
		teamRepo:       teamRepo,
		membershipRepo: membershipRepo,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, userID bson.ObjectID, orgID bson.ObjectID, name, slug string) (*domain.Team, error) {
	// Verify user is org member with team manage permission
	membership, err := s.membershipRepo.FindByUserAndOrg(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if membership == nil {
		return nil, ErrEntityHasNoRoles
	}
	if !utils.HasPermission(membership.RoleID, domain.PermTeamManage) {
		return nil, ErrEntityHasInvalidRoles
	}

	// Check slug uniqueness within org
	existing, err := s.teamRepo.FindBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrTeamSlugTaken
	}

	team := &domain.Team{
		ID:             bson.NewObjectID(),
		OrganizationID: orgID,
		Name:           name,
		Slug:           slug,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	return team, nil
}

func (s *TeamService) GetTeamBySlug(ctx context.Context, orgID bson.ObjectID, slug string) (*domain.Team, error) {
	team, err := s.teamRepo.FindBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, ErrTeamNotFound
	}
	return team, nil
}

func (s *TeamService) ListTeamsByOrganization(ctx context.Context, orgID bson.ObjectID) ([]domain.Team, error) {
	return s.teamRepo.FindByOrganization(ctx, orgID)
}

// UpdateTeam is a scoped lookup placeholder: there is no team-update DTO yet,
// so the method validates the team exists within the organization and returns
// it for callers (e.g. profile views) without mutating anything.
func (s *TeamService) UpdateTeam(ctx context.Context, orgID bson.ObjectID, teamID bson.ObjectID) (*domain.Team, error) {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil || team.OrganizationID != orgID {
		return nil, ErrTeamNotFound
	}
	return team, nil
}

func (s *TeamService) DeleteTeams(ctx context.Context, userID bson.ObjectID, orgID bson.ObjectID, teamIDs []bson.ObjectID) error {
	// Verify permissions
	membership, err := s.membershipRepo.FindByUserAndOrg(ctx, userID, orgID)
	if err != nil {
		return err
	}
	if membership == nil {
		return ErrEntityHasNoRoles
	}
	if !utils.HasPermission(membership.RoleID, domain.PermTeamManage) {
		return ErrEntityHasInvalidRoles
	}

	for _, teamID := range teamIDs {
		if err := s.teamRepo.Delete(ctx, teamID); err != nil {
			return err
		}
	}

	return nil
}
