package chat

import (
	"context"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/message"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/user"
)

type UserService interface {
	GetUsers(ctx context.Context, ids []uuid.UUID) ([]user.User, error)
}

type GroupService interface {
	GetGroup(ctx context.Context, id uuid.UUID) (string, error)
	GetGroupMembers(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error)
}

type MessageRepository interface {
	Save(ctx context.Context, message message.Message) error
	GetAll(ctx context.Context, groupID uuid.UUID) ([]message.Message, error)
}
