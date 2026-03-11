package client

import (
	"context"
	"fmt"
	"time"

	"tech-ip-sem2/pkg"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type AuthClient struct {
	client  authpb.AuthServiceClient
	conn    *grpc.ClientConn
	timeout time.Duration
}

func NewAuthClient(addr string, timeout time.Duration) (*AuthClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth gRPC server: %w", err)
	}

	client := authpb.NewAuthServiceClient(conn)

	return &AuthClient{
		client:  client,
		conn:    conn,
		timeout: timeout,
	}, nil
}

func (c *AuthClient) Close() error {
	return c.conn.Close()
}

func (c *AuthClient) VerifyToken(ctx context.Context, token string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	println("[gRPC Client] Calling Verify with token:", token)

	resp, err := c.client.Verify(ctx, &authpb.VerifyRequest{
		Token: token,
	})

	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.DeadlineExceeded:
				return "", fmt.Errorf("auth service timeout (deadline exceeded)")
			case codes.Unavailable:
				return "", fmt.Errorf("auth service unavailable")
			case codes.Unauthenticated:
				return "", fmt.Errorf("token invalid: %v", st.Message())
			default:
				return "", fmt.Errorf("auth service error: %v", st.Message())
			}
		}
		return "", fmt.Errorf("failed to verify token: %w", err)
	}

	if !resp.GetValid() {
		return "", fmt.Errorf("token invalid: %s", resp.GetError())
	}

	return resp.GetSubject(), nil
}
