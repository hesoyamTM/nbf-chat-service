package chat

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *ChatService) createChat(ctx context.Context, senderID, userID uuid.UUID) (uuid.UUID, error) {
	const op = "chat.createChat"

	if userID != uuid.Nil {
		return s.createChatByUser(ctx, senderID, userID)
	}

	return s.createChatByGroup(ctx, senderID, userID)
}

func (s *ChatService) createChatByGroup(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) (uuid.UUID, error) {
	const op = "chat.createChatbyGroup"

	group, err := s.getGroup(ctx, groupID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	chatName := group.Name
	chatID := uuid.New()

	_, err = s.messageRepository.CreateNewChatByUser(ctx, chatID, userID, groupID, chatName)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	for _, member := range group.Members {
		if member.ID == userID {
			continue
		}
		_, err = s.messageRepository.CreateNewChatByUser(ctx, chatID, member.ID, groupID, chatName)
		if err != nil {
			return uuid.Nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return chatID, nil
}

func (s *ChatService) createChatByUser(ctx context.Context, senderID, userID uuid.UUID) (uuid.UUID, error) {
	const op = "chat.createChatByUser"

	chatID := uuid.New()
	sender, err := s.userService.GetUser(ctx, senderID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}
	user, err := s.userService.GetUser(ctx, userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.messageRepository.CreateNewChatByUser(ctx,
		chatID,
		senderID,
		uuid.Nil,
		user.Name,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = s.messageRepository.CreateNewChatByUser(ctx,
		chatID,
		userID,
		uuid.Nil,
		sender.Name,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return chatID, nil
}
