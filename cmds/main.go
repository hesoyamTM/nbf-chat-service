package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	cfgtool "github.com/hesoyamTM/nbf-auth/pkg/config"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"github.com/hesoyamTM/nbf-chat-service/internal/adapters/repository/messages/psql"
	"github.com/hesoyamTM/nbf-chat-service/internal/adapters/services"
	"github.com/hesoyamTM/nbf-chat-service/internal/application/chat"
	"github.com/hesoyamTM/nbf-chat-service/internal/config"
	grpcv1 "github.com/hesoyamTM/nbf-chat-service/internal/delivery/grpc/v1"
	"github.com/hesoyamTM/nbf-chat-service/internal/delivery/redpanda"
)

func main() {
	cfg := cfgtool.MustParseConfig[config.Config]()
	ctx, err := logger.SetupLogger(context.Background(), cfg.Env)
	if err != nil {
		panic(err)
	}
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		panic(err)
	}
	log.Debug("Logger is ready")

	userService, err := services.NewUserService(cfg.UserServiceConfig)
	if err != nil {
		panic(err)
	}
	groupService, err := services.NewGroupService(cfg.GroupServiceConfig)
	if err != nil {
		panic(err)
	}
	messageRepository, err := psql.NewPostgresMessageRepository(ctx, cfg.PostgresMessageConfig)
	if err != nil {
		panic(err)
	}

	chatService := chat.NewChatService(userService, groupService, messageRepository)

	grpcApp := grpcv1.NewGrpcApp(ctx, chatService, &cfg.Grpc)
	redpandaApp := redpanda.NewRedPandaApp(ctx, chatService, cfg.Redpanda)

	go grpcApp.Run(ctx)
	go redpandaApp.Run(ctx)

	log.Info("Application started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Info("Application is shutting down...")

	grpcApp.Shutdown(ctx)
	redpandaApp.Stop(ctx)

	log.Info("Application stopped")
}
