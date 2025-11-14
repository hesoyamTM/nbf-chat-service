package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	cfgtool "github.com/hesoyamTM/nbf-auth/pkg/config"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"github.com/hesoyamTM/nbf-chat-service/internal/config"
	grpcv1 "github.com/hesoyamTM/nbf-chat-service/internal/delivery/grpc/v1"
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

	grpcApp := grpcv1.NewGrpcApp(ctx, cfg)
	go grpcApp.Run(ctx)

	log.Info("Application started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Info("Application is shutting down...")
	grpcApp.Shutdown(ctx)

	log.Info("Application stopped")
}
