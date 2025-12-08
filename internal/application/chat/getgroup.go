package chat

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/group"
	"go.uber.org/zap"
)

// getGroup returns the group with the given id.
func (s *ChatService) getGroup(ctx context.Context, groupID uuid.UUID) (group.Group, error) {
	const op = "chat.ChatService.getGroup"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return group.Group{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Get group", zap.String("group_id", groupID.String()))

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
