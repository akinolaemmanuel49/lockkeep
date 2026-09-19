package tests

import (
	"context"
	"testing"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/config"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func strPtr(s string) *string { return &s }

// Test Setup
func newTestUserService() (*services.UserService, *mocks.MockUserRepository) {
	cfg := &config.Config{}

	userRepo := new(mocks.MockUserRepository)

	svc := services.NewUserService(cfg, userRepo)

	return svc, userRepo
}

func TestUserService_GetUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, userRepo := newTestUserService()

		user := &domain.User{ID: bson.NewObjectID(), Email: "me@example.com"}

		userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil).Once()

		result, err := svc.GetUser(context.Background(), user.ID.Hex())

		assert.NoError(t, err)
		assert.Equal(t, user.Email, result.Email)
		userRepo.AssertExpectations(t)
	})

	t.Run("invalid user id", func(t *testing.T) {
		svc, _ := newTestUserService()

		_, err := svc.GetUser(context.Background(), "not-a-valid-id")
		assert.Error(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		svc, userRepo := newTestUserService()

		userID := bson.NewObjectID()
		userRepo.On("FindByID", mock.Anything, userID).Return(nil, nil).Once()

		_, err := svc.GetUser(context.Background(), userID.Hex())
		assert.ErrorIs(t, err, services.ErrUserNotFound)
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, userRepo := newTestUserService()

		user := &domain.User{ID: bson.NewObjectID(), Email: "me@example.com", Username: "old-name"}

		userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil).Once()
		userRepo.On("UpdateProfile", mock.Anything, user.ID, mock.AnythingOfType("domain.User")).Return(nil).Once()

		result, err := svc.UpdateUser(context.Background(), user.ID.Hex(), dto.UpdateUserProfileRequestDTO{
			Username:  strPtr("new-name"),
			AvatarURL: nil,
		})

		assert.NoError(t, err)
		assert.Equal(t, "new-name", result.Username)
		userRepo.AssertExpectations(t)
	})

	t.Run("invalid update - no fields", func(t *testing.T) {
		svc, userRepo := newTestUserService()

		user := &domain.User{ID: bson.NewObjectID(), Email: "me@example.com"}

		userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil).Once()

		_, err := svc.UpdateUser(context.Background(), user.ID.Hex(), dto.UpdateUserProfileRequestDTO{})

		assert.ErrorIs(t, err, services.ErrInvalidProfileUpdate)
		userRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		svc, userRepo := newTestUserService()

		userID := bson.NewObjectID()
		userRepo.On("FindByID", mock.Anything, userID).Return(nil, nil).Once()

		_, err := svc.UpdateUser(context.Background(), userID.Hex(), dto.UpdateUserProfileRequestDTO{
			Username: strPtr("new-name"),
		})

		assert.ErrorIs(t, err, services.ErrUserNotFound)
		userRepo.AssertExpectations(t)
	})
}