package grpcv1

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/chat"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/message"
	chatv1 "github.com/hesoyamTM/nbf-protos/gen/go/chat"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ChatService interface {
	SendMessage(
		ctx context.Context,
		senderID uuid.UUID, // userID that getting the message
		userID uuid.UUID, // userID that getting the message
		groupID uuid.UUID, // groupID that getting the message
		chatID uuid.UUID, // chatID that getting the message
		messageCh <-chan message.InputMessage,
	) (<-chan message.Message, error)
	GetChatsByUser(ctx context.Context, userID uuid.UUID) ([]chat.ChatDialog, error)
}

type serverAPI struct {
	chatv1.UnimplementedChatServiceServer
	chatService ChatService
}

func RegisterHandlers(server *grpc.Server, chatService ChatService) {
	chatv1.RegisterChatServiceServer(server, &serverAPI{chatService: chatService})
}

func (s *serverAPI) SendMessage(stream chatv1.ChatService_SendMessageServer) error {
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

			// chatID, err := uuid.Parse(msg.GetChatId())
			// if err != nil {
			// 	log.Error("failed to parse chat id", zap.Error(err))
			// 	return
			// }

			inputMessageCh <- message.InputMessage{
				Text:   msg.Text,
				UserID: userID,
				// ChatID: chatID,
			}
		}
	}()

	senderID, err := parseUID(stream.Context())
	if err != nil {
		log.Error("failed to parse uid", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	userID, groupID, chatID, err := parseMetdata(stream.Context())
	if err != nil {
		log.Error("failed to parse metadata", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	outputMessageCh, err := s.chatService.SendMessage(
		stream.Context(),
		senderID,
		userID,
		groupID,
		chatID,
		inputMessageCh,
	)
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

func (s *serverAPI) GetChatsByUser(ctx context.Context, req *chatv1.GetChatByUserRequest) (*chatv1.GetChatByUserResponse, error) {
	const op = "grpcv1.serverAPI.GetChatByUser"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		log.Error("failed to parse user id", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	chats, err := s.chatService.GetChatsByUser(ctx, userID)
	if err != nil {
		log.Error("failed to get chats by user", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	resp := &chatv1.GetChatByUserResponse{
		Chat: make([]*chatv1.Chat, len(chats)),
	}

	for i, chat := range chats {
		resp.Chat[i] = &chatv1.Chat{
			Id:   chat.ID.String(),
			Name: chat.Name,
		}
	}
	return resp, nil
}

func parseUID(ctx context.Context) (uuid.UUID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, fmt.Errorf("metadata not found")
	}

	var err error
	var userID uuid.UUID

	userIDmd, ok := md["uid"]
	if !ok {
		return uuid.Nil, fmt.Errorf("uid not found")
	}
	userID, err = uuid.Parse(userIDmd[0])
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse uid: %w", err)
	}

	return userID, nil
}

func parseMetdata(ctx context.Context) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("metadata not found")
	}

	var err error

	userIDmd, ok := md["user_id"]
	var userID uuid.UUID
	if !ok {
		userID = uuid.Nil
	} else {
		userID, err = uuid.Parse(userIDmd[0])
		if err != nil {
			return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("failed to parse user id: %w", err)
		}
	}

	groupIDmd, ok := md["group_id"]
	var groupID uuid.UUID
	if !ok {
		groupID = uuid.Nil
	} else {
		groupID, err = uuid.Parse(groupIDmd[0])
		if err != nil {
			return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("failed to parse group id: %w", err)
		}
	}

	chatIDmd, ok := md["chat_id"]
	var chatID uuid.UUID
	if !ok {
		chatID = uuid.Nil
	} else {
		chatID, err = uuid.Parse(chatIDmd[0])
		if err != nil {
			return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("failed to parse chat id: %w", err)
		}
	}

	if userID == uuid.Nil && groupID == uuid.Nil && chatID == uuid.Nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("user_id, group_id, and chat_id are all nil")
	}

	return userID, groupID, chatID, nil
}
