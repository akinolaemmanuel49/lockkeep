package jwt

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	UserID     string     `json:"sub"`
	Email      string     `json:"email"`
	SystemRole string     `json:"system_role"`
	Orgs       []OrgClaim `json:"orgs,omitempty"`
	jwt.RegisteredClaims
}

type OrgClaim struct {
	OrgID     string            `json:"org_id"`
	Role      string            `json:"role"`
	TeamRoles map[string]string `json:"team_roles,omitempty"`
}
