package domain

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenNotYetValid   = errors.New("token not yet valid")
	ErrInvalidToken       = errors.New("invalid token")
	ErrPasswordTooShort   = errors.New("password too short")
)
