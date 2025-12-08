package redpanda

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (c *Consumer) HandleMatcherGroupRequestAccepted(
	ctx context.Context,
	log *zap.Logger,
	message *sarama.ConsumerMessage,
) error {
	log.Info("HandleMatcherGroupRequestAccepted")

	var req MatcherGroupRequestAccepted
	if err := json.Unmarshal(message.Value, &req); err != nil {
		log.Error("failed to unmarshal message",
			zap.Error(err),
		)
		return err
	}

	userID, err := uuid.Parse(req.Request.UserID)
	if err != nil {
		log.Error("failed to parse user id",
			zap.Error(err),
		)
		return err
	}

	ownerID, err := uuid.Parse(req.Group.OwnerID)
	if err != nil {
		log.Error("failed to parse owner id",
			zap.Error(err),
		)
		return err
	}

	groupID, err := uuid.Parse(req.Request.GroupID)
	if err != nil {
		log.Error("failed to parse group id",
			zap.Error(err),
		)
		return err
	}

	if err := c.chatService.AddUserToChatByGroup(ctx, userID, groupID, req.Group.Parameters.Name); err != nil {
		log.Error("failed to add user to chat",
			zap.Error(err),
		)
		return err
	}
	if err := c.chatService.AddUserToChatByGroup(ctx, ownerID, groupID, req.Group.Parameters.Name); err != nil {
		log.Warn("failed to add owner to chat",
			zap.Error(err),
		)
	}

	return nil
}

func (c *Consumer) HandleMatcherGroupUserLeft(
	ctx context.Context,
	log *zap.Logger,
	message *sarama.ConsumerMessage,
) error {
	log.Info("HandleMatcherGroupUserLeft")

	var req MatcherGroupUserLeft
	if err := json.Unmarshal(message.Value, &req); err != nil {
		log.Error("failed to unmarshal message",
			zap.Error(err),
		)
		return err
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		log.Error("failed to parse user id",
			zap.Error(err),
		)
		return err
	}

	groupID, err := uuid.Parse(req.Group.GroupID)
	if err != nil {
		log.Error("failed to parse group id",
			zap.Error(err),
		)
		return err
	}

	if req.Remained <= 1 {
		if err := c.chatService.DeleteChatByGroup(ctx, groupID); err != nil {
			log.Error("failed to delete chat",
				zap.Error(err),
			)
			return err
		}
	}

	if err := c.chatService.DeleteUserFromChatByGroup(ctx, userID, groupID); err != nil {
		log.Error("failed to delete user from chat",
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (c *Consumer) HandleMatcherGroupUserKicked(
	ctx context.Context,
	log *zap.Logger,
	message *sarama.ConsumerMessage,
) error {
	log.Info("HandleMatcherGroupUserKicked")

	var req MatcherGroupUserKicked
	if err := json.Unmarshal(message.Value, &req); err != nil {
		log.Error("failed to unmarshal message",
			zap.Error(err),
		)
		return err
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		log.Error("failed to parse user id",
			zap.Error(err),
		)
		return err
	}

	groupID, err := uuid.Parse(req.Group.GroupID)
	if err != nil {
		log.Error("failed to parse group id",
			zap.Error(err),
		)
		return err
	}

	if req.Remained <= 1 {
		if err := c.chatService.DeleteChatByGroup(ctx, groupID); err != nil {
			log.Error("failed to delete chat",
				zap.Error(err),
			)
			return err
		}
	}

	if err := c.chatService.DeleteUserFromChatByGroup(ctx, userID, groupID); err != nil {
		log.Error("failed to delete user from chat",
			zap.Error(err),
		)
		return err
	}

	return nil
}
