package auth

import "time"

// JWTClaims represents custom JWT claims. Reserved for future authentication implementation.
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// TokenPair represents a token pair structure. Reserved for future authentication implementation.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}
