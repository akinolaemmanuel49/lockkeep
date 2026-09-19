package mocks

import (
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/pkg/jwt"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) GeneratePair(claims jwt.Claims) (*dto.TokenPair, error) {
	args := m.Called(claims)
	return args.Get(0).(*dto.TokenPair), args.Error(1)
}

func (m *MockJWTManager) ValidateAccessToken(token string) (*jwt.Claims, error) {
	args := m.Called(token)
	return args.Get(0).(*jwt.Claims), args.Error(1)
}

func (m *MockJWTManager) ValidateRefreshToken(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}

func (m *MockJWTManager) ParseUserID(userID string) (bson.ObjectID, error) {
	args := m.Called(userID)
	return args.Get(0).(bson.ObjectID), args.Error(1)
}
