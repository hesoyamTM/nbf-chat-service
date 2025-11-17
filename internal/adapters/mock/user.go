package mock

import (
	"context"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/user"
)

type UserService struct{}

func (s *UserService) GetUsers(ctx context.Context, userIDs []uuid.UUID) ([]user.User, error) {
	return []user.User{
		{
			ID:   uuid.MustParse("d5313639-46cf-42d1-9c23-1cd19c8dcfb9"),
			Name: "test1",
		},
		{
			ID:   uuid.MustParse("d5313639-46cf-42d1-9c23-1cd19c8dcfb8"),
			Name: "test2",
		},
		{
			ID:   uuid.MustParse("d5313639-46cf-42d1-9c23-1cd19c8dcfb7"),
			Name: "test3",
		},
	}, nil
}
