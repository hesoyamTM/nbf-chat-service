// Package chat contains the application layer for the chat service.
package chat

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/chat"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/message"
	"go.uber.org/zap"
)

type ChatService struct {
	userService       UserService
	groupService      GroupService
	messageRepository MessageRepository

	chatWorkers      map[uuid.UUID]*ChatWorker
	chatWorkersMutex sync.RWMutex
}

func NewChatService(
	userService UserService,
	groupService GroupService,
	messageRepository MessageRepository,
) *ChatService {
	return &ChatService{
		userService:       userService,
		groupService:      groupService,
		messageRepository: messageRepository,

		chatWorkers:      make(map[uuid.UUID]*ChatWorker),
		chatWorkersMutex: sync.RWMutex{},
	}
}

func (s *ChatService) SendMessage(
	ctx context.Context,
	senderID uuid.UUID,
	userID uuid.UUID,
	groupID uuid.UUID,
	chatID uuid.UUID,
	messageCh <-chan message.InputMessage,
) (<-chan message.Message, error) {
	const op = "chat.ChatService.SendMessage"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	outputMessageCh := make(chan message.Message)

	s.chatWorkersMutex.Lock()
	defer s.chatWorkersMutex.Unlock()

	if chatID == uuid.Nil {
		chatID, err = s.getChatID(ctx, senderID, userID, groupID)
		if err != nil {
			chatID, err = s.createChat(ctx, senderID, userID)
			if err != nil {
				log.Error("failed to create chat", zap.Error(err))
				return nil, fmt.Errorf("%s: %w", op, err)
			}
		}
	}

	chatWorker, ok := s.chatWorkers[chatID]

	if !ok {
		log.Info("chat worker not found, creating new chat worker", zap.String("chat_id", chatID.String()))

		newChat, err := s.getChat(ctx, senderID, chatID, groupID, userID)
		if err != nil {
			log.Error("failed to get chat", zap.Error(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		chatWorker, err = NewChatWorker(ctx, newChat, s.messageRepository)
		if err != nil {
			log.Error("failed to create chat worker", zap.Error(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		log.Info("created chat worker", zap.String("chat_id", chatID.String()))

		s.chatWorkers[chatID] = chatWorker
		go chatWorker.Run(ctx)
		log.Info("started chat worker", zap.String("chat_id", chatID.String()))
	}

	go s.runListenMessages(ctx, chatWorker, senderID, messageCh)
	err = chatWorker.AddConnection(ctx, senderID, outputMessageCh)
	if err != nil {
		log.Error("failed to add connection", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return outputMessageCh, nil
}

// GetChatsByUser returns chats that the user is in.
func (s *ChatService) GetChatsByUser(ctx context.Context, userID uuid.UUID) ([]chat.Chat, error) {
	const op = "chat.ChatService.GetChatsByUser"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Get chats by user", zap.String("user_id", userID.String()))

	chats, err := s.messageRepository.GetChatsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	chatsWithMember := make([]chat.Chat, 0, len(chats))

	for _, chatDialog := range chats {
		chat, err := s.getChatWithMembers(ctx, chatDialog)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		chatsWithMember = append(chatsWithMember, chat)
	}

	return chatsWithMember, nil
}

// DeleteUserFromChatByGroup deletes the user from the group.
func (s *ChatService) DeleteUserFromChatByGroup(ctx context.Context, userID, groupID uuid.UUID) error {
	const op = "chat.DeleteUserFromChatByGroup"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Delete user from chat",
		zap.String("user_id", userID.String()),
		zap.String("group_id", groupID.String()),
	)

	if err := s.messageRepository.DeleteUserFromChatByGroup(ctx, userID, groupID); err != nil {
		log.Error("failed to delete user from chat",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("group_id", groupID.String()),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// DeleteChatByGroup deletes the chat.
func (s *ChatService) DeleteChatByGroup(ctx context.Context, groupID uuid.UUID) error {
	const op = "chat.DeleteChatByGroup"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Delete chat",
		zap.String("group_id", groupID.String()),
	)

	if err := s.messageRepository.DeleteChatByGroup(ctx, groupID); err != nil {
		log.Error("failed to delete chat",
			zap.Error(err),
			zap.String("group_id", groupID.String()),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// AddUserToChatByGroup adds the user to the group.
func (s *ChatService) AddUserToChatByGroup(ctx context.Context, userID, groupID uuid.UUID, chatName string) error {
	const op = "chat.AddUserToChatByGroup"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Add user to chat",
		zap.String("user_id", userID.String()),
		zap.String("group_id", groupID.String()),
	)

	chatID, err := s.messageRepository.GetChatIDByGroupID(ctx, groupID)
	if err != nil {
		log.Error("failed to get chat id by group id",
			zap.Error(err),
			zap.String("group_id", groupID.String()),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	if _, err := s.messageRepository.CreateNewChatByUser(ctx, chatID, userID, groupID, chatName); err != nil {
		log.Error("failed to add user to chat",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("group_id", groupID.String()),
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// runListenMessages listens messages from users and redirect them to the chat.
func (s *ChatService) runListenMessages(ctx context.Context, chatWorker *ChatWorker, userID uuid.UUID, inputMessageCh <-chan message.InputMessage) {
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return
	}

	log.Info("Started listening messages from user")

	for inputMsg := range inputMessageCh {
		chatWorker.ChatCh <- inputMsg
	}
	log.Info("Stopped listening messages from user")
	s.disconnectUser(ctx, chatWorker, userID)
	log.Info("Disconnected user from chat")
}

// disconnectUser disconnects the user from the group. If the user is the last one, the group is disposed.
func (s *ChatService) disconnectUser(ctx context.Context, chatWorker *ChatWorker, userID uuid.UUID) {
	chatWorker.RemoveConnection(ctx, userID)

	if chatWorker.ConnectionLen() == 0 {
		s.chatWorkersMutex.Lock()
		defer s.chatWorkersMutex.Unlock()

		chatWorker.Dispose(ctx)
		delete(s.chatWorkers, chatWorker.Chat.ID)
	}
}
