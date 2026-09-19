package dto

import "github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"

type Auth0UserInfo struct {
	Sub      string `json:"sub"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

type OauthAuthorizeResponse struct {
	Method     domain.AuthMethod `json:"method" binding:"required,oneof=local oauth_google oauth_github"`
	ProviderID string            `json:"providerID" binding:"required"`
	Email      string            `json:"email" binding:"required,email"`
	Username   string            `json:"username" binding:"required"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	ID                string                `json:"id"`
	Email             string                `json:"email"`
	TenantID          string                `json:"tenantId"`
	HasMasterPassword bool                  `json:"hasMasterPassword"`
	AuthMethod        string                `json:"authMethod,omitempty"`
	Vault             *domain.VaultMetadata `json:"vault,omitempty"`
}
