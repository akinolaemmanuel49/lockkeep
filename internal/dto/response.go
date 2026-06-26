package dto

type Auth0UserInfo struct {
	Sub      string `json:"sub"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}
