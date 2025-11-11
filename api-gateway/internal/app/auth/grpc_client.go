package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/cwrk-planet/api-gateway/pkg/errs"
	"github.com/cwrk-planet/api-gateway/pkg/httputil"

	authv1 "github.com/cwrk-planet/auth-service/proto/gen/auth/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client interface {
	Login(ctx context.Context, in LoginRequest) (LoginResponse, error)
	Register(ctx context.Context, in RegisterRequest) (RegisterResponse, error)
	Refresh(ctx context.Context, refreshToken string) (RefreshResponse, error)
	Me(ctx context.Context) (MeResponse, error)
	Close() error
}

type client struct {
	conn    *grpc.ClientConn
	auth    authv1.AuthServiceClient
	timeout time.Duration
}

type Options struct {
	Target      string
	Timeout     time.Duration
	UseInsecure bool // зарезервировано под TLS в "Когда-нибудь будет" |:)
}

func New(opts Options) (Client, error) {
	if opts.Target == "" {
		return nil, fmt.Errorf("auth client: empty target")
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 5 * time.Second
	}
	dialOpts := []grpc.DialOption{
		// todo: потом добавить TLS
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(opts.Target, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("auth client: new client failed: %w", err)
	}

	return &client{
		conn:    conn,
		auth:    authv1.NewAuthServiceClient(conn),
		timeout: opts.Timeout,
	}, nil
}

func (c *client) Close() error {
	return c.conn.Close()
}

func (c *client) Login(ctx context.Context, in LoginRequest) (LoginResponse, error) {
	rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	rpcCtx = withOutboundMeta(rpcCtx, "") // x-request-id прокинули в metadata ctx

	req := &authv1.LoginRequest{
		Email:    in.Email,
		Password: in.Password,
	}
	res, err := c.auth.Login(rpcCtx, req)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("%w: %v", errs.ErrUpstream, err)
	}

	return LoginResponse{
		AccessToken:  res.GetAccessToken(),
		RefreshToken: res.GetRefreshToken(),
		ExpiresIn:    res.GetExpiresIn(),
		User: User{
			Id:            res.GetUser().GetId(),
			Email:         res.GetUser().GetEmail(),
			EmailVerified: res.GetUser().GetEmailVerified(),
			DisplayName:   res.GetUser().GetDisplayName(),
			AvatarURL:     res.GetUser().GetAvatarUrl(),
			CreatedAt:     res.GetUser().GetCreatedAt(),
			UpdatedAt:     res.GetUser().GetUpdatedAt(),
		},
	}, nil
}

func (c *client) Register(ctx context.Context, in RegisterRequest) (RegisterResponse, error) {
	rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	rpcCtx = withOutboundMeta(rpcCtx, "") // x-request-id прокинули в metadata ctx

	req := &authv1.RegisterRequest{
		Email:       in.Email,
		Password:    in.Password,
		DisplayName: in.DisplayName,
	}
	res, err := c.auth.Register(rpcCtx, req)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("%w: %v", errs.ErrUpstream, err)
	}

	return RegisterResponse{
		AccessToken:  res.GetAccessToken(),
		RefreshToken: res.GetRefreshToken(),
		ExpiresIn:    res.GetExpiresIn(),
		User: User{
			Id:            res.GetUser().GetId(),
			Email:         res.GetUser().GetEmail(),
			EmailVerified: res.GetUser().GetEmailVerified(),
			DisplayName:   res.GetUser().GetDisplayName(),
			AvatarURL:     res.GetUser().GetAvatarUrl(),
			CreatedAt:     res.GetUser().GetCreatedAt(),
			UpdatedAt:     res.GetUser().GetUpdatedAt(),
		},
	}, nil
}

func (c *client) Refresh(ctx context.Context, refreshToken string) (RefreshResponse, error) {
	rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	rpcCtx = withOutboundMeta(rpcCtx, "") // x-request-id прокинули в metadata ctx

	req := &authv1.RefreshRequest{RefreshToken: refreshToken}
	res, err := c.auth.Refresh(rpcCtx, req)
	if err != nil {
		return RefreshResponse{}, fmt.Errorf("%w: %v", errs.ErrUpstream, err)
	}

	return RefreshResponse{
		AccessToken:  res.GetAccessToken(),
		RefreshToken: res.GetRefreshToken(),
		ExpiresIn:    res.GetExpiresIn(),
	}, nil
}

func (c *client) Me(ctx context.Context) (MeResponse, error) {
	rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	rpcCtx = withOutboundMeta(rpcCtx, "") // x-request-id прокинули в metadata ctx

	res, err := c.auth.Me(rpcCtx, &authv1.MeRequest{})
	if err != nil {
		return MeResponse{}, fmt.Errorf("%w: %v", errs.ErrUpstream, err)
	}

	return MeResponse{
		User: User{
			Id:            res.GetUser().GetId(),
			Email:         res.GetUser().GetEmail(),
			EmailVerified: res.GetUser().GetEmailVerified(),
			DisplayName:   res.GetUser().GetDisplayName(),
			AvatarURL:     res.GetUser().GetAvatarUrl(),
			CreatedAt:     res.GetUser().GetCreatedAt(),
			UpdatedAt:     res.GetUser().GetUpdatedAt(),
		},
	}, nil
}

// Helper: вкладывает "Authorization: Bearer <access_token>" в ctx metadata.
func WithBearer(ctx context.Context, accessToken string) context.Context {
	if accessToken == "" {
		return ctx
	}
	md := metadata.Pairs("authorization", "Bearer "+accessToken)
	return metadata.NewOutgoingContext(ctx, md)
}

func withOutboundMeta(ctx context.Context, accessToken string) context.Context {
	// X-Request-ID из HTTP-контекста gateway
	if rid, ok := httputil.FromContext(ctx); ok && rid != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", rid)
	}
	// Authorization: Bearer <access_token>
	if accessToken != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken)
	}
	return ctx
}
