package domain

import "time"

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type TokenClaims struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	Issuer    string `json:"iss"`
	Audience  string `json:"aud"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func (c *TokenClaims) Valid() error {
	now := time.Now().Unix()
	if c.ExpiresAt < now {
		return ErrTokenExpired
	}
	if c.IssuedAt > now {
		return ErrTokenNotYetValid
	}
	return nil
}
