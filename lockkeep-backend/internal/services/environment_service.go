package services

import (
	"context"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/crypto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type EnvironmentService struct {
	appRepo        ports.ApplicationRepository
	envRepo        ports.EnvironmentRepository
	secretRepo     ports.EnvSecretRepository
	keyRepo        ports.KeyManagementRepository
	membershipRepo ports.MembershipRepository
	keyProvider    crypto.KeyProvider
}

func NewEnvironmentService(
	appRepo ports.ApplicationRepository,
	envRepo ports.EnvironmentRepository,
	secretRepo ports.EnvSecretRepository,
	keyRepo ports.KeyManagementRepository,
	membershipRepo ports.MembershipRepository,
	keyProvider crypto.KeyProvider,
) *EnvironmentService {
	return &EnvironmentService{
		appRepo:        appRepo,
		envRepo:        envRepo,
		secretRepo:     secretRepo,
		keyRepo:        keyRepo,
		membershipRepo: membershipRepo,
		keyProvider:    keyProvider,
	}
}

func (s *EnvironmentService) requireMember(ctx context.Context, userID, orgID bson.ObjectID, perm domain.Permission) error {
	membership, err := s.membershipRepo.FindByUserAndOrg(ctx, userID, orgID)
	if err != nil {
		return err
	}
	if membership == nil {
		return ErrEntityHasNoRoles
	}
	if !utils.HasPermission(membership.RoleID, perm) {
		return ErrEntityHasInvalidRoles
	}
	return nil
}

func (s *EnvironmentService) CreateApplication(ctx context.Context, userID, orgID bson.ObjectID, name, slug string, description *string) (*domain.Application, error) {
	if err := s.requireMember(ctx, userID, orgID, domain.PermSecretWrite); err != nil {
		return nil, err
	}

	existing, err := s.appRepo.FindBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrApplicationSlugTaken
	}

	app := &domain.Application{
		ID:             bson.NewObjectID(),
		OrganizationID: orgID,
		Name:           name,
		Slug:           slug,
		Description:    description,
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := s.appRepo.Create(ctx, app); err != nil {
		return nil, err
	}
	return app, nil
}

func (s *EnvironmentService) GetApplicationBySlug(ctx context.Context, orgID bson.ObjectID, slug string) (*domain.Application, error) {
	app, err := s.appRepo.FindBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, ErrApplicationNotFound
	}
	return app, nil
}

func (s *EnvironmentService) ListApplications(ctx context.Context, orgID bson.ObjectID) ([]domain.Application, error) {
	return s.appRepo.FindByOrganization(ctx, orgID)
}

func (s *EnvironmentService) UpdateApplication(ctx context.Context, userID, orgID, appID bson.ObjectID, update dto.UpdateApplicationDTO) (*domain.Application, error) {
	if err := s.requireMember(ctx, userID, orgID, domain.PermSecretWrite); err != nil {
		return nil, err
	}

	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil || app.OrganizationID != orgID {
		return nil, ErrApplicationNotFound
	}
	if update.Name == nil && update.Slug == nil && update.Description == nil {
		return nil, ErrInvalidOrganizationUpdate
	}

	next := *app
	if update.Name != nil {
		next.Name = *update.Name
	}
	if update.Slug != nil && *update.Slug != app.Slug {
		existing, err := s.appRepo.FindBySlug(ctx, orgID, *update.Slug)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrApplicationSlugTaken
		}
		next.Slug = *update.Slug
	}
	if update.Description != nil {
		next.Description = update.Description
	}
	next.UpdatedAt = time.Now()

	if err := s.appRepo.Update(ctx, appID, next); err != nil {
		return nil, err
	}
	return &next, nil
}

func (s *EnvironmentService) DeleteApplication(ctx context.Context, userID, orgID, appID bson.ObjectID) error {
	if err := s.requireMember(ctx, userID, orgID, domain.PermSecretWrite); err != nil {
		return err
	}

	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return err
	}
	if app == nil || app.OrganizationID != orgID {
		return ErrApplicationNotFound
	}

	envs, err := s.envRepo.FindByApplication(ctx, appID)
	if err != nil {
		return err
	}
	for _, env := range envs {
		if err := s.deleteEnvironmentData(ctx, env.ID); err != nil {
			return err
		}
	}

	return s.appRepo.Delete(ctx, appID)
}

func (s *EnvironmentService) CreateEnvironment(ctx context.Context, userID, orgID, appID bson.ObjectID, name, slug string, isProtected bool) (*domain.Environment, error) {
	if err := s.requireMember(ctx, userID, orgID, domain.PermSecretWrite); err != nil {
		return nil, err
	}

	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil || app.OrganizationID != orgID {
		return nil, ErrApplicationNotFound
	}

	existing, err := s.envRepo.FindBySlug(ctx, orgID, appID, slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEnvironmentSlugTaken
	}

	env := &domain.Environment{
		ID:             bson.NewObjectID(),
		OrganizationID: orgID,
		ApplicationID:  appID,
		Name:           name,
		Slug:           slug,
		IsProtected:    isProtected,
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := s.envRepo.Create(ctx, env); err != nil {
		return nil, err
	}

	// Provision the environment's data-encryption key. On failure roll back the
	// environment so we never leave a DEK-less environment behind.
	wrapped, _, err := s.keyProvider.GenerateDataKey(ctx, "", crypto.ScopeEnvironment, env.ID)
	if err != nil {
		_ = s.envRepo.Delete(ctx, env.ID)
		return nil, err
	}
	if err := s.keyRepo.Create(ctx, wrapped); err != nil {
		_ = s.envRepo.Delete(ctx, env.ID)
		return nil, err
	}

	return env, nil
}

func (s *EnvironmentService) GetEnvironmentBySlug(ctx context.Context, orgID, appID bson.ObjectID, slug string) (*domain.Environment, error) {
	env, err := s.envRepo.FindBySlug(ctx, orgID, appID, slug)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, ErrEnvironmentNotFound
	}
	return env, nil
}

func (s *EnvironmentService) ListEnvironments(ctx context.Context, appID bson.ObjectID) ([]domain.Environment, error) {
	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, ErrApplicationNotFound
	}
	return s.envRepo.FindByApplication(ctx, appID)
}

func (s *EnvironmentService) UpdateEnvironment(ctx context.Context, userID, orgID, envID bson.ObjectID, update dto.UpdateEnvironmentDTO) (*domain.Environment, error) {
	if err := s.requireMember(ctx, userID, orgID, domain.PermSecretWrite); err != nil {
		return nil, err
	}

	env, err := s.envRepo.FindByID(ctx, envID)
	if err != nil {
		return nil, err
	}
	if env == nil || env.OrganizationID != orgID {
		return nil, ErrEnvironmentNotFound
	}
	if update.Name == nil && update.Slug == nil && update.IsProtected == nil {
		return nil, ErrInvalidOrganizationUpdate
	}

	next := *env
	if update.Name != nil {
		next.Name = *update.Name
	}
	if update.Slug != nil && *update.Slug != env.Slug {
		existing, err := s.envRepo.FindBySlug(ctx, orgID, env.ApplicationID, *update.Slug)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrEnvironmentSlugTaken
		}
		next.Slug = *update.Slug
	}
	if update.IsProtected != nil {
		next.IsProtected = *update.IsProtected
	}
	next.UpdatedAt = time.Now()

	if err := s.envRepo.Update(ctx, envID, next); err != nil {
		return nil, err
	}
	return &next, nil
}

func (s *EnvironmentService) DeleteEnvironment(ctx context.Context, userID, orgID, envID bson.ObjectID) error {
	if err := s.requireMember(ctx, userID, orgID, domain.PermSecretWrite); err != nil {
		return err
	}

	env, err := s.envRepo.FindByID(ctx, envID)
	if err != nil {
		return err
	}
	if env == nil || env.OrganizationID != orgID {
		return ErrEnvironmentNotFound
	}

	return s.deleteEnvironmentData(ctx, envID)
}

func (s *EnvironmentService) deleteEnvironmentData(ctx context.Context, envID bson.ObjectID) error {
	secrets, err := s.secretRepo.FindByEnvironment(ctx, envID)
	if err != nil {
		return err
	}
	for _, secret := range secrets {
		_ = s.secretRepo.Delete(ctx, secret.ID)
	}

	wrapped, err := s.keyRepo.FindByScope(ctx, crypto.ScopeEnvironment, envID)
	if err != nil {
		return err
	}
	if wrapped != nil {
		if err := s.keyRepo.Delete(ctx, wrapped.ID); err != nil {
			return err
		}
	}

	return s.envRepo.Delete(ctx, envID)
}