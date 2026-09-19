package ports

import (
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/pkg/jwt"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type JWTManager interface {
	GeneratePair(claims jwt.Claims) (*dto.TokenPair, error)
	ValidateAccessToken(tokenString string) (*jwt.Claims, error)
	ValidateRefreshToken(tokenString string) (string, error)
	ParseUserID(userID string) (bson.ObjectID, error)
}
