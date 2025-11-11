package security

import (
	"github.com/cwrk-planet/auth-service/internal/errs"
	"golang.org/x/crypto/bcrypt"
)

// todo: такое надо перенести в config
type BcryptConfig struct {
	Cost      int // по умолчанию 10/12
	MinLength int // по умолчанию 6/8
}

func HashPassword(plain string, cfg *BcryptConfig) (string, error) {
	minLen := 6
	cost := bcrypt.DefaultCost

	if cfg != nil {
		if cfg.MinLength > 0 {
			minLen = cfg.MinLength
		}
		if cfg.Cost > 0 {
			cost = cfg.Cost
		}
	}

	if len(plain) < minLen {
		return "", errs.ErrPasswordTooShort
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func ComparePassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
