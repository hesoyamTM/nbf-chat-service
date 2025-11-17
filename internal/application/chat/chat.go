// Package chat contains the application layer for the chat service.
package chat

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/group"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/message"
	"go.uber.org/zap"
)

type ChatService struct {
	userService       UserService
	groupService      GroupService
	messageRepository MessageRepository

	groupWorkers      map[uuid.UUID]*GroupWorker
	groupWorkersMutex sync.RWMutex
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

		groupWorkers:      make(map[uuid.UUID]*GroupWorker),
		groupWorkersMutex: sync.RWMutex{},
	}
}

func (s *ChatService) SendMessage(ctx context.Context, messageCh <-chan message.InputMessage) (<-chan message.Message, error) {
	const op = "chat.ChatService.SendMessage"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	outputMessageCh := make(chan message.Message)

	msg, ok := <-messageCh
	if !ok {
		log.Error("messageCh closed")
		return nil, fmt.Errorf("%s: messageCh closed", op)
	}

	s.groupWorkersMutex.Lock()
	defer s.groupWorkersMutex.Unlock()

	groupWorker, ok := s.groupWorkers[msg.GroupID]

	if !ok {
		log.Info("group not found, creating new group", zap.String("group_id", msg.GroupID.String()))

		newGroup, err := s.getGroup(ctx, msg.GroupID)
		if err != nil {
			log.Error("failed to get group", zap.Error(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		log.Info("got group", zap.String("group_id", msg.GroupID.String()))

		groupWorker, err = NewGroupWorker(ctx, newGroup, s.messageRepository)
		if err != nil {
			log.Error("failed to create group worker", zap.Error(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		log.Info("created group worker", zap.String("group_id", msg.GroupID.String()))

		s.groupWorkers[msg.GroupID] = groupWorker
		go groupWorker.Run(ctx)
		log.Info("started group worker", zap.String("group_id", msg.GroupID.String()))
	}

	go s.runListenMessages(ctx, groupWorker, msg.UserID, messageCh)
	err = groupWorker.AddConnection(ctx, msg.UserID, outputMessageCh)
	if err != nil {
		log.Error("failed to add connection", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	groupWorker.GroupCh <- msg

	return outputMessageCh, nil
}

// runListenMessages listens messages from users and redirect them to the group.
func (s *ChatService) runListenMessages(ctx context.Context, groupWorker *GroupWorker, userID uuid.UUID, inputMessageCh <-chan message.InputMessage) {
	for inputMsg := range inputMessageCh {
		groupWorker.GroupCh <- inputMsg
	}
	s.disconnectUser(ctx, groupWorker, userID)
}

// disconnectUser disconnects the user from the group. If the user is the last one, the group is disposed.
func (s *ChatService) disconnectUser(ctx context.Context, groupWorker *GroupWorker, userID uuid.UUID) {
	groupWorker.RemoveConnection(ctx, userID)

	if groupWorker.ConnectionLen() == 0 {
		s.groupWorkersMutex.Lock()
		defer s.groupWorkersMutex.Unlock()

		groupWorker.Dispose(ctx)
		delete(s.groupWorkers, groupWorker.Group().ID)
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
