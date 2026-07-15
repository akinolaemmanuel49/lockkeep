package auth_errors

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrOAuthEmailRequired = errors.New("a public email address is required")
)
