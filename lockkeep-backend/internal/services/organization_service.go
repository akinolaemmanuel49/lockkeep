package services

import (
	"context"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrganizationService struct {
	orgRepo        ports.OrganizationRepository
	membershipRepo ports.MembershipRepository
}

func NewOrganizationService(orgRepo ports.OrganizationRepository, membershipRepo ports.MembershipRepository) *OrganizationService {
	return &OrganizationService{
		orgRepo:        orgRepo,
		membershipRepo: membershipRepo,
	}
}

func (s *OrganizationService) Create(ctx context.Context, userID bson.ObjectID, name, slug string) (*domain.Organization, error) {
	// Check slug uniqueness
	existing, err := s.orgRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrOrgSlugTaken
	}

	org := &domain.Organization{
		ID:      bson.NewObjectID(),
		Name:    name,
		Slug:    slug,
		OwnerID: userID,
		Settings: domain.OrgSettings{
			DefaultTeamRole: domain.RoleTeamUser,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, err
	}

	// Creator becomes owner
	membership := &domain.Membership{
		ID:             bson.NewObjectID(),
		UserID:         userID,
		OrganizationID: org.ID,
		RoleID:         domain.RoleOrgOwner,
		JoinedAt:       time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.membershipRepo.Create(ctx, membership); err != nil {
		// TODO: rollback org creation
		return nil, err
	}

	return org, nil
}

func (s *OrganizationService) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	org, err := s.orgRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, ErrOrganizationNotFound
	}
	return org, nil
}

func (s *OrganizationService) ListForUser(ctx context.Context, userID bson.ObjectID) ([]domain.Organization, error) {
	return s.orgRepo.FindByMember(ctx, userID)
}

func (s *OrganizationService) Update(ctx context.Context, userID bson.ObjectID, slug string, name string) (*domain.Organization, error) {
	org, err := s.orgRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, ErrOrganizationNotFound
	}

	// Verify user is owner or admin
	membership, err := s.membershipRepo.FindByUserAndOrg(ctx, userID, org.ID)
	if err != nil {
		return nil, err
	}
	if membership == nil {
		return nil, ErrEntityHasNoRoles
	}
	if membership.RoleID != domain.RoleOrgOwner && membership.RoleID != domain.RoleOrgAdmin {
		return nil, ErrEntityHasInvalidRoles
	}

	org.Name = name
	org.UpdatedAt = time.Now()

	if err := s.orgRepo.Update(ctx, org.ID, *org); err != nil {
		return nil, err
	}

	return org, nil
}

func (s *OrganizationService) Delete(ctx context.Context, userID bson.ObjectID, slug string) error {
	org, err := s.orgRepo.FindBySlug(ctx, slug)
	if err != nil {
		return err
	}
	if org == nil {
		return ErrOrganizationNotFound
	}

	// Only owner can delete
	if org.OwnerID != userID {
		return ErrNotOrgOwner
	}

	// TODO: cascade delete memberships, teams, shared secrets

	return s.orgRepo.Delete(ctx, org.ID)
}
