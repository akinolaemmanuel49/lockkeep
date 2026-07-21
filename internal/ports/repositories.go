package ports

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// IndexManager is implemented by repositories that need schema setup
type IndexManager interface {
	EnsureIndexes(ctx context.Context) error
}

type UserRepository interface {
	IndexManager
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.User, error)
	UpdateVaultMetadata(ctx context.Context, userID bson.ObjectID, vault domain.VaultMetadata) error
	UpdateEmail(ctx context.Context, userID bson.ObjectID, email string) error
}

type IdentityRepository interface {
	IndexManager
	Create(ctx context.Context, identity *domain.Identity) error
	FindByUserID(ctx context.Context, userID bson.ObjectID) ([]domain.Identity, error)
	FindByOAuth(ctx context.Context, method domain.AuthMethod, providerID string) (*domain.Identity, error)
	UpdatePassword(ctx context.Context, userID bson.ObjectID, passwordHash string) error
	RecordLogin(ctx context.Context, identityID bson.ObjectID) error
}

type OrganizationRepository interface {
	IndexManager
	Create(ctx context.Context, org *domain.Organization) error
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.Organization, error)
	FindBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	FindByMember(ctx context.Context, userID bson.ObjectID) ([]domain.Organization, error)
	Update(ctx context.Context, id bson.ObjectID, update domain.Organization) error
	Delete(ctx context.Context, id bson.ObjectID) error
}

type TeamRepository interface {
	IndexManager
	Create(ctx context.Context, team *domain.Team) error
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.Team, error)
	FindBySlug(ctx context.Context, orgID bson.ObjectID, slug string) (*domain.Team, error)
	FindByOrganization(ctx context.Context, orgID bson.ObjectID) ([]domain.Team, error)
	Update(ctx context.Context, id bson.ObjectID, update domain.Team) error
	Delete(ctx context.Context, id bson.ObjectID) error
}

type MembershipRepository interface {
	IndexManager
	Create(ctx context.Context, membership *domain.Membership) error
	FindByUserAndOrg(ctx context.Context, userID, orgID bson.ObjectID) (*domain.Membership, error)
	FindByUser(ctx context.Context, userID bson.ObjectID) ([]domain.Membership, error)
	FindByOrganization(ctx context.Context, orgID bson.ObjectID) ([]domain.Membership, error)
	UpdateRole(ctx context.Context, userID, orgID bson.ObjectID, roleID domain.RoleID) error
	UpdateTeamRoles(ctx context.Context, userID, orgID bson.ObjectID, teamRoles []domain.TeamRole) error
	Delete(ctx context.Context, userID, orgID bson.ObjectID) error
}

type SharedSecretRepository interface {
	IndexManager
	Create(ctx context.Context, secret *domain.SharedSecret) error
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.SharedSecret, error)
	FindByOrganization(ctx context.Context, orgID bson.ObjectID, teamID *bson.ObjectID) ([]domain.SharedSecret, error)
	Update(ctx context.Context, id bson.ObjectID, update domain.SharedSecret) error
	Delete(ctx context.Context, id bson.ObjectID) error
}

type AuditRepository interface {
	IndexManager
	Log(ctx context.Context, event *domain.AuditEvent) error
	FindByResource(ctx context.Context, resourceType domain.AuditResourceType, resourceID bson.ObjectID, limit int) ([]domain.AuditEvent, error)
	FindByUser(ctx context.Context, userID bson.ObjectID, limit int) ([]domain.AuditEvent, error)
	FindByOrganization(ctx context.Context, orgID bson.ObjectID, limit int) ([]domain.AuditEvent, error)
}

type VaultItemRepository interface {
	IndexManager
	Create(ctx context.Context, item *domain.VaultItem) error
	FindByUser(ctx context.Context, userID bson.ObjectID) ([]domain.VaultItem, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.VaultItem, error)
	Update(ctx context.Context, id bson.ObjectID, item *domain.VaultItem) error
	Delete(ctx context.Context, id bson.ObjectID) error
}

type CryptoPolicyRepository interface {
	IndexManager
	GetCurrent(ctx context.Context) (*domain.CryptoPolicy, error)
	SetCurrent(ctx context.Context, policy *domain.CryptoPolicy) error
}
