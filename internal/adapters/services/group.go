// Package services provides adapters for services.
package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	matcherv1 "github.com/hesoyamTM/nbf-protos/gen/go/matcher"
	"google.golang.org/grpc"
)

// GroupServiceConfig is a configuration for the GroupService.
type GroupServiceConfig struct {
	Address string
}

// GroupService is a service that provides information about groups.
type GroupService struct {
	api matcherv1.GroupQueryServiceClient
}

// NewGroupService returns a new instance of GroupService.
func NewGroupService(cfg GroupServiceConfig) (*GroupService, error) {
	const op = "services.NewGroupService"

	conn, err := grpc.NewClient(
		cfg.Address,
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	api := matcherv1.NewGroupQueryServiceClient(conn)

	return &GroupService{
		api: api,
	}, nil
}

// GetGroup returns the name of the group.
func (s *GroupService) GetGroup(ctx context.Context, groupID uuid.UUID) (string, error) {
	const op = "services.GroupService.GetGroup"

	respGroup, err := s.api.GetGroup(ctx, &matcherv1.GetGroupRequest{
		GroupId: groupID.String(),
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return respGroup.GetParameters().GetName(), nil
}

// GetGroupMembers returns the member ids  of the group.
func (s *GroupService) GetGroupMembers(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	const op = "services.GroupService.GetGroupMembers"

	respMembers, err := s.api.ListGroupMembers(ctx, &matcherv1.ListGroupMembersRequest{
		GroupId: groupID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	membersIDs := make([]uuid.UUID, len(respMembers.GetMembers()))

	for i, member := range respMembers.GetMembers() {
		id, err := uuid.Parse(member.GetUserId())
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		membersIDs[i] = id
	}

	return membersIDs, nil
}
