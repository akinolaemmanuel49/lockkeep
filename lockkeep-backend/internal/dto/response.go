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
	SystemRole        string                `json:"systemRole"`
	HasMasterPassword bool                  `json:"hasMasterPassword"`
	AuthMethod        string                `json:"authMethod,omitempty"`
	Vault             *domain.VaultMetadata `json:"vault,omitempty"`
}

// SerializeUser maps a domain user into the API response shape, deriving
// hasMasterPassword from the presence of a verification hash (zero-knowledge:
// the server never sees the master key). systemRole mirrors what is embedded
// in the user's JWT claims so the client can render role-gated surfaces.
func SerializeUser(u *domain.User, authMethod domain.AuthMethod, systemRole string) *UserResponse {
	if u == nil {
		return nil
	}
	return &UserResponse{
		ID:                u.ID.Hex(),
		Email:             u.Email,
		TenantID:          u.ID.Hex(),
		SystemRole:        systemRole,
		HasMasterPassword: u.Vault != nil && u.Vault.VerificationHash != "",
		AuthMethod:        string(authMethod),
		Vault:             u.Vault,
	}
}
