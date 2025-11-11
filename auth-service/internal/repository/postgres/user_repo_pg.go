package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/cwrk-planet/auth-service/internal/domain"
	"github.com/cwrk-planet/auth-service/internal/repository"
	"github.com/cwrk-planet/auth-service/internal/repository/queries"

	"github.com/jackc/pgx/v5"
)

type UserRepo struct {
	q querier
}

// NewUserRepoFromPool - конструктор от пула (*pgxpool.Pool)
func NewUserRepoFromPool(q querier) *UserRepo {
	return &UserRepo{q: q}
}

// NewUserRepoFromTx - конструктор от транзакции (pgx.Tx), удобно для составных операций
func NewUserRepoFromTx(tx pgx.Tx) *UserRepo {
	return &UserRepo{q: tx}
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) (domain.UserID, error) {
	var id int64
	err := r.q.QueryRow(
		ctx,
		queries.QueryCreateUser,
		u.Email,
		u.EmailVerified,
		u.PasswordHash,
		toNullStringPtr(u.DisplayName),
		toNullStringPtr(u.AvatarURL),
		u.CreatedAt,
		u.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, mapPgError(err)
	}

	return domain.UserID(id), nil
}

func (r *UserRepo) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	return r.getOne(ctx, queries.QueryGetUserByID, id)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.getOne(ctx, queries.QueryGetUserByEmail, email)
}

func (r *UserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var one int
	err := r.q.QueryRow(ctx, queries.QueryExistsUserByEmail, strings.TrimSpace(email)).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, mapPgError(err)
	}

	return true, nil
}

func (r *UserRepo) UpdatePasswordHash(ctx context.Context, id domain.UserID, newHash string, now time.Time) error {
	tag, err := r.q.Exec(ctx, queries.QueryUpdatePasswordHash, id, strings.TrimSpace(newHash), now)
	if err != nil {
		return mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *UserRepo) UpdateProfile(ctx context.Context, id domain.UserID, displayName *string, avatarURL *string, now time.Time) error {
	setParts := make([]string, 0, 3)
	args := make([]any, 0, 4)
	i := 1

	if displayName != nil {
		setParts = append(setParts, "display_name = $"+itoa(i))
		args = append(args, toNullStringPtr(displayName))
		i++
	}
	if avatarURL != nil {
		setParts = append(setParts, "avatar_url = $"+itoa(i))
		args = append(args, toNullStringPtr(avatarURL))
		i++
	}

	setParts = append(setParts, "updated_at = $"+itoa(i))
	args = append(args, now)
	i++

	if len(setParts) == 1 {
		// Если ничего не передали, то updated at менять нет смысла
		return nil
	}

	sql := "UPDATE users SET " + strings.Join(setParts, ", ") + " WHERE id = $" + itoa(i) + ";"
	args = append(args, id)

	tag, err := r.q.Exec(ctx, sql, args...)
	if err != nil {
		return mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *UserRepo) MarkEmailVerified(ctx context.Context, id domain.UserID, now time.Time) error {
	tag, err := r.q.Exec(ctx, queries.QueryUpdateEmailVerified, id, now)
	if err != nil {
		return mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}
