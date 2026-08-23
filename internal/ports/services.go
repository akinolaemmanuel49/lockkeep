package ports

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuthService interface {
	Register(ctx context.Context, input dto.RegisterRequestDTO) (*domain.User, error)
	Login(ctx context.Context, input dto.LoginRequestDTO) (*domain.User, *dto.TokenPair, error)
	OAuth(ctx context.Context, accessToken string) (*domain.User, *dto.TokenPair, bool, error)
	Refresh(ctx context.Context, refreshToken string) (*dto.TokenPair, error)
}

type UserService interface {
	Me(ctx context.Context, userID string) (*domain.User, error)
	Update(ctx context.Context, userID string, input dto.UpdateUserProfileRequestDTO) (*domain.User, error)
}

type OrganizationService interface {
	Create(ctx context.Context, userID bson.ObjectID, name, slug string) (*domain.Organization, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	ListForUser(ctx context.Context, userID bson.ObjectID) ([]domain.Organization, error)
	Update(ctx context.Context, userID bson.ObjectID, slug string, name string) (*domain.Organization, error)
	Delete(ctx context.Context, userID bson.ObjectID, slug string) error
}

type TeamService interface {
	Create(ctx context.Context, userID bson.ObjectID, orgID bson.ObjectID, name, slug string) (*domain.Team, error)
	GetBySlug(ctx context.Context, orgID bson.ObjectID, slug string) (*domain.Team, error)
	ListByOrganization(ctx context.Context, orgID bson.ObjectID) ([]domain.Team, error)
	Delete(ctx context.Context, userID bson.ObjectID, orgID bson.ObjectID, teamID bson.ObjectID) error
}
