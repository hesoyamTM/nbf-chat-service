// Package user contains the domain model for users.
package user

import "github.com/google/uuid"

type User struct {
	ID   uuid.UUID
	Name string
}
