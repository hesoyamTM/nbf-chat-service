// Package group contains the domain model for groups.
package group

import (
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/user"

	"github.com/google/uuid"
)

type Group struct {
	ID      uuid.UUID
	Name    string
	Members map[uuid.UUID]user.User
}

func NewGroup(id uuid.UUID, name string, members []user.User) Group {
	membersMap := make(map[uuid.UUID]user.User)

	for _, member := range members {
		membersMap[member.ID] = member
	}
	return Group{
		ID:      id,
		Name:    name,
		Members: membersMap,
	}
}

func (g *Group) AddMember(userID uuid.UUID, user user.User) {
	g.Members[userID] = user
}

func (g *Group) GetMember(userID uuid.UUID) (user.User, bool) {
	user, ok := g.Members[userID]
	return user, ok
}

func (g *Group) RemoveMember(userID uuid.UUID) {
	delete(g.Members, userID)
}
