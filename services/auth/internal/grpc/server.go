package grpc

import (
	"context"

	"auth-service/internal/service"
	authpb "tech-ip-sem2/pkg"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
	authService *service.AuthService
}

func NewAuthServer(authService *service.AuthService) *AuthServer {
	return &AuthServer{
		authService: authService,
	}
}

func (s *AuthServer) Verify(ctx context.Context, req *authpb.VerifyRequest) (*authpb.VerifyResponse, error) {
	println("[gRPC] Verify called with token:", req.Token)

	token := req.GetToken()
	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "token is empty")
	}

	username, valid := s.authService.Verify(token)

	if !valid {
		return &authpb.VerifyResponse{
			Valid: false,
			Error: "invalid token",
		}, nil
	}

	return &authpb.VerifyResponse{
		Valid:   true,
		Subject: username,
	}, nil
}
