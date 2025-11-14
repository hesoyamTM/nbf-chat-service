// Package message contains the domain model for messages.
package message

import (
	"time"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/user"
)

type InputMessage struct {
	UserID  uuid.UUID
	GroupID uuid.UUID
	Text    string
}

type Message struct {
	ID      uuid.UUID
	User    user.User
	GroupID uuid.UUID
	Text    string

	CreatedAt time.Time
}

func NewMessage(user user.User, groupID uuid.UUID, text string) Message {
	return Message{
		ID:      uuid.New(),
		User:    user,
		GroupID: groupID,
		Text:    text,

		CreatedAt: time.Now(),
	}
}
