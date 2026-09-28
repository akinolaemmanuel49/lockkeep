package ports

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuthService interface {
	Register(ctx context.Context, input dto.RegisterRequestDTO) (*dto.UserResponse, *dto.TokenPair, error)
	Login(ctx context.Context, input dto.LoginRequestDTO) (*dto.UserResponse, *dto.TokenPair, error)
	OAuth(ctx context.Context, accessToken string) (*dto.UserResponse, *dto.TokenPair, bool, error)
	Refresh(ctx context.Context, refreshToken string) (*dto.TokenPair, error)
	SetVerificationHash(ctx context.Context, userID bson.ObjectID, req dto.SetVerificationHashRequest) (*domain.User, error)
	GetKDFParams(ctx context.Context, userID bson.ObjectID) (*domain.KDFParams, error)
	UpdateEmail(ctx context.Context, userID bson.ObjectID, req dto.UpdateEmailRequest) (*dto.UserResponse, *dto.TokenPair, error)
	UpdateAccountPassword(ctx context.Context, userID bson.ObjectID, req dto.UpdateAccountPasswordRequest) error
}

type UserService interface {
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	UpdateUser(ctx context.Context, userID string, input dto.UpdateUserProfileRequestDTO) (*domain.User, error)
}

type OrganizationService interface {
	CreateOrganization(ctx context.Context, userID bson.ObjectID, name, slug string) (*domain.Organization, error)
	GetOrganizationBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	ListOrganizationsByUser(ctx context.Context, userID bson.ObjectID) ([]domain.Organization, error)
	UpdateOrganization(ctx context.Context, userID bson.ObjectID, orgID bson.ObjectID, update dto.UpdateOrganizationDTO) (*domain.Organization, error)
	DeleteOrganizations(ctx context.Context, userID bson.ObjectID, orgIDs []bson.ObjectID) error
}

type TeamService interface {
	CreateTeam(ctx context.Context, userID bson.ObjectID, orgID bson.ObjectID, name, slug string) (*domain.Team, error)
	GetTeamBySlug(ctx context.Context, orgID bson.ObjectID, slug string) (*domain.Team, error)
	ListTeamsByOrganization(ctx context.Context, orgID bson.ObjectID) ([]domain.Team, error)
	UpdateTeam(ctx context.Context, orgID bson.ObjectID, teamID bson.ObjectID) (*domain.Team, error)
	DeleteTeams(ctx context.Context, userID bson.ObjectID, orgID bson.ObjectID, teamIDs []bson.ObjectID) error
}

type VaultService interface {
	Me(ctx context.Context, vaultID string) (*domain.Vault, error)
	GetVaultByUserID(ctx context.Context, userID string) (*domain.Vault, error)
	VerifyVaultPassword(ctx context.Context, userID string, verificationHash string) (bool, error)
	Update(ctx context.Context, vaultID string, input dto.UpdateVaultProfileRequestDTO) (*domain.Vault, error)
}

type VaultItemService interface {
	CreateVaultItem(ctx context.Context, userID bson.ObjectID, vault *domain.VaultItem) (*domain.VaultItem, error)
	ListVaultItems(ctx context.Context, userID bson.ObjectID) ([]domain.VaultItem, error)
	GetVaultItemByID(ctx context.Context, userID bson.ObjectID, vaultItemID bson.ObjectID) (*domain.VaultItem, error)
	UpdateVaultItem(ctx context.Context, userID bson.ObjectID, vaultItemID bson.ObjectID, update dto.VaultItemUpdate) (*domain.VaultItem, error)
	DeleteVaultItems(ctx context.Context, userID bson.ObjectID, vaultItemIDs []bson.ObjectID) error
}

type CryptoPolicyService interface {
	GetCurrentPolicy(ctx context.Context) (*domain.CryptoPolicy, error)
	SetCurrentPolicy(ctx context.Context, policy *dto.SetCurrentPolicy) error
}
