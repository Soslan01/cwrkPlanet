package handler

import (
	"context"
	"net/netip"

	"github.com/cwrk-planet/auth-service/internal/domain"
	"github.com/cwrk-planet/auth-service/internal/service"
	"github.com/cwrk-planet/auth-service/proto/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	auth.UnimplementedAuthServiceServer
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	var avatarURL *string
	if req.AvatarUrl != nil && *req.AvatarUrl != "" {
		avatarURL = req.AvatarUrl
	}

	tokenPair, user, err := h.authService.Register(ctx, service.RegisterRequest{
		Email:       req.Email,
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.DisplayName,
		AvatarURL:   avatarURL,
	})

	if err != nil {
		return nil, h.mapError(err)
	}

	return &auth.RegisterResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Unix(),
		User:         h.domainUserToProto(user),
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	var userAgent *string
	if req.UserAgent != nil && *req.UserAgent != "" {
		userAgent = req.UserAgent
	}

	var ip *netip.Addr
	if req.IpAddress != nil && *req.IpAddress != "" {
		parsedIP, err := netip.ParseAddr(*req.IpAddress)
		if err == nil {
			ip = &parsedIP
		}
	}

	tokenPair, user, err := h.authService.Login(ctx, service.LoginRequest{
		EmailOrUsername: req.EmailOrUsername,
		Password:        req.Password,
		UserAgent:       userAgent,
		IP:              ip,
	})

	if err != nil {
		return nil, h.mapError(err)
	}

	return &auth.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Unix(),
		User:         h.domainUserToProto(user),
	}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.RefreshTokenResponse, error) {
	tokenPair, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, h.mapError(err)
	}

	return &auth.RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Unix(),
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *auth.LogoutRequest) (*auth.LogoutResponse, error) {
	if err := h.authService.Logout(ctx, req.RefreshToken); err != nil {
		return nil, h.mapError(err)
	}

	return &auth.LogoutResponse{}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	valid, userID, err := h.authService.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return &auth.ValidateTokenResponse{
			Valid:  false,
			UserId: 0,
		}, nil
	}

	return &auth.ValidateTokenResponse{
		Valid:  valid,
		UserId: int64(userID),
	}, nil
}

func (h *AuthHandler) GetUser(ctx context.Context, req *auth.GetUserRequest) (*auth.GetUserResponse, error) {
	user, err := h.authService.GetUser(ctx, domain.UserID(req.UserId))
	if err != nil {
		return nil, h.mapError(err)
	}

	return &auth.GetUserResponse{
		User: h.domainUserToProto(user),
	}, nil
}

func (h *AuthHandler) domainUserToProto(user *domain.User) *auth.User {
	var avatarURL *string
	if user.AvatarURL != nil {
		avatarURL = user.AvatarURL
	}

	return &auth.User{
		Id:          int64(user.ID),
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		AvatarUrl:   avatarURL,
		CreatedAt:   user.CreatedAt.Unix(),
		UpdatedAt:   user.UpdatedAt.Unix(),
	}
}

func (h *AuthHandler) mapError(err error) error {
	switch err {
	case domain.ErrUserNotFound:
		return status.Error(codes.NotFound, "user not found")
	case domain.ErrUserAlreadyExists:
		return status.Error(codes.AlreadyExists, "user already exists")
	case domain.ErrInvalidCredentials:
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case domain.ErrSessionNotFound:
		return status.Error(codes.NotFound, "session not found")
	case domain.ErrSessionExpired:
		return status.Error(codes.Unauthenticated, "session expired")
	case domain.ErrTokenExpired:
		return status.Error(codes.Unauthenticated, "token expired")
	case domain.ErrInvalidToken:
		return status.Error(codes.Unauthenticated, "invalid token")
	case domain.ErrPasswordTooShort:
		return status.Error(codes.InvalidArgument, "password too short")
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
