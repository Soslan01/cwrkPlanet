package auth

import (
	"context"
	"time"

	authpb "github.com/cwrk-planet/auth-service/proto/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type Client struct {
	conn    *grpc.ClientConn
	client  authpb.AuthServiceClient
	logger  *zap.Logger
	timeout time.Duration
}

// NewClient creates a new gRPC client for auth-service
func NewClient(address string, timeout time.Duration, logger *zap.Logger) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                60 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, err
	}

	// Test connection
	_, err = conn.NewStream(ctx, &grpc.StreamDesc{}, "/grpc.health.v1.Health/Check", grpc.WaitForReady(false))
	if err != nil && err.Error() != "rpc error: code = Unimplemented desc = unknown service grpc.health.v1.Health" {
		conn.Close()
		return nil, err
	}

	client := authpb.NewAuthServiceClient(conn)

	return &Client{
		conn:    conn,
		client:  client,
		logger:  logger,
		timeout: timeout,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// GetClient returns the underlying gRPC client
func (c *Client) GetClient() authpb.AuthServiceClient {
	return c.client
}

func (c *Client) GetTimeout() time.Duration {
	return c.timeout
}
