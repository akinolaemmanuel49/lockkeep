package service

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/config"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	user_errors "github.com/akinolaemmanuel49/lockkeep-backend/internal/errors/user"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type UserService struct {
	cfg      *config.Config
	userRepo ports.UserRepository
}

func NewUserService(cfg *config.Config, userRepo ports.UserRepository) *UserService {
	return &UserService{cfg: cfg, userRepo: userRepo}
}

func (s *UserService) Me(ctx context.Context, userID string) (*domain.User, error) {
	objID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, user_errors.ErrUserNotFound
	}
	return user, nil
}
