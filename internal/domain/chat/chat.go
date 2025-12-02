// Package chat contains the domain model for chats.
package chat

import (
	"time"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/group"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/user"
)

type ChatDialog struct {
	ID          uuid.UUID
	User        user.User
	Group       *group.Group
	Name        string
	LastReadAt  time.Time
	UnreadCount int
}

type Chat struct {
	ID      uuid.UUID
	Name    string
	Members map[uuid.UUID]user.User
}

func NewChatDialog(ID uuid.UUID, user user.User, group *group.Group, name string, lastReadAt time.Time, unreadCount int) ChatDialog {
	return ChatDialog{
		ID:          ID,
		User:        user,
		Group:       group,
		Name:        name,
		LastReadAt:  lastReadAt,
		UnreadCount: unreadCount,
	}
}
