package repository

import (
	"context"
	"time"

	"github.com/cwrk-planet/auth-service/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) (domain.UserID, error)
	GetByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdatePasswordHash(ctx context.Context, id domain.UserID, newHash string, now time.Time) error
	UpdateProfile(ctx context.Context, id domain.UserID, displayName *string, avatarURL *string, now time.Time) error
	MarkEmailVerified(ctx context.Context, id domain.UserID, now time.Time) error
}
