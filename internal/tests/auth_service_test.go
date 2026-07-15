package tests

import (
	"context"
	"testing"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/config"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	auth_errors "github.com/akinolaemmanuel49/lockkeep-backend/internal/errors/auth"
	user_errors "github.com/akinolaemmanuel49/lockkeep-backend/internal/errors/user"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/service"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

// Test Setup
func newTestAuthService() (*service.AuthService, *mocks.MockUserRepository, *mocks.MockIdentityRepository, *mocks.MockMembershipRepository, *mocks.MockUnitOfWork, *mocks.MockJWTManager) {
	cfg := &config.Config{}

	userRepo := new(mocks.MockUserRepository)
	identityRepo := new(mocks.MockIdentityRepository)
	membershipRepo := new(mocks.MockMembershipRepository)
	uow := new(mocks.MockUnitOfWork)
	jwtManager := new(mocks.MockJWTManager)

	svc := service.NewAuthService(cfg, userRepo, identityRepo, membershipRepo, uow, jwtManager)

	return svc, userRepo, identityRepo, membershipRepo, uow, jwtManager
}

func TestAuthService_Register(t *testing.T) {
	svc, userRepo, identityRepo, _, uow, _ := newTestAuthService()

	t.Run("successful registration", func(t *testing.T) {
		input := dto.RegisterRequestDTO{
			Username: "newuser",
			Email:    "new@example.com",
			Password: "password123",
		}

		userRepo.On("EmailExists", mock.Anything, input.Email).Return(false, nil)
		userRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
		userRepo.On("FindByID", mock.Anything, mock.Anything).Return(&domain.User{ID: bson.NewObjectID(), Email: input.Email, Username: input.Username}, nil)
		identityRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Identity")).Return(nil)
		uow.On("Within", mock.Anything, mock.Anything).Return(nil)

		user, err := svc.Register(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, input.Email, user.Email)
		userRepo.AssertExpectations(t)
		identityRepo.AssertExpectations(t)
		uow.AssertExpectations(t)
	})

	t.Run("email already taken", func(t *testing.T) {
		input := dto.RegisterRequestDTO{Email: "taken@example.com"}

		userRepo.On("EmailExists", mock.Anything, input.Email).Return(true, nil)
		uow.On("Within", mock.Anything, mock.Anything).Return(nil)

		_, err := svc.Register(context.Background(), input)

		assert.ErrorIs(t, err, user_errors.ErrEmailTaken)
		userRepo.AssertExpectations(t)
		uow.AssertExpectations(t)
	})
}

func TestAuthService_Login(t *testing.T) {
	svc, userRepo, identityRepo, membershipRepo, _, jwtManager := newTestAuthService()

	t.Run("successful login", func(t *testing.T) {
		email := "login@example.com"
		password := "password123"
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		user := &domain.User{ID: bson.NewObjectID(), Email: email}
		identity := domain.Identity{
			ID:           bson.NewObjectID(),
			UserID:       user.ID,
			AuthMethod:   domain.AuthMethodLocal,
			PasswordHash: string(hash),
		}

		membershipRepo.On("FindByUser", mock.Anything, user.ID).
			Return([]domain.Membership{}, nil)
		userRepo.On("FindByEmail", mock.Anything, email).Return(user, nil)
		identityRepo.On("FindByUserID", mock.Anything, user.ID).Return([]domain.Identity{identity}, nil)
		identityRepo.On("RecordLogin", mock.Anything, mock.Anything).Return(nil)
		jwtManager.On("GeneratePair", mock.Anything).Return(&dto.TokenPair{AccessToken: "access.token", RefreshToken: "refresh.token"}, nil)

		returnedUser, tokens, err := svc.Login(context.Background(), dto.LoginRequestDTO{
			Email:    email,
			Password: password,
		})

		assert.NoError(t, err)
		assert.NotNil(t, returnedUser)
		assert.NotNil(t, tokens)
		assert.Equal(t, "access.token", tokens.AccessToken)
		userRepo.AssertExpectations(t)
		identityRepo.AssertExpectations(t)
		membershipRepo.AssertExpectations(t)
		jwtManager.AssertExpectations(t)
	})
	t.Run("wrong password", func(t *testing.T) {
		email := "login@example.com"
		password := "password123"
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		user := &domain.User{ID: bson.NewObjectID(), Email: email}
		identity := domain.Identity{
			ID:           bson.NewObjectID(),
			UserID:       user.ID,
			AuthMethod:   domain.AuthMethodLocal,
			PasswordHash: string(hash),
		}

		userRepo.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()
		identityRepo.On("FindByUserID", mock.Anything, user.ID).Return([]domain.Identity{identity}, nil).Once()

		_, tokens, err := svc.Login(context.Background(), dto.LoginRequestDTO{
			Email:    email,
			Password: "wrongpassword",
		})

		assert.ErrorIs(t, err, auth_errors.ErrInvalidCredentials)
		assert.Nil(t, tokens)
	})

	t.Run("user not found", func(t *testing.T) {
		userRepo.On("FindByEmail", mock.Anything, "missing@example.com").Return(nil, nil).Once()

		_, tokens, err := svc.Login(context.Background(), dto.LoginRequestDTO{
			Email:    "missing@example.com",
			Password: "password123",
		})

		assert.ErrorIs(t, err, auth_errors.ErrInvalidCredentials)
		assert.Nil(t, tokens)
	})
}

func TestAuthService_Me(t *testing.T) {
	svc, userRepo, _, _, _, _ := newTestAuthService()

	t.Run("success", func(t *testing.T) {
		user := &domain.User{ID: bson.NewObjectID(), Email: "me@example.com"}

		userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil).Once()

		result, err := svc.Me(context.Background(), user.ID.Hex())

		assert.NoError(t, err)
		assert.Equal(t, user.Email, result.Email)
		userRepo.AssertExpectations(t)
	})

	t.Run("invalid user id", func(t *testing.T) {
		_, err := svc.Me(context.Background(), "not-a-valid-id")
		assert.Error(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		userID := bson.NewObjectID()
		userRepo.On("FindByID", mock.Anything, userID).Return(nil, nil).Once()

		_, err := svc.Me(context.Background(), userID.Hex())
		assert.ErrorIs(t, err, user_errors.ErrUserNotFound)
	})
}

func TestAuthService_Refresh(t *testing.T) {
	svc, userRepo, _, membershipRepo, _, jwtManager := newTestAuthService()

	t.Run("successful refresh", func(t *testing.T) {
		userID := bson.NewObjectID()
		user := &domain.User{ID: userID}

		jwtManager.On("ValidateRefreshToken", "valid.refresh").Return(userID.Hex(), nil).Once()
		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil).Once()
		membershipRepo.On("FindByUser", mock.Anything, user.ID).Return([]domain.Membership{}, nil).Once()
		jwtManager.On("GeneratePair", mock.Anything).Return(&dto.TokenPair{AccessToken: "new.acc", RefreshToken: "new.ref"}, nil).Once()

		tokens, err := svc.Refresh(context.Background(), "valid.refresh")

		assert.NoError(t, err)
		assert.NotNil(t, tokens)
		assert.Equal(t, "new.acc", tokens.AccessToken)
		jwtManager.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		membershipRepo.AssertExpectations(t)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		jwtManager.On("ValidateRefreshToken", "invalid.token").Return("", auth_errors.ErrInvalidCredentials).Once()

		_, err := svc.Refresh(context.Background(), "invalid.token")
		assert.Error(t, err)
		jwtManager.AssertExpectations(t)
	})

	t.Run("user not found after valid token", func(t *testing.T) {
		userID := bson.NewObjectID()

		jwtManager.On("ValidateRefreshToken", "valid.but.ghost").Return(userID.Hex(), nil).Once()
		userRepo.On("FindByID", mock.Anything, userID).Return(nil, nil).Once()

		_, err := svc.Refresh(context.Background(), "valid.but.ghost")
		assert.ErrorIs(t, err, user_errors.ErrUserNotFound)
	})
}

func TestAuthService_OAuth(t *testing.T) {
	t.Skip("OAuth requires Auth0 HTTP client abstraction — add ports.OAuthProvider interface")
}
