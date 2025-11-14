package grpcv1

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/message"
	chatv1 "github.com/hesoyamTM/nbf-protos/gen/go/chat"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ChatService interface {
	SendMessage(ctx context.Context, messageCh <-chan message.InputMessage) (<-chan message.Message, error)
}

type serverAPI struct {
	chatv1.UnimplementedChatServer
	chatService ChatService
}

func RegisterHandlers(server *grpc.Server, chatService ChatService) {
	chatv1.RegisterChatServer(server, &serverAPI{chatService: chatService})
}

func (s *serverAPI) SendMessage(stream chatv1.Chat_SendMessageServer) error {
	const op = "grpcv1.serverAPI.SendMessage"

	log, err := logger.LoggerFromCtx(stream.Context())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	inputMessageCh := make(chan message.InputMessage)

	go func() {
		log.Info("started reading messages from stream")
		defer log.Info("stopped reading messages from stream")

		defer close(inputMessageCh)
		for {
			msg, err := stream.Recv()
			if err != nil {
				return
			}

			userID, err := uuid.Parse(msg.GetUserId())
			if err != nil {
				log.Error("failed to parse user id", zap.Error(err))
				return
			}

			groupID, err := uuid.Parse(msg.GetGroupId())
			if err != nil {
				log.Error("failed to parse group id", zap.Error(err))
				return
			}

			inputMessageCh <- message.InputMessage{
				Text:    msg.Text,
				UserID:  userID,
				GroupID: groupID,
			}
		}
	}()

	outputMessageCh, err := s.chatService.SendMessage(stream.Context(), inputMessageCh)
	if err != nil {
		log.Error("failed to send message", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case message, ok := <-outputMessageCh:
			if !ok {
				return fmt.Errorf("%s: outputMessageCh closed", op)
			}

			user := &chatv1.User{
				Name: message.User.Name,
				Id:   message.User.ID.String(),
			}

			resp := &chatv1.SendMessageResponse{
				User:      user,
				Text:      message.Text,
				CreatedAt: timestamppb.New(message.CreatedAt),
			}

			if err := stream.Send(resp); err != nil {
				return fmt.Errorf("stream.Send: %w", err)
			}
		}
	}
}
