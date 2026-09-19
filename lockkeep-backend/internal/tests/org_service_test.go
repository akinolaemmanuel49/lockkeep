// internal/tests/organization_service_test.go
package tests

import (
	"context"
	"testing"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func newTestOrganizationService() (*services.OrganizationService, *mocks.MockOrganizationRepository, *mocks.MockMembershipRepository) {
	orgRepo := new(mocks.MockOrganizationRepository)
	membershipRepo := new(mocks.MockMembershipRepository)

	svc := services.NewOrganizationService(orgRepo, membershipRepo)

	return svc, orgRepo, membershipRepo
}

func TestOrganizationService_Create(t *testing.T) {

	t.Run("successful creation", func(t *testing.T) {
		svc, orgRepo, membershipRepo := newTestOrganizationService()

		userID := bson.NewObjectID()
		name := "Test Org"
		slug := "test-org"

		orgRepo.On("FindBySlug", mock.Anything, slug).Return(nil, nil)
		orgRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Organization")).Return(nil)
		membershipRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Membership")).Return(nil)

		org, err := svc.Create(context.Background(), userID, name, slug)

		assert.NoError(t, err)
		assert.NotNil(t, org)
		assert.Equal(t, name, org.Name)
		assert.Equal(t, slug, org.Slug)
		assert.Equal(t, userID, org.OwnerID)
		orgRepo.AssertExpectations(t)
		membershipRepo.AssertExpectations(t)
	})

	t.Run("slug already taken", func(t *testing.T) {
		svc, orgRepo, _ := newTestOrganizationService()

		userID := bson.NewObjectID()
		slug := "taken-slug"

		existing := &domain.Organization{ID: bson.NewObjectID(), Slug: slug}

		orgRepo.On("FindBySlug", mock.Anything, slug).Return(existing, nil)

		_, err := svc.Create(context.Background(), userID, "Test", slug)

		assert.ErrorIs(t, err, services.ErrOrgSlugTaken)
		orgRepo.AssertExpectations(t)
	})
}

func TestOrganizationService_GetBySlug(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		svc, orgRepo, _ := newTestOrganizationService()

		slug := "my-org"
		org := &domain.Organization{ID: bson.NewObjectID(), Slug: slug, Name: "My Org"}

		orgRepo.On("FindBySlug", mock.Anything, slug).Return(org, nil)

		result, err := svc.GetBySlug(context.Background(), slug)

		assert.NoError(t, err)
		assert.Equal(t, org.ID, result.ID)
		orgRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		svc, orgRepo, _ := newTestOrganizationService()

		slug := "missing-org"
		orgRepo.On("FindBySlug", mock.Anything, slug).Return(nil, nil)

		_, err := svc.GetBySlug(context.Background(), slug)

		assert.ErrorIs(t, err, services.ErrOrganizationNotFound)
		orgRepo.AssertExpectations(t)
	})
}

func TestOrganizationService_ListForUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, orgRepo, _ := newTestOrganizationService()

		userID := bson.NewObjectID()
		orgs := []domain.Organization{
			{ID: bson.NewObjectID(), Name: "Org 1"},
			{ID: bson.NewObjectID(), Name: "Org 2"},
		}

		orgRepo.On("FindByMember", mock.Anything, userID).Return(orgs, nil)

		result, err := svc.ListForUser(context.Background(), userID)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		orgRepo.AssertExpectations(t)
	})
}

func TestOrganizationService_Update(t *testing.T) {
	t.Run("successful update by owner", func(t *testing.T) {
		svc, orgRepo, membershipRepo := newTestOrganizationService()

		userID := bson.NewObjectID()
		slug := "my-org"
		newName := "Updated Org"
		orgID := bson.NewObjectID()

		org := &domain.Organization{
			ID:      orgID,
			Slug:    slug,
			Name:    "Old Name",
			OwnerID: userID,
		}

		membership := &domain.Membership{
			UserID: userID,
			RoleID: domain.RoleOrgOwner,
		}

		orgRepo.On("FindBySlug", mock.Anything, slug).Return(org, nil)
		membershipRepo.On("FindByUserAndOrg", mock.Anything, userID, orgID).Return(membership, nil)
		orgRepo.On("Update", mock.Anything, orgID, mock.AnythingOfType("domain.Organization")).Return(nil)

		updated, err := svc.Update(context.Background(), userID, slug, newName)

		assert.NoError(t, err)
		assert.Equal(t, newName, updated.Name)

		orgRepo.AssertExpectations(t)
		membershipRepo.AssertExpectations(t)
	})

	t.Run("insufficient permissions", func(t *testing.T) {
		svc, orgRepo, membershipRepo := newTestOrganizationService()

		userID := bson.NewObjectID()
		slug := "my-org"
		orgID := bson.NewObjectID()

		org := &domain.Organization{
			ID:   orgID,
			Slug: slug,
		}

		membership := &domain.Membership{
			UserID: userID,
			RoleID: domain.RoleTeamUser,
		}

		// Setup mocks with consistent IDs
		orgRepo.On("FindBySlug", mock.Anything, slug).Return(org, nil)
		membershipRepo.On("FindByUserAndOrg", mock.Anything, userID, orgID).
			Return(membership, nil)

		_, err := svc.Update(context.Background(), userID, slug, "New Name")

		assert.ErrorIs(t, err, services.ErrEntityHasInvalidRoles)

		orgRepo.AssertExpectations(t)
		membershipRepo.AssertExpectations(t)
	})
}

func TestOrganizationService_Delete(t *testing.T) {
	t.Run("successful delete by owner", func(t *testing.T) {
		svc, orgRepo, _ := newTestOrganizationService()

		userID := bson.NewObjectID()
		slug := "my-org"

		org := &domain.Organization{ID: bson.NewObjectID(), Slug: slug, OwnerID: userID}

		orgRepo.On("FindBySlug", mock.Anything, slug).Return(org, nil)
		orgRepo.On("Delete", mock.Anything, org.ID).Return(nil)

		err := svc.Delete(context.Background(), userID, slug)

		assert.NoError(t, err)
		orgRepo.AssertExpectations(t)
	})

	t.Run("not owner", func(t *testing.T) {
		svc, orgRepo, _ := newTestOrganizationService()

		userID := bson.NewObjectID()
		ownerID := bson.NewObjectID()
		slug := "my-org"

		org := &domain.Organization{ID: bson.NewObjectID(), Slug: slug, OwnerID: ownerID}

		orgRepo.On("FindBySlug", mock.Anything, slug).Return(org, nil)

		err := svc.Delete(context.Background(), userID, slug)

		assert.ErrorIs(t, err, services.ErrNotOrgOwner)
	})
}
