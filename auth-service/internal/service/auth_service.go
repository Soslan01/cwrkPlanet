package service

import (
	"context"
	"errors"
	"log/slog"
	"net/netip"
	"time"

	"github.com/cwrk-planet/auth-service/internal/domain"
	"github.com/cwrk-planet/auth-service/internal/errs"
	"github.com/cwrk-planet/auth-service/internal/repository"
	"github.com/cwrk-planet/auth-service/internal/security"
)

type RegisterResult struct {
	User         *domain.User
	AccessToken  string
	RefreshToken string
}

type LoginResult struct {
	User         *domain.User
	AccessToken  string
	RefreshToken string
}

type RefreshResult struct {
	UserID       domain.UserID
	AccessToken  string
	RefreshToken string
}

type AuthService struct {
	users      repository.UserRepository
	sessions   repository.SessionRepository
	jwt        *security.JWTSigner
	refreshTTL time.Duration
	passPolicy security.BcryptConfig
	now        func() time.Time
}

func NewAuthService(
	users repository.UserRepository,
	sessions repository.SessionRepository,
	jwt *security.JWTSigner,
	refreshTTL time.Duration,
	passPolicy security.BcryptConfig,
	now func() time.Time,
) *AuthService {
	if now == nil {
		now = time.Now
	}

	return &AuthService{
		users:      users,
		sessions:   sessions,
		jwt:        jwt,
		refreshTTL: refreshTTL,
		passPolicy: passPolicy,
		now:        now,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string, displayName *string) (*RegisterResult, error) {
	exists, err := s.users.ExistsByEmail(ctx, email)
	if err != nil {
		slog.Error("auth.register.existsByEmail failed by checking:", slog.Any("err", err))
		return nil, err
	}
	if exists {
		slog.Error("auth.register.existsByEmail:", slog.Any("err", repository.ErrAlreadyExists))
		return nil, repository.ErrAlreadyExists
	}

	hash, err := security.HashPassword(password, &s.passPolicy)
	if err != nil {
		slog.Error("auth.register.hashPassword failed", slog.Any("err", err))
		return nil, err
	}

	now := s.now()
	var opts []domain.UserOption
	if displayName != nil {
		opts = append(opts, domain.WithDisplayName(*displayName))
	}

	u, err := domain.NewUser(email, hash, now, opts...)
	if err != nil {
		slog.Error("auth.register.createNewUser failed", slog.Any("err", err))
		return nil, err
	}

	id, err := s.users.Create(ctx, u)
	if err != nil {
		slog.Error("auth.register.createUserProfile failed", slog.Any("err", err))
		return nil, err
	}
	u.ID = id

	// todo: добавить передачу LoginMeta и oldSessionID
	access, refresh, err := s.issueTokens(ctx, u.ID, nil, nil)
	if err != nil {
		slog.Error("auth.register.generateIssueToken failed", slog.Any("err", err))
		return nil, err
	}

	return &RegisterResult{
		User:         u,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

// Login аутентифицирует по email+пароль и выпускает пару токенов
func (s *AuthService) Login(ctx context.Context, email, password string, meta *LoginMeta) (*LoginResult, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			slog.Error("auth.login.getByEmail failed", slog.Any("err", err))
			return nil, errs.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := security.ComparePassword(u.PasswordHash, password); err != nil {
		slog.Error("auth.login.comparePassword failed", slog.Any("err", err))
		return nil, err
	}

	// todo: добавить реальную передачу oldSessionID
	access, refresh, err := s.issueTokens(ctx, u.ID, meta, nil)
	if err != nil {
		slog.Error("auth.login.generateIssueToken failed", slog.Any("err", err))
		return nil, err
	}

	return &LoginResult{
		User:         u,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

// Refresh по refresh-токену выдает новую пару; старую запись удаляет
func (s *AuthService) Refresh(ctx context.Context, refreshToken string, meta *LoginMeta) (*RefreshResult, error) {
	hash := security.SHA256HexOfString(refreshToken)
	sess, err := s.sessions.GetByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			slog.Error("auth.refresh.getByTokenHash failed", slog.Any("err", err))
			return nil, errs.ErrInvalidCredentials
		}
		return nil, err
	}

	now := s.now()
	if sess.IsExpired(now) {
		_ = s.sessions.DeleteByID(ctx, sess.ID)
		slog.Error("auth.refresh.isSessionExpired failed", slog.Any("err", err))
		return nil, errs.ErrSessionExpired
	}

	access, newRefresh, err := s.issueTokens(ctx, sess.UserID, meta, &sess.ID)
	if err != nil {
		slog.Error("auth.refresh.generateIssueToken failed", slog.Any("err", err))
		return nil, err
	}

	return &RefreshResult{
		UserID:       sess.UserID,
		AccessToken:  access,
		RefreshToken: newRefresh,
	}, nil
}

// Me возвращает профиль пользователя
func (s *AuthService) Me(ctx context.Context, userID domain.UserID) (*domain.User, error) {
	var user *domain.User
	var err error

	user, err = s.users.GetByID(ctx, userID)
	if err != nil {
		slog.Error("auth.me.getUserByID failed", slog.Any("err", err))
	}

	return user, nil
}

// Метаданные для записи сессии
type LoginMeta struct {
	UserAgent *string
	IP        *netip.Addr
}

func (s *AuthService) AccessTTL() time.Duration { return s.jwt.TTL() }

// Парсит access JWT и возвращает  userID
func (s *AuthService) UserIDFromAccessToken(token string) (domain.UserID, error) {
	claims, err := s.jwt.ParseAndValidate(token)
	if err != nil {
		return 0, err
	}

	return security.SubjectAsUserID(claims)
}

// issueTokens: создает refresh-сессию и подпистывает токен
// Если oldSessionID != nil - сначала удаляет старую запись
func (s *AuthService) issueTokens(ctx context.Context, userID domain.UserID, meta *LoginMeta, oldSessionID *domain.SessionID) (access string, refresh string, err error) {
	now := s.now()

	// access
	access, err = s.jwt.SignAccessToken(userID, now)
	if err != nil {
		return "", "", err
	}

	// refresh opaque + запись в БД
	refresh, err = security.RandomStringURLSafe(32)
	if err != nil {
		return "", "", err
	}

	hash := security.SHA256HexOfString(refresh)
	expires := now.Add(s.refreshTTL)

	sess, err := domain.NewSession(userID, hash, expires, now)
	if err != nil {
		return "", "", err
	}

	if meta != nil {
		if meta.UserAgent != nil && *meta.UserAgent != "" {
			sess.SetUserAgent(meta.UserAgent, now)
		}
		if meta.IP != nil {
			sess.SetIP(meta.IP, now)
		}
	}

	if oldSessionID != nil {
		_ = s.sessions.DeleteByID(ctx, *oldSessionID)
	}

	if _, err := s.sessions.Create(ctx, sess); err != nil {
		return "", "", err
	}

	return access, refresh, nil
}
