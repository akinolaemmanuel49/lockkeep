package services

import "errors"

// Service Level Auth Error
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrOAuthEmailRequired = errors.New("a public email address is required")
)

// Service Level User Error
var (
	ErrEmailTaken   = errors.New("email already registered")
	ErrUserNotFound = errors.New("user not found")
)

// Service Level Organization Error
var (
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrOrgSlugTaken         = errors.New("organization slug already taken")
	ErrNotOrgOwner          = errors.New("only organization owner can perform this action")
)

// Service Level Team Error
var (
	ErrTeamNotFound  = errors.New("team not found")
	ErrTeamSlugTaken = errors.New("team slug already taken in this organization")
)

// Shared
var (
	ErrEntityHasNoRoles      = errors.New("no role(s) found")
	ErrEntityHasInvalidRoles = errors.New("invalid role(s)")
	ErrInvalidProfileUpdate  = errors.New("username or avatar URL must be provided")
)
