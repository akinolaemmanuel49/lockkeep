// internal/tests/auth_service_test.go
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
	"golang.org/x/crypto/bcrypt"
)

// Test Setup
func newTestAuthService() (*services.AuthService, *mocks.MockUserRepository, *mocks.MockIdentityRepository, *mocks.MockMembershipRepository, *mocks.MockUnitOfWork, *mocks.MockJWTManager) {
	cfg := &config.Config{}

	userRepo := new(mocks.MockUserRepository)
	identityRepo := new(mocks.MockIdentityRepository)
	membershipRepo := new(mocks.MockMembershipRepository)
	uow := new(mocks.MockUnitOfWork)
	jwtManager := new(mocks.MockJWTManager)

	svc := services.NewAuthService(cfg, userRepo, identityRepo, membershipRepo, uow, jwtManager)

	return svc, userRepo, identityRepo, membershipRepo, uow, jwtManager
}

func TestAuthService_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		svc, userRepo, identityRepo, _, uow, _ := newTestAuthService()

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
		svc, userRepo, _, _, uow, _ := newTestAuthService()

		input := dto.RegisterRequestDTO{Email: "taken@example.com"}

		userRepo.On("EmailExists", mock.Anything, input.Email).Return(true, nil)
		uow.On("Within", mock.Anything, mock.Anything).Return(nil)

		_, err := svc.Register(context.Background(), input)

		assert.ErrorIs(t, err, services.ErrEmailTaken)
		userRepo.AssertExpectations(t)
		uow.AssertExpectations(t)
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		svc, userRepo, identityRepo, membershipRepo, _, jwtManager := newTestAuthService()

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
		svc, userRepo, identityRepo, _, _, _ := newTestAuthService()

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

		assert.ErrorIs(t, err, services.ErrInvalidCredentials)
		assert.Nil(t, tokens)
	})

	t.Run("user not found", func(t *testing.T) {
		svc, userRepo, _, _, _, _ := newTestAuthService()

		userRepo.On("FindByEmail", mock.Anything, "missing@example.com").Return(nil, nil).Once()

		_, tokens, err := svc.Login(context.Background(), dto.LoginRequestDTO{
			Email:    "missing@example.com",
			Password: "password123",
		})

		assert.ErrorIs(t, err, services.ErrInvalidCredentials)
		assert.Nil(t, tokens)
	})
}

func TestAuthService_Me(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, userRepo, _, _, _, _ := newTestAuthService()

		user := &domain.User{ID: bson.NewObjectID(), Email: "me@example.com"}

		userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil).Once()

		result, err := svc.Me(context.Background(), user.ID.Hex())

		assert.NoError(t, err)
		assert.Equal(t, user.Email, result.Email)
		userRepo.AssertExpectations(t)
	})

	t.Run("invalid user id", func(t *testing.T) {
		svc, _, _, _, _, _ := newTestAuthService()

		_, err := svc.Me(context.Background(), "not-a-valid-id")
		assert.Error(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		svc, userRepo, _, _, _, _ := newTestAuthService()

		userID := bson.NewObjectID()
		userRepo.On("FindByID", mock.Anything, userID).Return(nil, nil).Once()

		_, err := svc.Me(context.Background(), userID.Hex())
		assert.ErrorIs(t, err, services.ErrUserNotFound)
	})
}

func TestAuthService_Refresh(t *testing.T) {
	t.Run("successful refresh", func(t *testing.T) {
		svc, userRepo, _, membershipRepo, _, jwtManager := newTestAuthService()

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
		svc, _, _, _, _, jwtManager := newTestAuthService()

		jwtManager.On("ValidateRefreshToken", "invalid.token").Return("", services.ErrInvalidCredentials).Once()

		_, err := svc.Refresh(context.Background(), "invalid.token")
		assert.Error(t, err)
		jwtManager.AssertExpectations(t)
	})

	t.Run("user not found after valid token", func(t *testing.T) {
		svc, userRepo, _, _, _, jwtManager := newTestAuthService()
		userID := bson.NewObjectID()

		jwtManager.On("ValidateRefreshToken", "valid.but.ghost").Return(userID.Hex(), nil).Once()
		userRepo.On("FindByID", mock.Anything, userID).Return(nil, nil).Once()

		_, err := svc.Refresh(context.Background(), "valid.but.ghost")
		assert.ErrorIs(t, err, services.ErrUserNotFound)
	})
}

func TestAuthService_OAuth(t *testing.T) {
	t.Skip("OAuth requires Auth0 HTTP client abstraction — add ports.OAuthProvider interface")
}

func TestAuthService_SetVerificationHash(t *testing.T) {
	userID := bson.NewObjectID()

	t.Run("success", func(t *testing.T) {
		svc, userRepo, _, _, _, _ := newTestAuthService()

		user := &domain.User{ID: userID, Email: "vault@example.com"}
		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil).Once()

		var stored domain.VaultMetadata
		userRepo.On("UpdateVaultMetadata", mock.Anything, userID, mock.MatchedBy(func(v domain.VaultMetadata) bool {
			stored = v
			return v.VerificationHash == "hash123" && v.KDF.Algorithm == "scrypt" && v.KDF.Salt == "salt" && v.Version == 1
		})).Return(nil).Once()

		req := dto.SetVerificationHashRequest{
			VerificationHash: "hash123",
			KDFParams: dto.KDFDTO{
				Algorithm:   "scrypt",
				Salt:        "salt",
				Memory:      128,
				Iterations:  17,
				Parallelism: 1,
			},
		}

		result, err := svc.SetVerificationHash(context.Background(), userID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "hash123", stored.VerificationHash)
		userRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		svc, userRepo, _, _, _, _ := newTestAuthService()

		userRepo.On("FindByID", mock.Anything, userID).Return(nil, nil).Once()

		_, err := svc.SetVerificationHash(context.Background(), userID, dto.SetVerificationHashRequest{})
		assert.ErrorIs(t, err, services.ErrUserNotFound)
	})
}

func TestAuthService_GetKDFParams(t *testing.T) {
	userID := bson.NewObjectID()

	t.Run("success", func(t *testing.T) {
		svc, userRepo, _, _, _, _ := newTestAuthService()

		user := &domain.User{
			ID: userID,
			Vault: &domain.VaultMetadata{
				Version:          1,
				VerificationHash: "hash",
				KDF: domain.KDFParams{
					Algorithm:   "argon2id",
					Salt:        "somesalt",
					Memory:      65536,
					Iterations:  3,
					Parallelism: 4,
				},
			},
		}
		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil).Once()

		kdf, err := svc.GetKDFParams(context.Background(), userID)

		assert.NoError(t, err)
		assert.Equal(t, "argon2id", kdf.Algorithm)
		assert.Equal(t, "somesalt", kdf.Salt)
		userRepo.AssertExpectations(t)
	})

	t.Run("vault not set up", func(t *testing.T) {
		svc, userRepo, _, _, _, _ := newTestAuthService()

		user := &domain.User{ID: userID}
		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil).Once()

		_, err := svc.GetKDFParams(context.Background(), userID)
		assert.ErrorIs(t, err, services.ErrVaultNotFound)
	})

	t.Run("user not found", func(t *testing.T) {
		svc, userRepo, _, _, _, _ := newTestAuthService()

		userRepo.On("FindByID", mock.Anything, userID).Return(nil, nil).Once()

		_, err := svc.GetKDFParams(context.Background(), userID)
		assert.ErrorIs(t, err, services.ErrUserNotFound)
	})
}

func TestAuthService_UpdateEmail(t *testing.T) {
	userID := bson.NewObjectID()

	t.Run("success", func(t *testing.T) {
		svc, userRepo, identityRepo, membershipRepo, _, jwtManager := newTestAuthService()

		user := &domain.User{ID: userID, Email: "old@example.com"}
		identity := domain.Identity{ID: bson.NewObjectID(), UserID: userID, AuthMethod: domain.AuthMethodLocal}

		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil).Once()
		identityRepo.On("FindByUserID", mock.Anything, userID).Return([]domain.Identity{identity}, nil).Once()
		userRepo.On("EmailExists", mock.Anything, "new@example.com").Return(false, nil).Once()
		userRepo.On("UpdateEmail", mock.Anything, userID, "new@example.com").Return(nil).Once()
		membershipRepo.On("FindByUser", mock.Anything, userID).Return([]domain.Membership{}, nil).Once()
		jwtManager.On("GeneratePair", mock.Anything).Return(&dto.TokenPair{AccessToken: "new.acc", RefreshToken: "new.ref"}, nil).Once()

		result, tokens, err := svc.UpdateEmail(context.Background(), userID, dto.UpdateEmailRequest{Email: "new@example.com"})

		assert.NoError(t, err)
		assert.Equal(t, "new@example.com", result.Email)
		assert.Equal(t, "new.acc", tokens.AccessToken)
		userRepo.AssertExpectations(t)
		identityRepo.AssertExpectations(t)
		membershipRepo.AssertExpectations(t)
		jwtManager.AssertExpectations(t)
	})

	t.Run("email already in use", func(t *testing.T) {
		svc, userRepo, identityRepo, _, _, _ := newTestAuthService()

		user := &domain.User{ID: userID, Email: "old@example.com"}
		identity := domain.Identity{ID: bson.NewObjectID(), UserID: userID, AuthMethod: domain.AuthMethodLocal}

		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil).Once()
		identityRepo.On("FindByUserID", mock.Anything, userID).Return([]domain.Identity{identity}, nil).Once()
		userRepo.On("EmailExists", mock.Anything, "taken@example.com").Return(true, nil).Once()

		_, _, err := svc.UpdateEmail(context.Background(), userID, dto.UpdateEmailRequest{Email: "taken@example.com"})
		assert.ErrorIs(t, err, services.ErrEmailInUse)
	})

	t.Run("oauth only user rejected", func(t *testing.T) {
		svc, userRepo, identityRepo, _, _, _ := newTestAuthService()

		user := &domain.User{ID: userID, Email: "old@example.com"}
		identity := domain.Identity{ID: bson.NewObjectID(), UserID: userID, AuthMethod: domain.AuthMethodGoogle}

		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil).Once()
		identityRepo.On("FindByUserID", mock.Anything, userID).Return([]domain.Identity{identity}, nil).Once()

		_, _, err := svc.UpdateEmail(context.Background(), userID, dto.UpdateEmailRequest{Email: "new@example.com"})
		assert.ErrorIs(t, err, services.ErrCannotEditOAuth)
	})
}

func TestAuthService_UpdateAccountPassword(t *testing.T) {
	userID := bson.NewObjectID()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)

	t.Run("success", func(t *testing.T) {
		svc, userRepo, identityRepo, _, _, _ := newTestAuthService()

		user := &domain.User{ID: userID, Email: "pw@example.com"}
		identity := domain.Identity{ID: bson.NewObjectID(), UserID: userID, AuthMethod: domain.AuthMethodLocal, PasswordHash: string(hash)}

		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil).Once()
		identityRepo.On("FindByUserID", mock.Anything, userID).Return([]domain.Identity{identity}, nil).Once()
		identityRepo.On("UpdatePassword", mock.Anything, userID, mock.AnythingOfType("string")).Return(nil).Once()

		err := svc.UpdateAccountPassword(context.Background(), userID, dto.UpdateAccountPasswordRequest{
			CurrentPassword: "correct-password",
			NewPassword:     "new-password-123",
		})

		assert.NoError(t, err)
		userRepo.AssertExpectations(t)
		identityRepo.AssertExpectations(t)
	})

	t.Run("wrong current password", func(t *testing.T) {
		svc, userRepo, identityRepo, _, _, _ := newTestAuthService()

		user := &domain.User{ID: userID, Email: "pw@example.com"}
		identity := domain.Identity{ID: bson.NewObjectID(), UserID: userID, AuthMethod: domain.AuthMethodLocal, PasswordHash: string(hash)}

		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil).Once()
		identityRepo.On("FindByUserID", mock.Anything, userID).Return([]domain.Identity{identity}, nil).Once()

		err := svc.UpdateAccountPassword(context.Background(), userID, dto.UpdateAccountPasswordRequest{
			CurrentPassword: "wrong-password",
			NewPassword:     "new-password-123",
		})

		assert.ErrorIs(t, err, services.ErrInvalidPassword)
	})

	t.Run("oauth only user rejected", func(t *testing.T) {
		svc, userRepo, identityRepo, _, _, _ := newTestAuthService()

		user := &domain.User{ID: userID, Email: "pw@example.com"}
		identity := domain.Identity{ID: bson.NewObjectID(), UserID: userID, AuthMethod: domain.AuthMethodGoogle}

		userRepo.On("FindByID", mock.Anything, userID).Return(user, nil).Once()
		identityRepo.On("FindByUserID", mock.Anything, userID).Return([]domain.Identity{identity}, nil).Once()

		err := svc.UpdateAccountPassword(context.Background(), userID, dto.UpdateAccountPasswordRequest{})
		assert.ErrorIs(t, err, services.ErrCannotEditOAuth)
	})
}
