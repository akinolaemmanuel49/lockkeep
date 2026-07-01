package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/config"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/repository"
	"github.com/akinolaemmanuel49/lockkeep-backend/pkg/jwt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("an account with this email already exists")
	ErrEmailInUse         = errors.New("email already in use")
	ErrUserNotFound       = errors.New("user not found")
	ErrOAuthEmailRequired = errors.New("a public email address is required")
	ErrCannotEditOAuth    = errors.New("cannot edit for oauth accounts")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrPolicyNotFound     = errors.New("no policy not found")
)

type AuthService struct {
	cfg              *config.Config
	userRepo         *repository.UserRepository
	cryptoPolicyRepo *repository.CryptoPolicyRepository
	jwtManager       *jwt.Manager
}

func NewAuthService(cfg *config.Config,
	userRepo *repository.UserRepository,
	cryptoPolicyRepo *repository.CryptoPolicyRepository,
	jwtManager *jwt.Manager) *AuthService {
	return &AuthService{
		cfg:              cfg,
		userRepo:         userRepo,
		cryptoPolicyRepo: cryptoPolicyRepo,
		jwtManager:       jwtManager,
	}
}

// --- OAuth ---

func (s *AuthService) OAuth(ctx context.Context, accessToken string) (*dto.TokenPair, *dto.UserResponse, error) {
	userInfo, err := s.getAuth0UserInfo(accessToken)
	if err != nil {
		return nil, nil, err
	}

	provider := extractProvider(userInfo.Sub)

	user, err := s.userRepo.FindByOAuth(ctx, provider, userInfo.Sub)
	if err != nil {
		return nil, nil, err
	}

	if user == nil {
		if userInfo.Email == "" {
			return nil, nil, ErrOAuthEmailRequired
		}

		existing, err := s.userRepo.FindByEmail(ctx, userInfo.Email)
		if err != nil {
			return nil, nil, err
		}
		if existing != nil {
			return nil, nil, ErrEmailExists
		}

		user = &domain.User{
			Email:          userInfo.Email,
			TenantID:       generateTenantID(),
			AuthMethod:     "oauth_" + provider,
			AuthProviderID: &userInfo.Sub,
			Vault:          nil,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, nil, err
		}
	}

	tokens, err := s.jwtManager.Generate(user)
	if err != nil {
		return nil, nil, err
	}

	return &dto.TokenPair{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, s.toUserResponse(user), nil
}

func (s *AuthService) getAuth0UserInfo(accessToken string) (*dto.Auth0UserInfo, error) {
	req, err := http.NewRequest("GET", "https://"+s.cfg.Auth0Domain+"/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.New("auth0 userinfo failed: " + string(body))
	}

	var userInfo dto.Auth0UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	return &userInfo, nil
}

// --- Local Auth ---

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.TokenPair, *dto.UserResponse, error) {
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, err
	}
	if existing != nil {
		return nil, nil, ErrEmailExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	user := &domain.User{
		Email:        req.Email,
		TenantID:     generateTenantID(),
		AuthMethod:   "local",
		PasswordHash: string(passwordHash),
		Vault:        nil,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	tokens, err := s.jwtManager.Generate(user)
	if err != nil {
		return nil, nil, err
	}

	return &dto.TokenPair{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, s.toUserResponse(user), nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenPair, *dto.UserResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil || user.AuthMethod != "local" {
		return nil, nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	tokens, err := s.jwtManager.Generate(user)
	if err != nil {
		return nil, nil, err
	}

	return &dto.TokenPair{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, s.toUserResponse(user), nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*dto.TokenPair, error) {
	userID, err := s.jwtManager.ValidateRefresh(refreshToken)
	if err != nil {
		return nil, err
	}

	objectID, err := s.jwtManager.ParseUserID(userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, objectID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	tokens, err := s.jwtManager.Generate(user)
	if err != nil {
		return nil, err
	}

	return &dto.TokenPair{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *AuthService) SetVerificationHash(ctx context.Context, userID bson.ObjectID, req dto.SetVerificationHashRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	currentPolicy, err := s.cryptoPolicyRepo.GetCurrent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current crypto policy: %w", err)
	}

	vault := domain.VaultMetadata{
		Version:          currentPolicy.Version,
		VerificationHash: req.VerificationHash,
		KDF: domain.KDFParams{
			Algorithm:   req.KDFParams.Algorithm,
			Salt:        req.KDFParams.Salt,
			Memory:      req.KDFParams.Memory,
			Iterations:  req.KDFParams.Iterations,
			Parallelism: req.KDFParams.Parallelism,
		},
	}

	if err := s.userRepo.UpdateVaultMetadata(ctx, userID, vault); err != nil {
		return nil, err
	}

	user.Vault = &vault
	return s.toUserResponse(user), nil
}

func (s *AuthService) GetKDFParams(ctx context.Context, userID bson.ObjectID) (*domain.KDFParams, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if user.Vault != nil {
		return &user.Vault.KDF, nil
	}

	policy, err := s.cryptoPolicyRepo.GetCurrent(ctx)
	if err != nil {
		return nil, err
	}

	if policy != nil {
		return &policy.KDFParams, nil

	}

	return nil, ErrPolicyNotFound
}

func (s *AuthService) UpdateEmail(ctx context.Context, userID bson.ObjectID, req dto.UpdateEmailRequest) (*dto.TokenPair, *dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, ErrUserNotFound
	}

	if user.AuthMethod != "local" {
		return nil, nil, ErrCannotEditOAuth
	}

	if user.Email == req.Email {
		tokens, err := s.jwtManager.Generate(user)
		if err != nil {
			return nil, nil, err
		}
		return &dto.TokenPair{AccessToken: tokens.AccessToken, RefreshToken: tokens.RefreshToken}, s.toUserResponse(user), nil
	}

	ok, err := s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, nil, err
	}
	if ok {
		return nil, nil, ErrEmailInUse
	}

	if err := s.userRepo.UpdateEmail(ctx, userID, req.Email); err != nil {
		return nil, nil, err
	}

	user, err = s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	tokens, err := s.jwtManager.Generate(user)
	if err != nil {
		return nil, nil, err
	}

	return &dto.TokenPair{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, s.toUserResponse(user), nil
}

func (s *AuthService) UpdateAccountPassword(ctx context.Context, userID bson.ObjectID, req dto.UpdateAccountPasswordRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	if user.AuthMethod != "local" {
		return ErrCannotEditOAuth
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return ErrInvalidPassword
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(ctx, userID, string(passwordHash))
}

func (s *AuthService) toUserResponse(user *domain.User) *dto.UserResponse {
	hasMasterPassword := false
	if user.Vault != nil {
		hasMasterPassword = user.Vault.VerificationHash != ""
	}

	return &dto.UserResponse{
		ID:                user.ID.Hex(),
		Email:             user.Email,
		TenantID:          user.TenantID,
		HasMasterPassword: hasMasterPassword,
		AuthMethod:        user.AuthMethod,
		Vault:             user.Vault,
	}
}

func generateTenantID() string {
	return "tenant_" + bson.NewObjectID().Hex()[:8]
}

func extractProvider(sub string) string {
	if strings.HasPrefix(sub, "google-oauth2|") {
		return "google"
	}
	if strings.HasPrefix(sub, "github|") {
		return "github"
	}
	if strings.HasPrefix(sub, "auth0|") {
		return "local"
	}
	return "oauth"
}
