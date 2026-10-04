package services

import "errors"

// Service Level Auth Error
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrOAuthEmailRequired = errors.New("a public email address is required")
	ErrCannotEditOAuth    = errors.New("cannot perform this action with an OAuth account")
	ErrEmailInUse         = errors.New("email already in use")
	ErrInvalidPassword    = errors.New("current password is incorrect")
)

// Service Level User Error
var (
	ErrEmailTaken   = errors.New("email already registered")
	ErrUserNotFound = errors.New("user not found")
)

// Service Level Organization Error
var (
	ErrOrganizationNotFound      = errors.New("organization not found")
	ErrOrgSlugTaken              = errors.New("organization slug already taken")
	ErrNotOrgOwner               = errors.New("only organization owner can perform this action")
	ErrInvalidOrganizationUpdate = errors.New("organization name or slug must be provided")
)

// Service Level Team Error
var (
	ErrTeamNotFound  = errors.New("team not found")
	ErrTeamSlugTaken = errors.New("team slug already taken in this organization")
)

// Service Level Vault Error
var (
	ErrVaultNotFound      = errors.New("vault not found")
	ErrVaultItemNotFound  = errors.New("vault item not found")
	ErrDuplicateVaultItem = errors.New("vault item with this name and type already exists")
)

// Shared
var (
	ErrEntityHasNoRoles      = errors.New("no role(s) found")
	ErrEntityHasInvalidRoles = errors.New("invalid role(s)")
	ErrInvalidProfileUpdate  = errors.New("username or avatar URL must be provided")
)

// Service Level Environment / Secret Errors
var (
	ErrApplicationNotFound     = errors.New("application not found")
	ErrApplicationSlugTaken    = errors.New("application slug already taken in this organization")
	ErrEnvironmentNotFound     = errors.New("environment not found")
	ErrEnvironmentSlugTaken    = errors.New("environment slug already taken in this application")
	ErrEnvironmentKeyNotFound  = errors.New("environment encryption key not provisioned")
	ErrSecretNotFound          = errors.New("secret not found")
	ErrInvalidSecretKind       = errors.New("invalid secret kind")
)

// Service Level Service Account / API Key Errors
var (
	ErrServiceAccountNotFound = errors.New("service account not found")
	ErrApiKeyNotFound         = errors.New("api key not found")
	ErrApiKeyRevoked          = errors.New("api key revoked")
	ErrApiKeyExpired          = errors.New("api key expired")
	ErrApiKeyDisabled         = errors.New("service account disabled")
	ErrApiKeyScopesInvalid    = errors.New("api key must include at least one known scope")
	ErrApiKeyPermission       = errors.New("api key lacks the required scope")
)
