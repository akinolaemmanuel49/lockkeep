package ports

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
)

type AuthService interface {
	Register(ctx context.Context, input dto.RegisterRequestDTO) (*domain.User, error)
	Login(ctx context.Context, input dto.LoginRequestDTO) (*domain.User, *dto.TokenPair, error)
	OAuth(ctx context.Context, accessToken string) (*domain.User, *dto.TokenPair, bool, error)
	Refresh(ctx context.Context, refreshToken string) (*dto.TokenPair, error)
}

type UserService interface {
	Me(ctx context.Context, userID string) (*domain.User, error)
}
