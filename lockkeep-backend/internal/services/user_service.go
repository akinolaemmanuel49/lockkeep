package services

import (
	"context"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/config"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
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
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) Update(ctx context.Context, userID string, input dto.UpdateUserProfileRequestDTO) (*domain.User, error) {
	objID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if (input.Username == nil || *input.Username == "") && input.AvatarURL == nil {
		return nil, ErrInvalidProfileUpdate
	}

	if input.Username != nil && *input.Username != "" {
		user.Username = *input.Username
	}

	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.UpdateProfile(ctx, user.ID, *user); err != nil {
		return nil, err
	}

	return user, nil
}
