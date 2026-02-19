package service

import (
	"fmt"

	"github.com/cwrk-planet/auth-service/internal/config"
	"github.com/cwrk-planet/auth-service/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type PasswordService struct {
	cfg config.Password
}

func NewPasswordService(cfg config.Password) *PasswordService {
	return &PasswordService{cfg: cfg}
}

func (s *PasswordService) Hash(password string) (string, error) {
	if err := s.Validate(password); err != nil {
		return "", err
	}

	cost := s.cfg.BcryptCost
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

func (s *PasswordService) Verify(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (s *PasswordService) Validate(password string) error {
	if len(password) < s.cfg.MinLength {
		return domain.ErrPasswordTooShort
	}
	return nil
}
