package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	authgrpc "auth-service/internal/grpc"
	"auth-service/internal/handler"
	"auth-service/internal/service"
	authpb "tech-ip-sem2/pkg"
	"tech-ip-sem2/shared/middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	httpPort := os.Getenv("AUTH_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}

	grpcPort := os.Getenv("AUTH_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	authService := service.NewAuthService()

	authHandler := handler.NewAuthHandler(authService)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)
	mux.HandleFunc("GET /v1/auth/verify", authHandler.Verify)

	httpHandler := middleware.RequestIDMiddleware(
		middleware.LoggingMiddleware(mux),
	)

	httpServer := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      httpHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	grpcListener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen gRPC: %v", err)
	}

	grpcServer := grpc.NewServer()

	authpb.RegisterAuthServiceServer(grpcServer, authgrpc.NewAuthServer(authService))

	reflection.Register(grpcServer)

	go func() {
		log.Printf("Auth HTTP service starting on port %s", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	go func() {
		log.Printf("Auth gRPC service starting on port %s", grpcPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Auth service...")

	grpcServer.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal("HTTP server forced to shutdown:", err)
	}

	log.Println("Auth service stopped")
}
