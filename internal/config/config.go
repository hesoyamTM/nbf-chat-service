// Package config contains the configuration for the application.
package config

import (
	"github.com/hesoyamTM/nbf-chat-service/internal/adapters/repository/messages/psql"
	"github.com/hesoyamTM/nbf-chat-service/internal/adapters/services"
	"github.com/hesoyamTM/nbf-chat-service/pkg/redpanda"
)

type Config struct {
	Env                   string                      `yaml:"env" env:"ENV" env-required:"true"`
	Grpc                  GrpcConfig                  `yaml:"grpc"`
	UserServiceConfig     services.UserServiceConfig  `yaml:"user_service"`
	GroupServiceConfig    services.GroupServiceConfig `yaml:"group_service"`
	PostgresMessageConfig psql.PostgresMessageConfig  `yaml:"postgres_message"`
	Redpanda              redpanda.RedpandaConfig     `yaml:"redpanda"`
}

type GrpcConfig struct {
	Host string `yaml:"host" env:"GRPC_HOST" env-required:"true"`
	Port int    `yaml:"port" env:"GRPC_PORT" env-required:"true"`
}
