package domain

import "errors"

var (
	ErrInvalidRequest      = errors.New("invalid request")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrNotFound           = errors.New("not found")
	ErrInternalServer     = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service unavailable")
)
