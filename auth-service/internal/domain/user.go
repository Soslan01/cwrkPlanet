package domain

import (
	"strings"
	"time"

	"github.com/cwrk-planet/auth-service/internal/errs"
)

type UserID int64

type User struct {
	ID            UserID
	Email         string
	EmailVerified bool
	PasswordHash  string
	DisplayName   *string
	AvatarURL     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Создает нового пользователя
// Ожидает уже посчитанный хеш пароля
func NewUser(email, passwordHash string, now time.Time, opts ...UserOption) (*User, error) {
	email = normalizeEmail(email)
	if email == "" {
		return nil, errs.ErrInvalidEmail
	}
	if strings.TrimSpace(passwordHash) == "" {
		return nil, errs.ErrEmptyPasswordHash
	}

	user := &User{
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	for _, opt := range opts {
		opt(user)
	}

	return user, nil
}

func (u *User) SetPasswordHash(hash string, now time.Time) error {
	if strings.TrimSpace(hash) == "" {
		return errs.ErrEmptyPasswordHash
	}
	u.PasswordHash = hash
	u.UpdatedAt = now

	return nil
}

func (u *User) SetDisplayName(name *string, now time.Time) {
	u.DisplayName = trimPtr(name)
	u.UpdatedAt = now
}

func (u *User) SetAvatarURL(url *string, now time.Time) {
	u.AvatarURL = trimPtr(url)
	u.UpdatedAt = now
}

func (u *User) VeriyEmail(now time.Time) {
	u.EmailVerified = true
	u.UpdatedAt = now
}

// Options конструктора
type UserOption func(*User)

func WithDisplayName(name string) UserOption {
	return func(u *User) { u.DisplayName = trimPtr(&name) }
}

func WithAvatarURL(url string) UserOption {
	return func(u *User) { u.AvatarURL = trimPtr(&url) }
}

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func trimPtr(p *string) *string {
	if p == nil {
		return nil
	}
	t := strings.TrimSpace(*p)
	if t == "" {
		return nil
	}

	return &t
}
