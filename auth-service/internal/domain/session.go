package domain

import (
	"net/netip"
	"time"
)

type SessionID int64

type Session struct {
	ID               SessionID
	UserID           UserID
	TokenHash        string
	RefreshTokenHash string
	ExpiresAt        time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	UserAgent        *string
	IP               *netip.Addr
}

type SessionOption func(*Session)

func WithUserAgent(ua string) SessionOption {
	return func(s *Session) {
		s.UserAgent = &ua
	}
}

func WithIP(ip netip.Addr) SessionOption {
	return func(s *Session) {
		s.IP = &ip
	}
}

func NewSession(userID UserID, tokenHash, refreshTokenHash string, expiresAt, now time.Time, opts ...SessionOption) *Session {
	s := &Session{
		UserID:           userID,
		TokenHash:        tokenHash,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        expiresAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Session) IsExpired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}
