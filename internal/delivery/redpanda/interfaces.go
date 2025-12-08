package redpanda

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ChatService interface {
	DeleteUserFromChatByGroup(ctx context.Context, userID, groupID uuid.UUID) error
	DeleteChatByGroup(ctx context.Context, groupID uuid.UUID) error
	AddUserToChatByGroup(ctx context.Context, userID, groupID uuid.UUID, chatName string) error
}

type TopicHandler func(
	ctx context.Context,
	log *zap.Logger,
	message *sarama.ConsumerMessage,
) error
