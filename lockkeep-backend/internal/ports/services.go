package ports

import (
	"context"
	"time"

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

type EnvironmentService interface {
	CreateApplication(ctx context.Context, userID, orgID bson.ObjectID, name, slug string, description *string) (*domain.Application, error)
	GetApplicationBySlug(ctx context.Context, orgID bson.ObjectID, slug string) (*domain.Application, error)
	ListApplications(ctx context.Context, orgID bson.ObjectID) ([]domain.Application, error)
	UpdateApplication(ctx context.Context, userID, orgID, appID bson.ObjectID, update dto.UpdateApplicationDTO) (*domain.Application, error)
	DeleteApplication(ctx context.Context, userID, orgID, appID bson.ObjectID) error
	CreateEnvironment(ctx context.Context, userID, orgID, appID bson.ObjectID, name, slug string, isProtected bool) (*domain.Environment, error)
	GetEnvironmentBySlug(ctx context.Context, orgID, appID bson.ObjectID, slug string) (*domain.Environment, error)
	ListEnvironments(ctx context.Context, appID bson.ObjectID) ([]domain.Environment, error)
	UpdateEnvironment(ctx context.Context, userID, orgID, envID bson.ObjectID, update dto.UpdateEnvironmentDTO) (*domain.Environment, error)
	DeleteEnvironment(ctx context.Context, userID, orgID, envID bson.ObjectID) error
}

type EnvSecretService interface {
	UpsertSecret(ctx context.Context, actor bson.ObjectID, orgID, environmentID bson.ObjectID, secret dto.SecretInput) (*dto.EnvSecretDTO, error)
	ListSecrets(ctx context.Context, actor bson.ObjectID, environmentID bson.ObjectID) ([]dto.EnvSecretDTO, error)
	GetSecret(ctx context.Context, actor bson.ObjectID, environmentID bson.ObjectID, key string) (*dto.EnvSecretDTO, error)
	DeleteSecret(ctx context.Context, actor bson.ObjectID, orgID, environmentID bson.ObjectID, key string) error
	ExportEnvelope(ctx context.Context, actorID bson.ObjectID, environmentID bson.ObjectID, bearerSecret []byte) (*dto.ExportEnvelope, error)
}

type ServiceAccountService interface {
	CreateServiceAccount(ctx context.Context, userID, orgID bson.ObjectID, name string, description *string) (*domain.ServiceAccount, error)
	ListServiceAccounts(ctx context.Context, orgID bson.ObjectID) ([]domain.ServiceAccount, error)
	DeleteServiceAccount(ctx context.Context, userID, orgID, serviceAccountID bson.ObjectID) error
	MintApiKey(ctx context.Context, userID, orgID, serviceAccountID, environmentID bson.ObjectID, scopes []domain.ApiKeyScope, expiresAt *time.Time) (*domain.ApiKey, string, error)
	ListApiKeys(ctx context.Context, orgID, serviceAccountID bson.ObjectID) ([]domain.ApiKey, error)
	RevokeApiKey(ctx context.Context, userID, orgID, keyID bson.ObjectID) error
	VerifyApiKey(ctx context.Context, rawKey string) (*domain.ApiKey, *domain.ServiceAccount, error)
}
