package security

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/cwrk-planet/auth-service/internal/domain"
	"github.com/cwrk-planet/auth-service/internal/errs"

	"github.com/golang-jwt/jwt"
)

// Используется SigningMethodRS256
type JWTSigner struct {
	private   *rsa.PrivateKey
	public    *rsa.PublicKey
	issuer    string
	audience  string
	ttl       time.Duration
	clockSkew time.Duration
}

func NewJWTSigner(private *rsa.PrivateKey, public *rsa.PublicKey, issuer, audience string, ttl, clockSkew time.Duration) *JWTSigner {
	return &JWTSigner{
		private:   private,
		public:    public,
		issuer:    issuer,
		audience:  audience,
		ttl:       ttl,
		clockSkew: clockSkew,
	}
}

func (s *JWTSigner) TTL() time.Duration {
	return s.ttl
}

type AccessClaims struct {
	jwt.StandardClaims // включает поля Issuer, Audience, ExpiresAt, NotBefore, IssuedAt, Subject
	// todo: скорее всего надо буде добавить поля roles, emails и т.п.
}

func (s *JWTSigner) SignAccessToken(userID domain.UserID, now time.Time) (string, error) {
	claims := AccessClaims{
		StandardClaims: jwt.StandardClaims{
			Subject:   fmt.Sprint(int64(userID)),
			Issuer:    s.issuer,
			Audience:  s.audience,
			IssuedAt:  now.Unix(),
			NotBefore: now.Add(-s.clockSkew).Unix(),
			ExpiresAt: now.Add(s.ttl).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	return token.SignedString(s.private)
}

// SignAccessToken выпускает JWT с sub=userID и exp=now+ttl
func (s *JWTSigner) ParseAndValidate(tokenStr string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok || t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, errs.ErrInvalidToken
		}
		return s.public, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errs.ErrInvalidToken
	}

	now := time.Now()

	// issuer
	if !claims.VerifyIssuer(s.issuer, true) {
		return nil, errs.ErrInvalidIssuer
	}
	// audience
	if !claims.VerifyAudience(s.audience, true) {
		return nil, errs.ErrInvalidAudience
	}

	// временные клеймы с допуском clockSkew
	nbf := time.Unix(claims.NotBefore, 0).Add(-s.clockSkew)
	exp := time.Unix(claims.ExpiresAt, 0).Add(s.clockSkew) // даём люфт на «часы»
	// exp / nbf / iat
	if now.Before(nbf) || now.After(exp) {
		return nil, errs.ErrTokenExpired
	}

	return claims, nil
}

// SubjectAsUserID парсит sub в domain.UserID.
func SubjectAsUserID(claims *AccessClaims) (domain.UserID, error) {
	if claims == nil || claims.Subject == "" {
		return 0, errs.ErrInvalidSubject
	}
	var id int64
	_, err := fmt.Sscan(claims.Subject, &id)
	if err != nil {
		return 0, errs.ErrInvalidSubject
	}

	return domain.UserID(id), nil
}

func LoadRSAPrivateKeyFromPEM(path string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("no PEM block in %s", path)
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pk, ok := k.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not RSA private key")
	}

	return pk, nil
}

func LoadRSAPublicKeyFromPEM(path string) (*rsa.PublicKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	pub, err := jwt.ParseRSAPublicKeyFromPEM(b)
	if err != nil {
		return nil, err
	}

	return pub, nil
}
