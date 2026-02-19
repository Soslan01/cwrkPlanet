package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/cwrk-planet/auth-service/internal/config"
	"github.com/cwrk-planet/auth-service/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	cfg        config.JWT
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewTokenService(cfg config.JWT) (*TokenService, error) {
	privateKey, err := loadPrivateKey(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	publicKey, err := loadPublicKey(cfg.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	return &TokenService{
		cfg:        cfg,
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failedf to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		rsaKey, ok := parsedKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("key is not RSA private key")
		}
		return rsaKey, nil
	}

	return key, nil
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not RSA public key")
	}

	return rsaPub, nil
}

func (s *TokenService) GenerateAccessToken(userID domain.UserID, email string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": int64(userID),
		"email":   email,
		"iss":     s.cfg.Issuer,
		"aud":     s.cfg.Audience,
		"iat":     now.Unix(),
		"exp":     now.Add(s.cfg.AccessTTL).Unix(),
		"nbf":     now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *TokenService) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
}

func (s *TokenService) ValidateAccessToken(tokenString string) (*domain.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.ErrInvalidToken
	}

	// Validate standard claims
	now := time.Now().Unix()
	exp, ok := claims["exp"].(float64)
	if !ok || int64(exp) < now {
		return nil, domain.ErrTokenExpired
	}

	iat, ok := claims["iat"].(float64)
	if !ok || int64(iat) > now+int64(s.cfg.ClockSkew.Seconds()) {
		return nil, domain.ErrTokenNotYetValid
	}

	iss, _ := claims["iss"].(string)
	if iss != s.cfg.Issuer {
		return nil, domain.ErrInvalidToken
	}

	aud, _ := claims["aud"].(string)
	if aud != s.cfg.Audience {
		return nil, domain.ErrInvalidToken
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, domain.ErrInvalidToken
	}

	email, _ := claims["email"].(string)

	return &domain.TokenClaims{
		UserID:    int64(userID),
		Email:     email,
		Issuer:    iss,
		Audience:  aud,
		IssuedAt:  int64(iat),
		ExpiresAt: int64(exp),
	}, nil
}
