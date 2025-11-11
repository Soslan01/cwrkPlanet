package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/cwrk-planet/auth-service/internal/domain"
	"github.com/cwrk-planet/auth-service/internal/repository"

	"github.com/jackc/pgx/v5"
	pgconn "github.com/jackc/pgx/v5/pgconn"
)

/*
абстрактный слой над *pgxpool.Pool / pgx.Tx
чтобы запросы можно было делать атомарно а не по одному
*/
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func (r *UserRepo) getOne(ctx context.Context, sql string, arg any) (*domain.User, error) {
	var (
		id            int64
		email         string
		emailVerified bool
		passwordHash  string
		displayName   *string
		avatarURL     *string
		createdAt     time.Time
		updatedAt     time.Time
	)

	err := r.q.QueryRow(ctx, sql, arg).Scan(
		&id,
		&email,
		&emailVerified,
		&passwordHash,
		&displayName,
		&avatarURL,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, mapPgError(err)
	}

	return &domain.User{
		ID:            domain.UserID(id),
		Email:         email,
		EmailVerified: emailVerified,
		PasswordHash:  passwordHash,
		DisplayName:   displayName,
		AvatarURL:     avatarURL,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}

func mapPgError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 23505 - unique violation
		if pgErr.Code == "23505" {
			return repository.ErrAlreadyExists
		}
		// todo:  добавить маппинг для других кодов
	}

	return err
}

func toNullStringPtr(p *string) *string {
	if p == nil {
		return nil
	}
	s := strings.TrimSpace(*p)
	if s == "" {
		return nil
	}

	return &s
}

func itoa(i int) string {
	const digits = "0123456789"
	if i == 0 {
		return "0"
	}
	var b [20]byte
	n := len(b)
	for i > 0 {
		n--
		b[n] = digits[i%10]
		i /= 10
	}

	return string(b[n:])
}
