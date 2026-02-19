package auth

import (
	"context"
	"fmt"
	"time"

	authclient "github.com/cwrk-planet/api-gateway/internal/client/auth"
	"github.com/cwrk-planet/api-gateway/internal/domain"
	authpb "github.com/cwrk-planet/auth-service/proto/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	client *authclient.Client
	logger *zap.Logger
}

func NewService(client *authclient.Client, logger *zap.Logger) *Service {
	return &Service{
		client: client,
		logger: logger,
	}
}

// Register handles user registration
func (s *Service) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.RegisterResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.client.GetTimeout())
	defer cancel()

	grpcReq := &authpb.RegisterRequest{
		Email:       req.Email,
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	}
	if req.AvatarURL != nil {
		grpcReq.AvatarUrl = req.AvatarURL
	}

	resp, err := s.client.GetClient().Register(ctx, grpcReq)
	if err != nil {
		return nil, s.mapError(err)
	}

	return &domain.RegisterResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		User:         protoUserToDomain(resp.User),
	}, nil
}

// Login handles user login
func (s *Service) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.client.GetTimeout())
	defer cancel()

	grpcReq := &authpb.LoginRequest{
		EmailOrUsername: req.Email,
		Password:        req.Password,
	}

	resp, err := s.client.GetClient().Login(ctx, grpcReq)
	if err != nil {
		return nil, s.mapError(err)
	}

	return &domain.LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		User:         protoUserToDomain(resp.User),
	}, nil
}

// RefreshToken handles token refresh
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*domain.RefreshResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.client.GetTimeout())
	defer cancel()

	grpcReq := &authpb.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	resp, err := s.client.GetClient().RefreshToken(ctx, grpcReq)
	if err != nil {
		return nil, s.mapError(err)
	}

	// Calculate expires_in in seconds
	expiresIn := resp.ExpiresAt - time.Now().Unix()
	if expiresIn < 0 {
		expiresIn = 0
	}

	return &domain.RefreshResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// Logout handles user logout
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	ctx, cancel := context.WithTimeout(ctx, s.client.GetTimeout())
	defer cancel()

	grpcReq := &authpb.LogoutRequest{
		RefreshToken: refreshToken,
	}

	_, err := s.client.GetClient().Logout(ctx, grpcReq)
	if err != nil {
		return s.mapError(err)
	}

	return nil
}

// ValidateToken validates an access token
func (s *Service) ValidateToken(ctx context.Context, accessToken string) (bool, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, s.client.GetTimeout())
	defer cancel()

	grpcReq := &authpb.ValidateTokenRequest{
		AccessToken: accessToken,
	}

	resp, err := s.client.GetClient().ValidateToken(ctx, grpcReq)
	if err != nil {
		return false, 0, s.mapError(err)
	}

	return resp.Valid, resp.UserId, nil
}

// GetUser retrieves user information by ID
func (s *Service) GetUser(ctx context.Context, userID int64) (*domain.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.client.GetTimeout())
	defer cancel()

	grpcReq := &authpb.GetUserRequest{
		UserId: userID,
	}

	resp, err := s.client.GetClient().GetUser(ctx, grpcReq)
	if err != nil {
		return nil, s.mapError(err)
	}

	return protoUserToDomain(resp.User), nil
}

// mapError maps gRPC errors to domain errors
func (s *Service) mapError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		s.logger.Error("Unknown error type", zap.Error(err))
		return domain.ErrInternalServer
	}

	switch st.Code() {
	case codes.NotFound:
		return domain.ErrNotFound
	case codes.Unauthenticated:
		return domain.ErrUnauthorized
	case codes.PermissionDenied:
		return domain.ErrForbidden
	case codes.InvalidArgument:
		return fmt.Errorf("%w: %s", domain.ErrInvalidRequest, st.Message())
	case codes.Internal:
		return domain.ErrInternalServer
	case codes.Unavailable:
		return domain.ErrServiceUnavailable
	default:
		s.logger.Error("Unmapped gRPC error", zap.String("code", st.Code().String()), zap.String("message", st.Message()))
		return domain.ErrInternalServer
	}
}

// protoUserToDomain converts proto User to domain UserResponse
func protoUserToDomain(user *authpb.User) *domain.UserResponse {
	var avatarURL *string
	if user.AvatarUrl != nil && *user.AvatarUrl != "" {
		avatarURL = user.AvatarUrl
	}

	return &domain.UserResponse{
		ID:          user.Id,
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		AvatarURL:   avatarURL,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
