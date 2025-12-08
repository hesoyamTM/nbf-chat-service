package chat

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/chat"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/user"
	"go.uber.org/zap"
)

// getChat returns the chat with the given ids. Only one of chatID, groupID, and userID should be non-nil.
func (s *ChatService) getChat(ctx context.Context, senderID, chatID, groupID, userID uuid.UUID) (chat.Chat, error) {
	const op = "chat.getChat"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return chat.Chat{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Get chat", zap.String("chat_id", chatID.String()), zap.String("group_id", groupID.String()), zap.String("user_id", userID.String()))

	if chatID != uuid.Nil {
		chatDialog, err := s.messageRepository.GetChatByID(ctx, senderID, chatID)
		if err != nil {
			return chat.Chat{}, fmt.Errorf("%s: %w", op, err)
		}
		return s.getChatWithMembers(ctx, chatDialog)
	}

	if groupID != uuid.Nil {
		chatDialog, err := s.messageRepository.GetChatByGroup(ctx, senderID, groupID)
		if err != nil {
			return chat.Chat{}, fmt.Errorf("%s: %w", op, err)
		}

		return s.getChatWithMembers(ctx, chatDialog)
	}

	if userID != uuid.Nil {
		chatDialog, err := s.messageRepository.GetChatByUser(ctx, senderID, userID)
		if err != nil {
			return chat.Chat{}, fmt.Errorf("%s: %w", op, err)
		}
		return s.getChatWithMembers(ctx, chatDialog)
	}

	return chat.Chat{}, fmt.Errorf("%s: chatID, groupID, and userID are all nil", op)
}

func (s *ChatService) getChatWithMembers(ctx context.Context, chatDialog chat.Chat) (chat.Chat, error) {
	const op = "chat.getChatWithMembers"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return chat.Chat{}, fmt.Errorf("%s: %w", op, err)
	}

	membersIDs := make([]uuid.UUID, 0, len(chatDialog.Members))
	for _, member := range chatDialog.Members {
		membersIDs = append(membersIDs, member.ID)
	}

	members, err := s.userService.GetUsers(ctx, membersIDs)
	if err != nil {
		return chat.Chat{}, fmt.Errorf("%s: %w", op, err)
	}

	mapMembers := make(map[uuid.UUID]user.User)
	for _, member := range members {
		mapMembers[member.ID] = member
		log.Info("Member", zap.String("id", member.ID.String()), zap.String("name", member.Name))
	}

	return chat.Chat{
		ID:      chatDialog.ID,
		Name:    chatDialog.Name,
		Members: mapMembers,
	}, nil
}

func (s *ChatService) getChatID(ctx context.Context, senderID, userID, groupID uuid.UUID) (uuid.UUID, error) {
	const op = "chat.getChatID"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	if userID != uuid.Nil {
		chatID, err := s.messageRepository.GetChatIDByUser(ctx, senderID, userID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("%s: %w", op, err)
		}
		return chatID, nil
	}

	if groupID != uuid.Nil {
		chatID, err := s.messageRepository.GetChatIDByGroup(ctx, senderID, groupID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("%s: %w", op, err)
		}
		return chatID, nil
	}

	log.Error("userID and groupID are all nil")
	return uuid.Nil, fmt.Errorf("%s: userID and groupID are all nil", op)
}
