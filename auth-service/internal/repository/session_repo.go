package repository

import (
	"context"
	"time"

	"github.com/cwrk-planet/auth-service/internal/domain"
)

type SessionRepository interface {
	// Создает новую refresh сессию
	Create(ctx context.Context, s *domain.Session) (domain.SessionID, error)
	// Ищет сессию по хешу refresh - токена
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)
	// Удаляет запись сессии
	DeleteByID(ctx context.Context, id domain.SessionID) error
	// Удаляет все сессии пользователя
	DeleteByUser(ctx context.Context, userID domain.UserID) (int64, error)
	// Очистка просроченных сессий на момент now
	DeleteExpired(ctx context.Context, now time.Time) (int64, error)
}
