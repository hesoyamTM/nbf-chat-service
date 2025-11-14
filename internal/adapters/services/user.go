// Package services provides adapters for services.
package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/user"
	userv1 "github.com/hesoyamTM/nbf-protos/gen/go/user"
	"google.golang.org/grpc"
)

// UserServiceConfig is a configuration for the UserService.
type UserServiceConfig struct {
	Address string
}

// UserService is a service that provides information about users.
type UserService struct {
	api userv1.UserClient
}

// NewUserService returns a new instance of UserService.
func NewUserService(cfg UserServiceConfig) (*UserService, error) {
	const op = "services.NewUserService"

	conn, err := grpc.NewClient(
		cfg.Address,
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	api := userv1.NewUserClient(conn)

	return &UserService{
		api: api,
	}, nil
}

// GetUsers returns information about users.
func (s *UserService) GetUsers(ctx context.Context, userIDs []uuid.UUID) ([]user.User, error) {
	const op = "services.UserService.GetUser"

	ids := make([]string, len(userIDs))
	for i, id := range userIDs {
		ids[i] = id.String()
	}

	resp, err := s.api.GetUsers(ctx, &userv1.GetUsersRequest{
		Ids: ids,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	users := make([]user.User, len(resp.Users))

	for i, respUser := range resp.Users {
		users[i] = user.User{
			ID:   userIDs[i],
			Name: respUser.Name,
		}
	}

	return users, nil
}
