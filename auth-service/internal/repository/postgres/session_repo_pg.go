package postgres

import (
	"context"
	"errors"
	"net/netip"
	"strings"
	"time"

	"github.com/cwrk-planet/auth-service/internal/domain"
	"github.com/cwrk-planet/auth-service/internal/repository"
	"github.com/cwrk-planet/auth-service/internal/repository/queries"

	"github.com/jackc/pgx/v5"
)

type SessionRepo struct {
	q querier
}

func NewSessionRepoFromPool(q querier) *SessionRepo {
	return &SessionRepo{q: q}
}

func NewSessionRepoFromTx(tx pgx.Tx) *SessionRepo {
	return &SessionRepo{q: tx}
}

// Create — создает новую refresh-сессию и возвращает её ID.
func (r *SessionRepo) Create(ctx context.Context, s *domain.Session) (domain.SessionID, error) {
	var id int64
	var ip any
	if s.IP != nil {
		ip = s.IP.String()
	} else {
		ip = nil
	}
	err := r.q.QueryRow(
		ctx,
		queries.QueryCreateSession,
		s.UserID,
		strings.TrimSpace(s.TokenHash),
		s.ExpiresAt,
		s.CreatedAt,
		s.UpdatedAt,
		toNullStringPtr(s.UserAgent),
		ip,
	).Scan(&id)
	if err != nil {
		return 0, mapPgError(err)
	}
	return domain.SessionID(id), nil
}

// GetByTokenHash — ищет сессию по точному хешу refresh-токена.
func (r *SessionRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	var (
		id        int64
		userID    int64
		hash      string
		expiresAt time.Time
		createdAt time.Time
		updatedAt time.Time
		userAgent *string
		ipText    *string
	)

	err := r.q.QueryRow(ctx, queries.QueryGetSessionByTokenHash, strings.TrimSpace(tokenHash)).Scan(
		&id,
		&userID,
		&hash,
		&expiresAt,
		&createdAt,
		&updatedAt,
		&userAgent,
		&ipText,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, mapPgError(err)
	}

	var ipPtr *netip.Addr
	if ipText != nil && *ipText != "" {
		if addr, perr := netip.ParseAddr(*ipText); perr == nil {
			ipPtr = &addr
		}
	}

	return &domain.Session{
		ID:        domain.SessionID(id),
		UserID:    domain.UserID(userID),
		TokenHash: hash,
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		UserAgent: userAgent,
		IP:        ipPtr,
	}, nil
}

func (r *SessionRepo) DeleteByID(ctx context.Context, id domain.SessionID) error {
	tag, err := r.q.Exec(ctx, queries.QueryDeleteSessionByID, id)
	if err != nil {
		return mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *SessionRepo) DeleteByUser(ctx context.Context, userID domain.UserID) (int64, error) {
	const sql = `DELETE FROM auth_sessions WHERE user_id = $1;`
	tag, err := r.q.Exec(ctx, sql, userID)
	if err != nil {
		return 0, mapPgError(err)
	}
	return int64(tag.RowsAffected()), nil
}

func (r *SessionRepo) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	tag, err := r.q.Exec(ctx, queries.QueryDeleteSessionsExpiredByTime, now)
	if err != nil {
		return 0, mapPgError(err)
	}
	return int64(tag.RowsAffected()), nil
}
