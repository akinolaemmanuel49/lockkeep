package dto

import "github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"

type VaultItemUpdate struct {
	ID     string        `json:"id" binding:"required"`
	Type   VaultItemType `json:"type" binding:"required"`
	Name   string        `json:"name" binding:"required"`
	Secret SecretDTO     `json:"secret" binding:"required"`
}

type MigrateVaultRequest struct {
	ExpectedVersion  uint32            `json:"expected_version" binding:"required"`
	VerificationHash string            `json:"verification_hash" binding:"required"`
	KDFParams        KDFDTO            `json:"kdf_params" binding:"required"`
	VaultUpdates     []VaultItemUpdate `json:"vault_updates" binding:"required"`
}

type CreateVaultItemRequest struct {
	Type     VaultItemType  `json:"type" binding:"required"`
	Name     string         `json:"name" binding:"required"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Secret   SecretDTO      `json:"secret" binding:"required"`
}

type UpdateVaultItemRequest struct {
	Type     VaultItemType  `json:"type" binding:"required"`
	Name     string         `json:"name" binding:"required"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Secret   SecretDTO      `json:"secret" binding:"required"`
}

type UpdateAccountPasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
}

type UpdateEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type SetVerificationHashRequest struct {
	VerificationHash string `json:"verification_hash" binding:"required"`
	KDFParams        KDFDTO `json:"kdf_params" binding:"required"`
}

type RegisterRequestDTO struct {
	Username string `json:"username" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequestDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type OauthAuthorizeDTO struct {
	Method     domain.AuthMethod `json:"method" binding:"required,oneof=local oauth_google oauth_github oauth_auth0"`
	ProviderID string            `json:"providerID" binding:"required"`
	Email      string            `json:"email" binding:"required,email"`
	Username   string            `json:"username" binding:"required"`
}

type UpdateUserProfileRequestDTO struct {
	Username  *string `json:"username,omitempty"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
}

type VerifyVaultPasswordRequest struct {
	VerificationHash string `json:"verification_hash" binding:"required"`
}

type SecretDTO struct {
	Ciphertext string `json:"ciphertext" binding:"required"`
	IV         string `json:"iv" binding:"required"`
	Tag        string `json:"tag" binding:"required"`
	Version    uint32 `json:"version,omitempty"`
}

type KDFDTO struct {
	Algorithm   string `json:"algorithm"`
	Salt        string `json:"salt"`
	Memory      uint32 `json:"memory"`
	Iterations  uint32 `json:"iterations"`
	Parallelism uint8  `json:"parallelism"`
}

type SetCurrentPolicy struct {
	KDFParams struct {
		Algorithm   string `json:"algorithm"`
		Memory      uint32 `json:"memory"`
		Iterations  uint32 `json:"iterations"`
		Parallelism uint8  `json:"parallelism"`
	} `json:"kdfParams"`
}

type VaultItemType string

const (
	ItemLogin       VaultItemType = "login"
	ItemEnvironment VaultItemType = "environment"
	ItemSSHKey      VaultItemType = "ssh_key"
	ItemSecureNote  VaultItemType = "secure_note"
	ItemCard        VaultItemType = "payment_card"
	ItemAPIKey      VaultItemType = "api_key"
)
