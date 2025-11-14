package mock

import (
	"context"

	"github.com/google/uuid"
)

type GroupService struct{}

func (s *GroupService) GetGroup(ctx context.Context, groupID uuid.UUID) (string, error) {
	return "test", nil
}

func (s *GroupService) GetGroupMembers(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	return []uuid.UUID{
		uuid.MustParse("d5313639-46cf-42d1-9c23-1cd19c8dcfb9"),
		uuid.MustParse("d5313639-46cf-42d1-9c23-1cd19c8dcfb8"),
		uuid.MustParse("d5313639-46cf-42d1-9c23-1cd19c8dcfb7"),
	}, nil
}
