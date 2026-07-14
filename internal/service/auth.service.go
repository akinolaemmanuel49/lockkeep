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
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/infrastructure/persistence/mongo"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/akinolaemmanuel49/lockkeep-backend/pkg/jwt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already registered")
	ErrUserNotFound       = errors.New("user not found")
	ErrOAuthEmailRequired = errors.New("a public email address is required")
)

type AuthService struct {
	cfg            *config.Config
	userRepo       ports.UserRepository
	identityRepo   ports.IdentityRepository
	membershipRepo ports.MembershipRepository
	unitOfWork     mongo.UnitOfWork
	jwtManager     *jwt.Manager
}

func NewAuthService(
	cfg *config.Config,
	userRepo ports.UserRepository,
	identityRepo ports.IdentityRepository,
	membershipRepo ports.MembershipRepository,
	unitOfWork mongo.UnitOfWork,
	jwtManager *jwt.Manager,
) *AuthService {
	return &AuthService{
		cfg:            cfg,
		userRepo:       userRepo,
		identityRepo:   identityRepo,
		membershipRepo: membershipRepo,
		unitOfWork:     unitOfWork,
		jwtManager:     jwtManager,
	}
}

// Register creates a new user with local authentication
func (s *AuthService) Register(ctx context.Context, input dto.RegisterRequestDTO) (*domain.User, error) {
	var createdUser *domain.User

	err := s.unitOfWork.Within(ctx, func(txCtx context.Context) error {
		exists, err := s.userRepo.EmailExists(txCtx, input.Email)
		if err != nil {
			return err
		}
		if exists {
			return ErrEmailTaken
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}

		user := &domain.User{
			ID:       bson.NewObjectID(),
			Username: input.Username,
			Email:    input.Email,
		}

		if err := s.userRepo.Create(txCtx, user); err != nil {
			return err
		}
		// Feels weird, might be better off manually passing timestamps
		createdUser, err = s.userRepo.FindByID(ctx, user.ID)
		if err != nil {
			return err
		}

		identity := &domain.Identity{
			ID:           bson.NewObjectID(),
			UserID:       user.ID,
			AuthMethod:   domain.AuthMethodLocal,
			PasswordHash: string(hash),
		}

		if err := s.identityRepo.Create(txCtx, identity); err != nil {
			return fmt.Errorf("create identity: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

// Login validates credentials and returns user + token pair
func (s *AuthService) Login(ctx context.Context, input dto.LoginRequestDTO) (*domain.User, *jwt.TokenPair, error) {
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, ErrInvalidCredentials
	}

	identities, err := s.identityRepo.FindByUserID(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	var localIdentity *domain.Identity
	for i := range identities {
		if identities[i].AuthMethod == domain.AuthMethodLocal {
			localIdentity = &identities[i]
			break
		}
	}

	if localIdentity == nil {
		return nil, nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(localIdentity.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	_ = s.identityRepo.RecordLogin(ctx, localIdentity.ID)

	claims := s.buildClaims(ctx, user)
	tokens, err := s.jwtManager.GeneratePair(claims)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) Me(ctx context.Context, userID string) (*domain.User, error) {
	userObjID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	return s.userRepo.FindByID(ctx, userObjID)

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

func extractProvider(sub string) (domain.AuthMethod, error) {
	if strings.HasPrefix(sub, "google-oauth2|") {
		return "oauth_google", nil
	}
	if strings.HasPrefix(sub, "github|") {
		return "oauth_github", nil
	}
	if strings.HasPrefix(sub, "auth0|") {
		return "oauth_auth0", nil
	}
	return "", fmt.Errorf("invalid provider")
}

// OAuth handles OAuth registration/login
func (s *AuthService) OAuth(ctx context.Context, accessToken string) (*domain.User, *jwt.TokenPair, bool, error) {
	userInfo, err := s.getAuth0UserInfo(accessToken)
	if err != nil {
		return nil, nil, false, err
	}

	provider, err := extractProvider(userInfo.Sub)
	if err != nil {
		return nil, nil, false, err
	}

	identity, err := s.identityRepo.FindByOAuth(ctx, domain.AuthMethod(provider), userInfo.Sub)
	if err != nil {
		return nil, nil, false, err
	}

	input := dto.OauthAuthorizeDTO{
		Method:     provider,
		ProviderID: userInfo.Sub,
		Email:      userInfo.Email,
		Username:   userInfo.Nickname,
	}

	var user *domain.User
	var isNewUser bool
	var tokens *jwt.TokenPair

	if identity != nil {
		user, err = s.userRepo.FindByID(ctx, identity.UserID)
		if err != nil {
			return nil, nil, false, err
		}

		_ = s.identityRepo.RecordLogin(ctx, identity.ID)

		isNewUser = false

		claims := s.buildClaims(ctx, user)
		tokens, err = s.jwtManager.GeneratePair(claims)
		if err != nil {
			return nil, nil, isNewUser, err
		}

	} else {
		user, err = s.registerOAuthUser(ctx, input)
		if err != nil {
			return nil, nil, false, err
		}

		isNewUser = true
	}

	return user, tokens, isNewUser, nil
}

func (s *AuthService) registerOAuthUser(ctx context.Context, input dto.OauthAuthorizeDTO) (*domain.User, error) {
	var createdOauthUser *domain.User

	err := s.unitOfWork.Within(ctx, func(txCtx context.Context) error {
		existing, err := s.userRepo.FindByEmail(txCtx, input.Email)
		if err != nil {
			return err
		}

		var userID bson.ObjectID
		if existing != nil {
			userID = existing.ID
		} else {
			user := &domain.User{
				ID:       bson.NewObjectID(),
				Username: input.Username,
				Email:    input.Email,
			}
			if err := s.userRepo.Create(txCtx, user); err != nil {
				return err
			}
			userID = user.ID
			createdOauthUser = user
		}

		identity := &domain.Identity{
			ID:             bson.NewObjectID(),
			UserID:         userID,
			AuthMethod:     input.Method,
			AuthProviderID: &input.ProviderID,
		}

		if err := s.identityRepo.Create(txCtx, identity); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdOauthUser, nil
}

func (s *AuthService) buildClaims(ctx context.Context, user *domain.User) jwt.Claims {
	role := domain.RoleSystemUser

	memberships, _ := s.membershipRepo.FindByUser(ctx, user.ID)

	orgClaims := make([]jwt.OrgClaim, len(memberships))
	for i, m := range memberships {
		orgClaims[i] = jwt.OrgClaim{
			OrgID:     m.OrganizationID.Hex(),
			Role:      string(m.RoleID),
			TeamRoles: extractTeamRoles(m.TeamRoles),
		}
	}

	return jwt.Claims{
		UserID:     user.ID.Hex(),
		Email:      user.Email,
		SystemRole: string(role),
		Orgs:       orgClaims,
	}
}

func extractTeamRoles(roles []domain.TeamRole) map[string]string {
	result := make(map[string]string, len(roles))
	for _, r := range roles {
		result[r.TeamID.Hex()] = string(r.RoleID)
	}
	return result
}

// Refresh rotates tokens using refresh token
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	userID, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	objUserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, objUserID)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	claims := s.buildClaims(ctx, user)
	return s.jwtManager.GeneratePair(claims)
}
