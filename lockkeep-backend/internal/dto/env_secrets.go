package dto

import (
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
)

type CreateApplicationRequest struct {
	Name        string  `json:"name" binding:"required"`
	Slug        string  `json:"slug" binding:"required"`
	Description *string `json:"description,omitempty"`
}

type UpdateApplicationDTO struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
}

type CreateEnvironmentRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	IsProtected bool   `json:"isProtected"`
}

type UpdateEnvironmentDTO struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	IsProtected *bool   `json:"isProtected"`
}

type SecretInput struct {
	Key   string            `json:"key" binding:"required"`
	Kind  domain.SecretKind `json:"kind" binding:"required,oneof=env note file json"`
	Value string            `json:"value"`
}

type EnvSecretDTO struct {
	Key       string            `json:"key"`
	Kind      domain.SecretKind `json:"kind"`
	Version   int               `json:"version"`
	Value     string            `json:"value"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

func (s *EnvSecretDTO) WithValue(value string) *EnvSecretDTO {
	copy := *s
	copy.Value = value
	return &copy
}

// ExportEnvelope is the machine-facing payload. It never contains plaintext:
// all secret values are ciphertexts sharing a single wrapped DEK. The client
// derives the wrapping key from its API key via HKDF (see KDF) and unwraps.
type ExportEnvelope struct {
	EnvironmentID string         `json:"environmentId"`
	KDF           KDFSpec        `json:"kdf"`
	WrappedDEK    WrappedKey     `json:"wrappedDek"`
	Secrets       []ExportSecret `json:"secrets"`
}

type KDFSpec struct {
	Algorithm string `json:"algorithm"`
	Salt      string `json:"salt"`
	Info      string `json:"info"`
	Length    int    `json:"length"`
}

type WrappedKey struct {
	KeyID      string `json:"keyId"`
	Algorithm  string `json:"algorithm"`
	Ciphertext []byte `json:"ciphertext"`
	Nonce      []byte `json:"nonce"`
	AAD        string `json:"aad"`
}

type ExportSecret struct {
	Key       string            `json:"key"`
	Kind      domain.SecretKind `json:"kind"`
	Version   int               `json:"version"`
	Ciphertext []byte           `json:"ciphertext"`
	Nonce     []byte            `json:"nonce"`
}

type CreateServiceAccountRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
}

type MintApiKeyRequest struct {
	ApplicationSlug string     `json:"applicationSlug" binding:"required"`
	EnvironmentSlug string     `json:"environmentSlug" binding:"required"`
	Scopes          []string   `json:"scopes" binding:"required"`
	ExpiresAt       *time.Time `json:"expiresAt,omitempty"`
}

type MintedApiKeyResponse struct {
	ApiKey     string            `json:"apiKey"`
	Key        *domain.ApiKey    `json:"key"`
	ServiceAccount *domain.ServiceAccount `json:"serviceAccount"`
}