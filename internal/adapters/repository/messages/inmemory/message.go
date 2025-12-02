// Package inmemory provides in-memory adapters for messages.
package inmemory

//
// import (
// 	"context"
// 	"sync"
//
// 	"github.com/google/uuid"
// 	"github.com/hesoyamTM/nbf-chat-service/internal/domain/message"
// )
//
// type InMemoryMessageRepository struct {
// 	mutex    sync.RWMutex
// 	messages map[uuid.UUID][]message.Message
// }
//
// func NewInMemoryMessageRepository() *InMemoryMessageRepository {
// 	return &InMemoryMessageRepository{
// 		messages: make(map[uuid.UUID][]message.Message),
// 	}
// }
//
// func (r *InMemoryMessageRepository) Save(ctx context.Context, message message.Message) error {
// 	r.mutex.Lock()
// 	defer r.mutex.Unlock()
//
// 	r.messages[message.GroupID] = append(r.messages[message.GroupID], message)
//
// 	return nil
// }
//
// func (r *InMemoryMessageRepository) GetAll(ctx context.Context, groupID uuid.UUID) ([]message.Message, error) {
// 	r.mutex.RLock()
// 	defer r.mutex.RUnlock()
//
// 	return r.messages[groupID], nil
// }
