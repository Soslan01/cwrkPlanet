package handler

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strconv"
	"strings"

	"github.com/cwrk-planet/auth-service/internal/domain"
	"github.com/cwrk-planet/auth-service/internal/errs"
	"github.com/cwrk-planet/auth-service/internal/repository"
	"github.com/cwrk-planet/auth-service/internal/service"

	authv1 "github.com/cwrk-planet/auth-service/proto/gen/auth/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register: email+password (+optional display_name) - resp: access, refresh, user
func (h *AuthHandler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	email := strings.TrimSpace(req.GetEmail())
	pass := req.GetPassword()
	var displayName *string
	if dn := strings.TrimSpace(req.GetDisplayName()); dn != "" {
		displayName = &dn
	}

	res, err := h.svc.Register(ctx, email, pass, displayName)
	if err != nil {
		return nil, mapError(err)
	}

	expires := int64(h.svc.AccessTTL().Seconds())

	return &authv1.RegisterResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiresIn:    expires,
		User:         toUserPB(res.User),
	}, nil
}

// Login: email+password - resp: access, refresh, user
func (h *AuthHandler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	meta := extractLoginMeta(ctx)

	res, err := h.svc.Login(ctx, strings.TrimSpace(req.GetEmail()), req.GetPassword(), meta)
	if err != nil {
		return nil, mapError(err)
	}

	expires := int64(h.svc.AccessTTL().Seconds())

	return &authv1.LoginResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiresIn:    expires,
		User:         toUserPB(res.User),
	}, nil
}

// Refresh: refresh_token - resp: new access, new refresh
func (h *AuthHandler) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	meta := extractLoginMeta(ctx)

	res, err := h.svc.Refresh(ctx, strings.TrimSpace(req.GetRefreshToken()), meta)
	if err != nil {
		return nil, mapError(err)
	}

	expires := int64(h.svc.AccessTTL().Seconds())

	return &authv1.RefreshResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiresIn:    expires,
	}, nil
}

// Me: получить профиль по user_id, который кладёт API-Gateway после валидации access-JWT.
func (h *AuthHandler) Me(ctx context.Context, req *authv1.MeRequest) (*authv1.MeResponse, error) {
	// Попытка №1: x-user-id из метаданных которые кладет API-Gateway
	if uid, ok := userIDFromMD(ctx); ok {
		u, err := h.svc.Me(ctx, uid)
		if err != nil {
			return nil, mapError(err)
		}
		return &authv1.MeResponse{User: toUserPB(u)}, nil
	}

	// Попытка №2: Authorization: Bearer <accessToken>
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		authz := firstNonEmpty(md, "authorization")
		if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
			token := strings.TrimSpace(authz[len("bearer "):])
			if token != "" {
				if uid, err := h.svc.UserIDFromAccessToken(token); err == nil {
					u, err := h.svc.Me(ctx, uid)
					if err != nil {
						return nil, mapError(err)
					}
					return &authv1.MeResponse{User: toUserPB(u)}, nil
				}
				return nil, status.Error(codes.Unauthenticated, "invalid access token")
			}
		}
	}

	return nil, status.Error(codes.Unauthenticated, "missing user id (x-user-id)")
}

// ---- helpers ----

func toUserPB(u *domain.User) *authv1.User {
	if u == nil {
		return nil
	}
	var dn, av string
	if u.DisplayName != nil {
		dn = *u.DisplayName
	}
	if u.AvatarURL != nil {
		av = *u.AvatarURL
	}

	return &authv1.User{
		Id:            int64(u.ID),
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		DisplayName:   dn,
		AvatarUrl:     av,
		CreatedAt:     u.CreatedAt.Unix(),
		UpdatedAt:     u.UpdatedAt.Unix(),
	}
}

func mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, repository.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, repository.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, errs.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, errs.ErrSessionExpired):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

// userIDFromMD пытается извлечь user_id из gRPC метаданных.
func userIDFromMD(ctx context.Context) (domain.UserID, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, false
	}

	for _, key := range []string{"x-user-id", "x-userid", "x-user_id"} {
		if vals := md.Get(key); len(vals) > 0 {
			if id, err := strconv.ParseInt(strings.TrimSpace(vals[0]), 10, 64); err == nil && id > 0 {
				return domain.UserID(id), true
			}
		}
	}

	return 0, false
}

// extractLoginMeta собирает метаданные для записи сессии: UserAgent и IP.
func extractLoginMeta(ctx context.Context) *service.LoginMeta {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	var ua *string
	if v := firstNonEmpty(md, "user-agent"); v != "" {
		ua = &v
	}

	// Источники IP в приоритете: x-forwarded-for, x-real-ip, peer addr.
	var ip *netip.Addr
	if xff := firstNonEmpty(md, "x-forwarded-for"); xff != "" {
		if p := parseXFF(xff); p.IsValid() {
			ip = &p
		}
	}
	if ip == nil {
		if xr := firstNonEmpty(md, "x-real-ip"); xr != "" {
			if p, err := netip.ParseAddr(xr); err == nil {
				ip = &p
			}
		}
	}
	if ip == nil {
		// Пытаемся достать peer IP (если gRPC прокинул в контекст)
		if pr, ok := peerIP(ctx); ok {
			ip = &pr
		}
	}

	if ua == nil && ip == nil {
		return nil
	}

	return &service.LoginMeta{UserAgent: ua, IP: ip}
}

func firstNonEmpty(md metadata.MD, key string) string {
	if vals := md.Get(key); len(vals) > 0 {
		if s := strings.TrimSpace(vals[0]); s != "" {
			return s
		}
	}

	return ""
}

// parseXFF берёт первый IP из x-forwarded-for
func parseXFF(xff string) netip.Addr {
	parts := strings.Split(xff, ",")
	if len(parts) == 0 {
		return netip.Addr{}
	}
	s := strings.TrimSpace(parts[0])
	addr, _ := netip.ParseAddr(s)

	return addr
}

// peerIP пытается извлечь IP удалённой стороны из контекста (если доступно).
func peerIP(ctx context.Context) (netip.Addr, bool) {

	p, ok := peer.FromContext(ctx)
	if !ok || p == nil || p.Addr == nil {
		return netip.Addr{}, false
	}

	host, _, err := net.SplitHostPort(p.Addr.String())
	if err != nil {
		return netip.Addr{}, false
	}

	ip, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, false
	}

	return ip, true
}
