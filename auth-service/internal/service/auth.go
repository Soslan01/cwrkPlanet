package service

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"github.com/cwrk-planet/auth-service/internal/domain"
	"github.com/cwrk-planet/auth-service/internal/repository"
)

type AuthService struct {
	userRepo        *repository.UserRepository
	sessionRepo     *repository.SessionRepository
	passwordService *PasswordService
	tokenService    *TokenService
	refreshTTL      time.Duration
}

func NewAuthService(
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	passwordService *PasswordService,
	tokenService *TokenService,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		passwordService: passwordService,
		tokenService:    tokenService,
		refreshTTL:      refreshTTL,
	}
}

type RegisterRequest struct {
	Email       string
	Username    string
	Password    string
	DisplayName string
	AvatarURL   *string
}

type LoginRequest struct {
	EmailOrUsername string
	Password        string
	UserAgent       *string
	IP              *netip.Addr
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*domain.TokenPair, *domain.User, error) {
	// Validate input
	if req.Email == "" || req.Username == "" || req.Password == "" || req.DisplayName == "" {
		return nil, nil, fmt.Errorf("all fields are required")
	}

	// Hash password
	passwordHash, err := s.passwordService.Hash(req.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	now := time.Now()
	user := &domain.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: passwordHash,
		DisplayName:  req.DisplayName,
		AvatarURL:    req.AvatarURL,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	// Generate tokens
	accessToken, err := s.tokenService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create session
	expiresAt := time.Now().Add(s.refreshTTL)
	session := domain.NewSession(
		user.ID,
		accessToken,
		refreshToken,
		expiresAt,
		now,
	)

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, user, nil
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*domain.TokenPair, *domain.User, error) {
	// Find user by email or username
	var user *domain.User
	var err error

	user, err = s.userRepo.GetByEmail(ctx, req.EmailOrUsername)
	if err != nil && err != domain.ErrUserNotFound {
		return nil, nil, err
	}

	if user == nil {
		user, err = s.userRepo.GetByUsername(ctx, req.EmailOrUsername)
		if err != nil {
			return nil, nil, domain.ErrInvalidCredentials
		}
	}

	// Verify password
	if err := s.passwordService.Verify(user.PasswordHash, req.Password); err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.tokenService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create session
	now := time.Now()
	expiresAt := now.Add(s.refreshTTL)

	var opts []domain.SessionOption
	if req.UserAgent != nil {
		opts = append(opts, domain.WithUserAgent(*req.UserAgent))
	}
	if req.IP != nil {
		opts = append(opts, domain.WithIP(*req.IP))
	}

	session := domain.NewSession(
		user.ID,
		accessToken,
		refreshToken,
		expiresAt,
		now,
		opts...,
	)

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, user, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	// Find session by refresh token
	session, err := s.sessionRepo.GetByRefreshTokenHash(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	// Check if session is expired
	if session.IsExpired(time.Now()) {
		return nil, domain.ErrSessionExpired
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	// Generate new tokens
	accessToken, err := s.tokenService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Delete old session
	if err := s.sessionRepo.Delete(ctx, session.ID); err != nil {
		return nil, fmt.Errorf("failed to delete old session: %w", err)
	}

	// Create new session
	now := time.Now()
	expiresAt := now.Add(s.refreshTTL)

	var opts []domain.SessionOption
	if session.UserAgent != nil {
		opts = append(opts, domain.WithUserAgent(*session.UserAgent))
	}
	if session.IP != nil {
		opts = append(opts, domain.WithIP(*session.IP))
	}

	newSession := domain.NewSession(
		user.ID,
		accessToken,
		newRefreshToken,
		expiresAt,
		now,
		opts...,
	)

	if err := s.sessionRepo.Create(ctx, newSession); err != nil {
		return nil, fmt.Errorf("failed to create new session: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	session, err := s.sessionRepo.GetByRefreshTokenHash(ctx, refreshToken)
	if err != nil {
		return err
	}

	return s.sessionRepo.Delete(ctx, session.ID)
}

func (s *AuthService) ValidateToken(ctx context.Context, accessToken string) (bool, domain.UserID, error) {
	claims, err := s.tokenService.ValidateAccessToken(accessToken)
	if err != nil {
		return false, 0, err
	}

	return true, domain.UserID(claims.UserID), nil
}

func (s *AuthService) GetUser(ctx context.Context, userID domain.UserID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
