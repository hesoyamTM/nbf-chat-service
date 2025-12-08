package chat

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/chat"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/message"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/user"
)

type UserService interface {
	GetUsers(ctx context.Context, ids []uuid.UUID) ([]user.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (user.User, error)
}

type GroupService interface {
	GetGroup(ctx context.Context, id uuid.UUID) (string, error)
	GetGroupMembers(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error)
}

type MessageRepository interface {
	Save(ctx context.Context, message message.Message) error
	GetAll(ctx context.Context, chatID uuid.UUID) ([]message.Message, error)

	CreateNewChatByUser(ctx context.Context, chatID, userID, groupID uuid.UUID, chatName string) (uuid.UUID, error)

	GetChatByUser(ctx context.Context, senderID, userID uuid.UUID) (chat.Chat, error)
	GetChatsByUser(ctx context.Context, userID uuid.UUID) ([]chat.Chat, error)
	GetChatByGroup(ctx context.Context, senderID, groupID uuid.UUID) (chat.Chat, error)
	GetChatByID(ctx context.Context, senderID, chatID uuid.UUID) (chat.Chat, error)

	GetChatIDByUser(ctx context.Context, senderID, userID uuid.UUID) (uuid.UUID, error)
	GetChatIDByGroup(ctx context.Context, senderID, groupID uuid.UUID) (uuid.UUID, error)

	IncrementUnreadCount(ctx context.Context, userID, chatID uuid.UUID, receiverID uuid.UUID) error
	SetLastReadAt(ctx context.Context, userID, chatID uuid.UUID, lastReadAt time.Time) error

	CreateChatByGroup(ctx context.Context, groupID uuid.UUID) (uuid.UUID, error)
	DeleteUserFromChatByGroup(ctx context.Context, userID, groupID uuid.UUID) error
	DeleteChatByGroup(ctx context.Context, groupID uuid.UUID) error
	GetChatIDByGroupID(ctx context.Context, groupID uuid.UUID) (uuid.UUID, error)
}
