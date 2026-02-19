package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cwrk-planet/auth-service/internal/database"
	"github.com/cwrk-planet/auth-service/internal/domain"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	err := r.db.Pool.QueryRow(
		ctx,
		QueryCreateUser,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.DisplayName,
		user.AvatarURL,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		if isUniqueViolation(err) {
			if contains(err.Error(), "email") {
				return domain.ErrUserAlreadyExists
			}
			if contains(err.Error(), "username") {
				return domain.ErrUserAlreadyExists
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	var user domain.User
	err := r.db.Pool.QueryRow(
		ctx,
		QueryGetUserByID,
		id,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Pool.QueryRow(
		ctx,
		QueryGetUserByEmail,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.Pool.QueryRow(
		ctx,
		QueryGetUserByUsername,
		username,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	result, err := r.db.Pool.Exec(
		ctx,
		QueryUpdateUser,
		user.ID,
		user.Email,
		user.Username,
		user.DisplayName,
		user.AvatarURL,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func isUniqueViolation(err error) bool {
	errStr := err.Error()
	return contains(errStr, "unique constraint") || contains(errStr, "duplicate key")
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
