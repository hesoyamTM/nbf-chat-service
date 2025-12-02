package chat

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/chat"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/message"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/user"
	"go.uber.org/zap"
)

type ChatWorker struct {
	connectionsMutex sync.RWMutex
	connections      map[uuid.UUID]chan message.Message

	Chat chat.Chat

	ChatCh chan message.InputMessage

	messageRepository MessageRepository
}

type Connection struct {
	inputChan  chan message.InputMessage
	outputChan chan message.Message
}

func NewChatWorker(ctx context.Context, newChat chat.Chat, messageRepository MessageRepository) (*ChatWorker, error) {
	const op = "chat.NewChatWorker"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("New chat worker for chat", zap.String("chat_id", newChat.ID.String()))

	return &ChatWorker{
		Chat:              newChat,
		ChatCh:            make(chan message.InputMessage),
		messageRepository: messageRepository,
		connections:       make(map[uuid.UUID]chan message.Message),
		connectionsMutex:  sync.RWMutex{},
	}, nil
}

func (g *ChatWorker) AddConnection(ctx context.Context, userID uuid.UUID, messageCh chan message.Message) error {
	const op = "chat.ChatWorker.AddConnection"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Add connection for user", zap.String("user_id", userID.String()), zap.String("chat_id", g.Chat.ID.String()))

	g.connectionsMutex.Lock()
	g.connections[userID] = messageCh
	g.connectionsMutex.Unlock()

	log.Info("Added connection for user", zap.String("user_id", userID.String()))

	messages, err := g.messageRepository.GetAll(ctx, g.Chat.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		log.Info("Sending messages to new connection", zap.String("user_id", userID.String()))
		for _, message := range messages {
			messageCh <- message
		}
		log.Info("Sent messages to new connection", zap.String("user_id", userID.String()))
	}()

	return nil
}

func (g *ChatWorker) RemoveConnection(ctx context.Context, userID uuid.UUID) {
	const op = "chat.ChatWorker.RemoveConnection"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return
	}

	log.Info("Remove connection for user", zap.String("user_id", userID.String()), zap.String("chat_id", g.Chat.ID.String()))

	g.connectionsMutex.Lock()
	defer g.connectionsMutex.Unlock()

	delete(g.connections, userID)
}

func (g *ChatWorker) Run(ctx context.Context) {
	const op = "chat.ChatWorker.Run"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return
	}

	log.Info("Start chat worker for chat", zap.String("chat_id", g.Chat.ID.String()))

	for {
		select {
		case inputMsg, ok := <-g.ChatCh:
			if !ok {
				log.Info("Chat worker stopped")
				return
			}

			log.Info("Chat worker received message", zap.String("from", inputMsg.UserID.String()))

			member, ok := g.Chat.Members[inputMsg.UserID]

			if !ok {
				continue
			}

			user := user.User{ID: inputMsg.UserID, Name: member.Name}
			msg := message.NewMessage(user, g.Chat.ID, inputMsg.Text)
			g.messageRepository.Save(ctx, msg)

			g.connectionsMutex.RLock()
			for _, outputChan := range g.connections {
				outputChan <- msg
			}
			g.connectionsMutex.RUnlock()
		case <-ctx.Done():
			log.Info("Group worker stopped")
			return
		}
	}
}

func (g *ChatWorker) ConnectionLen() int {
	g.connectionsMutex.RLock()
	defer g.connectionsMutex.RUnlock()

	return len(g.connections)
}

func (g *ChatWorker) Dispose(ctx context.Context) error {
	const op = "chat.ChatWorker.Dispose"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Dispose group worker for group", zap.String("group_id", g.Chat.ID.String()))

	g.connectionsMutex.Lock()
	defer g.connectionsMutex.Unlock()

	for _, outputChan := range g.connections {
		close(outputChan)
	}

	close(g.ChatCh)
	return nil
}
