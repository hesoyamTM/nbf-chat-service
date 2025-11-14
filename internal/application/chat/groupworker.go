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

type GroupWorker struct {
	connectionsMutex sync.RWMutex
	connections      map[uuid.UUID]chan message.Message
	group            group.Group

	GroupCh chan message.InputMessage
}

type Connection struct {
	inputChan  chan message.InputMessage
	outputChan chan message.Message
}

func NewGroupWorker(ctx context.Context, newGroup group.Group) (*GroupWorker, error) {
	const op = "chat.NewGroupWorker"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("New group worker for group", zap.String("group_id", newGroup.ID.String()))

	return &GroupWorker{
		connections:      make(map[uuid.UUID]chan message.Message),
		connectionsMutex: sync.RWMutex{},
		group:            newGroup,

		GroupCh: make(chan message.InputMessage),
	}, nil
}

func (g *GroupWorker) AddConnection(ctx context.Context, userID uuid.UUID, messageCh chan message.Message) error {
	const op = "chat.GroupWorker.AddConnection"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Add connection for user",
		zap.String("user_id", userID.String()),
		zap.String("group_id", g.group.ID.String()),
	)

	g.connectionsMutex.Lock()
	defer g.connectionsMutex.Unlock()

	g.connections[userID] = messageCh
	return nil
}

func (g *GroupWorker) RemoveConnection(ctx context.Context, userID uuid.UUID) {
	const op = "chat.GroupWorker.RemoveConnection"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return
	}

	log.Info("Remove connection for user",
		zap.String("user_id", userID.String()),
		zap.String("group_id", g.group.ID.String()),
	)

	g.connectionsMutex.Lock()
	defer g.connectionsMutex.Unlock()

	delete(g.connections, userID)
}

func (g *GroupWorker) Run(ctx context.Context) {
	const op = "chat.GroupWorker.Run"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return
	}

	log.Info("Start group worker for group", zap.String("group_id", g.group.ID.String()))

	for {
		select {
		case inputMsg := <-g.GroupCh:

			for userID, outputChan := range g.connections {
				if userID == inputMsg.UserID {
					continue
				}

				member, ok := g.group.GetMember(userID)

				if !ok {
					continue
				}

				outputChan <- message.NewMessage(member, inputMsg.GroupID, inputMsg.Text)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (g *GroupWorker) ConnectionLen() int {
	g.connectionsMutex.RLock()
	defer g.connectionsMutex.RUnlock()

	return len(g.connections)
}

func (g *GroupWorker) Group() group.Group {
	return g.group
}

func (g *GroupWorker) Dispose(ctx context.Context) error {
	const op = "chat.GroupWorker.Dispose"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Dispose group worker for group", zap.String("group_id", g.group.ID.String()))

	g.connectionsMutex.Lock()
	defer g.connectionsMutex.Unlock()

	for _, outputChan := range g.connections {
		close(outputChan)
	}

	close(g.GroupCh)
	return nil
}
