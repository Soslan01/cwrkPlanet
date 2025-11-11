package domain

import (
	"net/netip"
	"strings"
	"time"

	"github.com/cwrk-planet/auth-service/internal/errs"
)

type SessionID int64

// Запись о refresh-сессии пользователя
type Session struct {
	ID        SessionID
	UserID    UserID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	UserAgent *string
	IP        *netip.Addr
}

func NewSession(userID UserID, tokenHash string, expiresAt, now time.Time, opts ...SessionOption) (*Session, error) {
	if strings.TrimSpace(tokenHash) == "" {
		return nil, errs.ErrEmptyTokenHash
	}
	if !expiresAt.After(now) {
		return nil, errs.ErrPastExpiry
	}

	s := &Session{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

func (s *Session) Touch(now time.Time) {
	s.UpdatedAt = now
}

func (s *Session) SetUserAgent(ua *string, now time.Time) {
	s.UserAgent = trimPtr(ua)
	s.UpdatedAt = now
}

func (s *Session) SetIP(addr *netip.Addr, now time.Time) {
	s.IP = addr
	s.UpdatedAt = now
}

func (s *Session) IsExpired(now time.Time) bool {
	return !s.ExpiresAt.After(now)
}

// Options конструктора
type SessionOption func(*Session)

func WithUserAgent(ua string) SessionOption {
	return func(s *Session) { s.UserAgent = trimPtr(&ua) }
}

func WithIP(addr netip.Addr) SessionOption {
	return func(s *Session) { s.IP = &addr }
}
