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

func (s *TeamService) Create(ctx context.Context, userID bson.ObjectID, orgID bson.ObjectID, name, slug string) (*domain.Team, error) {
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

func (s *TeamService) GetBySlug(ctx context.Context, orgID bson.ObjectID, slug string) (*domain.Team, error) {
	team, err := s.teamRepo.FindBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, ErrTeamNotFound
	}
	return team, nil
}

func (s *TeamService) ListByOrganization(ctx context.Context, orgID bson.ObjectID) ([]domain.Team, error) {
	return s.teamRepo.FindByOrganization(ctx, orgID)
}

func (s *TeamService) Delete(ctx context.Context, userID bson.ObjectID, orgID bson.ObjectID, teamID bson.ObjectID) error {
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

	return s.teamRepo.Delete(ctx, teamID)
}
