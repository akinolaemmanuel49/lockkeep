package dto

type VaultUpdate struct {
	ID                string `json:"id" binding:"required"`
	Organization      string `json:"organization" binding:"required"`
	SiteURL           string `json:"siteUrl" binding:"required"`
	Identifier        string `json:"identifier" binding:"required"`
	Notes             string `json:"notes,omitempty"`
	EncryptedPassword string `json:"encryptedPassword" binding:"required"`
	IV                string `json:"iv" binding:"required"`
	Tag               string `json:"tag" binding:"required"`
}
type MigrateVaultRequest struct {
	ExpectedVersion  uint32 `json:"expected_version" binding:"required"`
	VerificationHash string `json:"verification_hash" binding:"required"`
	KDFParams        struct {
		Algorithm   string `json:"algorithm"`
		Salt        string `json:"salt"`
		Memory      uint32 `json:"memory"`
		Iterations  uint32 `json:"iterations"`
		Parallelism uint8  `json:"parallelism"`
	} `json:"kdf_params" binding:"required"`
	VaultUpdates []VaultUpdate `json:"vault_updates" binding:"required"`
}

type UpdateCredentialRequest struct {
	Organization      string `json:"organization" binding:"required"`
	SiteURL           string `json:"siteUrl" binding:"required"`
	Identifier        string `json:"identifier" binding:"required"`
	Notes             string `json:"notes,omitempty"`
	EncryptedPassword string `json:"encryptedPassword" binding:"required"`
	IV                string `json:"iv" binding:"required"`
	Tag               string `json:"tag" binding:"required"`
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
	KDFParams        struct {
		Algorithm   string `json:"algorithm"`
		Salt        string `json:"salt"`
		Memory      uint32 `json:"memory"`
		Iterations  uint32 `json:"iterations"`
		Parallelism uint8  `json:"parallelism"`
	} `json:"kdf_params" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateCredentialRequest struct {
	Organization      string `bson:"organization" json:"organization"`
	SiteURL           string `bson:"site_url" json:"siteUrl"`
	Identifier        string `bson:"identifier" json:"identifier"`
	Notes             string `bson:"notes" json:"notes"`
	EncryptedPassword string `bson:"encrypted_password" json:"encryptedPassword"`
	IV                string `bson:"iv" json:"iv"`
	Tag               string `bson:"tag" json:"tag"`
}

type VerifyVaultPasswordRequest struct {
	VerificationHash string `json:"verification_hash" binding:"required"`
}
