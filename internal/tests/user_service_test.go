package tests

import (
	"context"
	"testing"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/config"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Test Setup
func newTestUserService() (*services.UserService, *mocks.MockUserRepository) {
	cfg := &config.Config{}

	userRepo := new(mocks.MockUserRepository)

	svc := services.NewUserService(cfg, userRepo)

	return svc, userRepo
}

func TestUserService_Me(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, userRepo := newTestUserService()

		user := &domain.User{ID: bson.NewObjectID(), Email: "me@example.com"}

		userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil).Once()

		result, err := svc.Me(context.Background(), user.ID.Hex())

		assert.NoError(t, err)
		assert.Equal(t, user.Email, result.Email)
		userRepo.AssertExpectations(t)
	})

	t.Run("invalid user id", func(t *testing.T) {
		svc, _ := newTestUserService()

		_, err := svc.Me(context.Background(), "not-a-valid-id")
		assert.Error(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		svc, userRepo := newTestUserService()

		userID := bson.NewObjectID()
		userRepo.On("FindByID", mock.Anything, userID).Return(nil, nil).Once()

		_, err := svc.Me(context.Background(), userID.Hex())
		assert.ErrorIs(t, err, services.ErrUserNotFound)
	})
}
