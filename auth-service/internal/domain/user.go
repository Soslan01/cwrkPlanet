package domain

import "time"

type UserID int64

type User struct {
	ID           UserID
	Email        string
	Username     string
	PasswordHash string
	DisplayName  string
	AvatarURL    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) Validate() error {
	if u.Email == "" {
		return ErrInvalidEmail
	}
	if u.Username == "" {
		return ErrInvalidUsername
	}
	if u.PasswordHash == "" {
		return ErrInvalidPassword
	}
	return nil
}
