package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/netip"

	"github.com/cwrk-planet/auth-service/internal/database"
	"github.com/cwrk-planet/auth-service/internal/domain"
	"github.com/jackc/pgx/v5"
)

type SessionRepository struct {
	db *database.DB
}

func NewSessionRepository(db *database.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	err := r.db.Pool.QueryRow(
		ctx,
		QueryCreateSession,
		session.UserID,
		hashToken(session.TokenHash),
		hashToken(session.RefreshTokenHash),
		session.ExpiresAt,
		session.CreatedAt,
		session.UpdatedAt,
		session.UserAgent,
		session.IP,
	).Scan(&session.ID)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (r *SessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	hashed := hashToken(tokenHash)

	var session domain.Session
	var ipStr *string
	err := r.db.Pool.QueryRow(
		ctx,
		QueryGetByTokenHash,
		hashed,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.RefreshTokenHash,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.UserAgent,
		&ipStr,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session by token hash: %w", err)
	}

	if ipStr != nil {
		ip, err := netip.ParseAddr(*ipStr)
		if err == nil {
			session.IP = &ip
		}
	}

	return &session, nil
}

func (r *SessionRepository) GetByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*domain.Session, error) {
	hashed := hashToken(refreshTokenHash)

	var session domain.Session
	var ipStr *string
	err := r.db.Pool.QueryRow(
		ctx,
		QueryGetSessionByRefreshToken,
		hashed,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.RefreshTokenHash,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.UserAgent,
		&ipStr,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session by refresh token hash: %w", err)
	}

	if ipStr != nil {
		ip, err := netip.ParseAddr(*ipStr)
		if err == nil {
			session.IP = &ip
		}
	}

	return &session, nil
}

func (r *SessionRepository) Delete(ctx context.Context, sessionID domain.SessionID) error {
	result, err := r.db.Pool.Exec(
		ctx,
		QueryDeleteSession,
		sessionID,
	)

	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.db.Pool.Exec(
		ctx,
		QueryDeleteExpiredSession,
	)

	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}
