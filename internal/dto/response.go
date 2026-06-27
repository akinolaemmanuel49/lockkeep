package dto

type Auth0UserInfo struct {
	Sub      string `json:"sub"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	ID                string `json:"id"`
	Email             string `json:"email"`
	TenantID          string `json:"tenantId"`
	HasMasterPassword bool   `json:"hasMasterPassword"`
	AuthMethod        string `json:"authMethod,omitempty"`
}
