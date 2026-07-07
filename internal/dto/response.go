package dto

import "github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"

type Auth0UserInfo struct {
	Sub      string `json:"sub"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	ID                string                `json:"id"`
	Email             string                `json:"email"`
	TenantID          string                `json:"tenantId"`
	UserType          domain.UserType       `json:"userType"`
	HasMasterPassword bool                  `json:"hasMasterPassword"`
	AuthMethod        string                `json:"authMethod,omitempty"`
	Vault             *domain.VaultMetadata `json:"vault,omitempty"`
}
