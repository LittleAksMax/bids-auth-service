package contracts

import "github.com/golang-jwt/jwt/v5"

// Claims represents access token payload.
type Claims struct {
	Role string `json:"role"`
	Name string `json:"name"`
	jwt.RegisteredClaims
}
