package errs

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrEmptyPasswordHash  = errors.New("empty password hash")
	ErrEmptyTokenHash     = errors.New("empty token hash")
	ErrPastExpiry         = errors.New("expires_at is in the past")
	ErrPasswordTooShort   = errors.New("password too short")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidIssuer      = errors.New("invalid issuer")
	ErrInvalidAudience    = errors.New("invalid audience")
	ErrTokenExpired       = errors.New("token expired or not valid yet")
	ErrInvalidSubject     = errors.New("invalid subject")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionExpired     = errors.New("session expired")
)
