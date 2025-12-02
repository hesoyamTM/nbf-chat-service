// Package grpcv1 provides the implementation of the gRPC server.
package grpcv1

import (
	"context"
	"fmt"
	"net"

	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"github.com/hesoyamTM/nbf-chat-service/internal/adapters/repository/messages/psql"
	"github.com/hesoyamTM/nbf-chat-service/internal/adapters/services"
	"github.com/hesoyamTM/nbf-chat-service/internal/application/chat"
	"github.com/hesoyamTM/nbf-chat-service/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GrpcApp is a gRPC server.
type GrpcApp struct {
	server *grpc.Server
	host   string
	port   int
}

// NewGrpcApp returns a new instance of GrpcApp.
func NewGrpcApp(ctx context.Context, cfg *config.Config) *GrpcApp {
	const op = "grpcv1.NewGrpcApp"

	loggingUnaryInterceptor, err := logger.NewLoggingInterceptor(ctx)
	if err != nil {
		panic(fmt.Errorf("%s: %w", op, err))
	}

	loggingStreamInterceptor, err := logger.NewLoggingStreamServerInterceptor(ctx)
	if err != nil {
		panic(fmt.Errorf("%s: %w", op, err))
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingUnaryInterceptor),
		grpc.StreamInterceptor(loggingStreamInterceptor),
	)

	userService, err := services.NewUserService(cfg.UserServiceConfig)
	if err != nil {
		panic(fmt.Errorf("%s: %w", op, err))
	}
	groupService, err := services.NewGroupService(cfg.GroupServiceConfig)
	if err != nil {
		panic(fmt.Errorf("%s: %w", op, err))
	}
	messageRepository, err := psql.NewPostgresMessageRepository(ctx, cfg.PostgresMessageConfig)
	if err != nil {
		panic(fmt.Errorf("%s: %w", op, err))
	}

	chatService := chat.NewChatService(userService, groupService, messageRepository)

	RegisterHandlers(grpcServer, chatService)
	reflection.Register(grpcServer)

	return &GrpcApp{
		server: grpcServer,
		host:   cfg.Grpc.Host,
		port:   cfg.Grpc.Port,
	}
}

// Run starts the gRPC server.
func (g *GrpcApp) Run(ctx context.Context) error {
	const op = "grpcv1.GrpcApp.Run"

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", g.host, g.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	g.server.Serve(lis)

	return nil
}

// Shutdown stops the gRPC server.
func (g *GrpcApp) Shutdown(ctx context.Context) {
	const op = "grpcv1.GrpcApp.Shutdown"

	g.server.GracefulStop()
}
