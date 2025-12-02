// Package chat contains the application layer for the chat service.
package chat

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/chat"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/group"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/message"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/user"
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
func (s *ChatService) GetChatsByUser(ctx context.Context, userID uuid.UUID) ([]chat.ChatDialog, error) {
	const op = "chat.ChatService.GetChatsByUser"

	chats, err := s.messageRepository.GetChatsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return chats, nil
}

// runListenMessages listens messages from users and redirect them to the chat.
func (s *ChatService) runListenMessages(ctx context.Context, chatWorker *ChatWorker, userID uuid.UUID, inputMessageCh <-chan message.InputMessage) {
	for inputMsg := range inputMessageCh {
		chatWorker.ChatCh <- inputMsg
	}
	s.disconnectUser(ctx, chatWorker, userID)
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

// getGroup returns the group with the given id.
func (s *ChatService) getGroup(ctx context.Context, groupID uuid.UUID) (group.Group, error) {
	const op = "chat.ChatService.getGroup"

	groupName, err := s.groupService.GetGroup(ctx, groupID)
	if err != nil {
		return group.Group{}, fmt.Errorf("%s: %w", op, err)
	}

	memberIDs, err := s.groupService.GetGroupMembers(ctx, groupID)
	if err != nil {
		return group.Group{}, fmt.Errorf("%s: %w", op, err)
	}

	members, err := s.userService.GetUsers(ctx, memberIDs)
	if err != nil {
		return group.Group{}, fmt.Errorf("%s: %w", op, err)
	}

	return group.NewGroup(groupID, groupName, members), nil
}

// getChat returns the chat with the given ids. Only one of chatID, groupID, and userID should be non-nil.
func (s *ChatService) getChat(ctx context.Context, senderID, chatID, groupID, userID uuid.UUID) (chat.Chat, error) {
	const op = "chat.getChat"

	if chatID == uuid.Nil {
		chatDialog, err := s.messageRepository.GetChatByID(ctx, senderID, chatID)
		if err != nil {
			return chat.Chat{}, fmt.Errorf("%s: %w", op, err)
		}
		return s.getChatWithMembers(ctx, chatDialog)
	}

	if groupID == uuid.Nil {
		chatDialog, err := s.messageRepository.GetChatByGroup(ctx, senderID, groupID)
		if err != nil {
			return chat.Chat{}, fmt.Errorf("%s: %w", op, err)
		}

		return s.getChatWithMembers(ctx, chatDialog)
	}

	if userID == uuid.Nil {
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
	}

	return chat.Chat{
		ID:      chatDialog.ID,
		Name:    chatDialog.Name,
		Members: mapMembers,
	}, nil
}

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

	_, err := s.messageRepository.CreateNewChatByUser(ctx, chatID, senderID, uuid.Nil, userID.String())
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = s.messageRepository.CreateNewChatByUser(ctx, chatID, userID, uuid.Nil, senderID.String())
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return chatID, nil
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
