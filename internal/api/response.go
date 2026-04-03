package api

type HealthServiceStatusResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type HealthResponseData map[string]HealthServiceStatusResponse

type AuthUserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
	Role      string `json:"role"`
}

type AuthTokensResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

type AuthResponseData struct {
	User   AuthUserResponse   `json:"user"`
	Tokens AuthTokensResponse `json:"tokens"`
}
