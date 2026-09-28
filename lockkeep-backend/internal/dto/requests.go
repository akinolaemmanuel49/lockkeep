package dto

import (
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type VaultItemUpdate struct {
	Type     *VaultItemType `json:"type" binding:"omitempty,oneof=login environment ssh_key secure_note payment_card api_key"`
	Name     *string        `json:"name" binding:"omitempty"`
	Secret   *SecretDTO     `json:"secret" binding:"omitempty"`
	Metadata *bson.M        `json:"metadata,omitempty" binding:"omitempty"`
}

type MigrateVaultRequest struct {
	ExpectedVersion  uint32            `json:"expected_version" binding:"required"`
	VerificationHash string            `json:"verification_hash" binding:"required"`
	KDFParams        KDFDTO            `json:"kdf_params" binding:"required"`
	VaultUpdates     []VaultItemUpdate `json:"vault_updates" binding:"required"`
}

type CreateVaultItemRequest struct {
	Type     VaultItemType  `json:"type" binding:"required,oneof=login environment ssh_key secure_note payment_card api_key"`
	Name     string         `json:"name" binding:"required"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Secret   SecretDTO      `json:"secret" binding:"required"`
}

type UpdateVaultItemRequest struct {
	Type     VaultItemType  `json:"type" binding:"required,oneof=login environment ssh_key secure_note payment_card api_key"`
	Name     string         `json:"name" binding:"required"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Secret   SecretDTO      `json:"secret" binding:"required"`
}

type UpdateOrganizationDTO struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
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

type UpdateVaultProfileRequestDTO struct {
	Vaultname *string `json:"vaultname,omitempty"`
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
