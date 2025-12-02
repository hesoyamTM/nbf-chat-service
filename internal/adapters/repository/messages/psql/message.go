// Package psql provides postgres adapters for messages.
package psql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"github.com/hesoyamTM/nbf-chat-service/internal/adapters/repository"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/chat"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/message"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/user"
	"github.com/hesoyamTM/nbf-chat-service/internal/migrations"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type PostgresMessageRepository struct {
	db *sqlx.DB
}

func NewPostgresMessageRepository(ctx context.Context, cfg PostgresMessageConfig) (*PostgresMessageRepository, error) {
	const op = "repository.NewPostgresMessageRepository"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Error("Failed to connect to database", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := migrations.Migrate(ctx, db.DB); err != nil {
		log.Error("Failed to migrate database", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &PostgresMessageRepository{
		db: db,
	}, nil
}

func (r *PostgresMessageRepository) Save(ctx context.Context, message message.Message) error {
	const op = "repository.PostgresMessageRepository.Save"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	query := `INSERT INTO messages (id, user_id, chat_id, content, content_type, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = r.db.ExecContext(ctx, query,
		message.ID,
		message.User.ID,
		message.ChatID,
		message.Text,
		"text",
		message.CreatedAt,
	)
	if err != nil {
		log.Error("Failed to save message", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *PostgresMessageRepository) GetAll(ctx context.Context, chatID uuid.UUID) ([]message.Message, error) {
	const op = "repository.PostgresMessageRepository.GetAll"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	query := `SELECT user_id, chat_id, content, content_type, created_at
	FROM messages
	WHERE chat_id = $1`

	rows, err := r.db.QueryContext(ctx, query, chatID)
	if err != nil {
		log.Error("Failed to get messages", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	messages := make([]message.Message, 0)

	for rows.Next() {
		var userID, chatID uuid.NullUUID
		var content, contentType string
		var createdAt time.Time

		err := rows.Scan(&userID, &chatID, &content, &contentType, &createdAt)
		if err != nil {
			log.Error("Failed to scan row", zap.Error(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if !userID.Valid || !chatID.Valid {
			log.Error("Invalid user or chat id",
				zap.String("user_id", userID.UUID.String()),
				zap.String("chat_id", chatID.UUID.String()),
			)
			// TODO: handle this case
		}

		messages = append(messages, message.Message{
			ID:        uuid.New(),
			User:      user.User{ID: userID.UUID},
			ChatID:    chatID.UUID,
			Text:      content,
			CreatedAt: createdAt,
		})
	}

	return messages, nil
}

func (r *PostgresMessageRepository) CreateNewChatByUser(ctx context.Context, chatID, userID, groupID uuid.UUID, chatName string) (uuid.UUID, error) {
	const op = "repository.PostgresMessageRepository.CreateNewChatByUser"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("create new chat by user")

	query := `INSERT INTO chats (id, user_id, group_id, chat_name, created_at, last_read_at, unread_count)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = r.db.ExecContext(ctx, query, chatID, userID, groupID, chatName, time.Now(), time.Now(), 0)
	if err != nil {
		log.Error("Failed to create chat", zap.Error(err))
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return chatID, nil
}

func (r *PostgresMessageRepository) GetChatIDByUser(ctx context.Context, senderID, userID uuid.UUID) (uuid.UUID, error) {
	const op = "repository.PostgresMessageRepository.GetChatByUser"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	query := `SELECT c.id
	FROM chats
	JOIN chats AS c ON c.id = chats.id
	WHERE c.user_id = $1 AND chats.user_id = $2`

	row := r.db.QueryRowContext(ctx, query, senderID, userID)

	var chatID uuid.NullUUID
	err = row.Scan(&chatID)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, repository.ErrNotFound
		}

		log.Error("Failed to get chat", zap.Error(err))
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	if !chatID.Valid {
		log.Error("Invalid chat id",
			zap.String("chat_id", chatID.UUID.String()),
		)
		// TODO: handle this case
	}

	return chatID.UUID, nil
}

func (r *PostgresMessageRepository) GetChatIDByGroup(ctx context.Context, senderID, groupID uuid.UUID) (uuid.UUID, error) {
	const op = "repository.PostgresMessageRepository.GetChatByGroup"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	query := `SELECT id
	FROM chats
	WHERE group_id = $1 AND user_id = $2`

	row := r.db.QueryRowContext(ctx, query, groupID, senderID)

	var chatID uuid.NullUUID
	err = row.Scan(&chatID)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, repository.ErrNotFound
		}

		log.Error("Failed to get chat", zap.Error(err))
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	if !chatID.Valid {
		log.Error("Invalid chat id",
			zap.String("chat_id", chatID.UUID.String()),
		)
		// TODO: handle this case
	}

	return chatID.UUID, nil
}

func (r *PostgresMessageRepository) GetChatByID(ctx context.Context, senderID, chatID uuid.UUID) (chat.Chat, error) {
	const op = "repository.PostgresMessageRepository.GetChatByID"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return chat.Chat{}, fmt.Errorf("%s: %w", op, err)
	}

	query := `SELECT id, chat_name
	FROM chats
	WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, chatID)

	var (
		id   uuid.NullUUID
		name string
	)
	err = row.Scan(&id, &name)
	if err != nil {
		if err == sql.ErrNoRows {
			return chat.Chat{}, repository.ErrNotFound
		}

		log.Error("Failed to get chat", zap.Error(err))
		return chat.Chat{}, fmt.Errorf("%s: %w", op, err)
	}

	if !id.Valid {
		log.Error("Invalid chat id",
			zap.String("chat_id", id.UUID.String()),
		)
		return chat.Chat{}, fmt.Errorf("%s: Invalid chat id", op)
	}

	membersQuery := `SELECT user_id
	FROM chats
	WHERE id = $1`

	membersRows, err := r.db.QueryContext(ctx, membersQuery, id.UUID)
	if err != nil {
		log.Error("Failed to get chat members", zap.Error(err))
		return chat.Chat{}, fmt.Errorf("%s: %w", op, err)
	}
	defer membersRows.Close()

	members := make(map[uuid.UUID]user.User, 0)

	for membersRows.Next() {
		var userID uuid.NullUUID

		err := membersRows.Scan(&userID)
		if err != nil {
			log.Error("Failed to scan row", zap.Error(err))
			return chat.Chat{}, fmt.Errorf("%s: %w", op, err)
		}

		if !userID.Valid {
			log.Error("Invalid user id",
				zap.String("user_id", userID.UUID.String()),
			)
			// TODO: handle this case
			continue
		}

		members[userID.UUID] = user.User{ID: userID.UUID}
	}

	return chat.Chat{ID: id.UUID, Name: name, Members: members}, nil
}

func (r *PostgresMessageRepository) GetChatByGroup(ctx context.Context, senderID, groupID uuid.UUID) (chat.Chat, error) {
	const op = "repository.PostgresMessageRepository.GetChatByGroup"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return chat.Chat{}, fmt.Errorf("%s: %w", op, err)
	}

	query := `SELECT id
	FROM chats
	WHERE group_id = $1 AND user_id = $2`

	row := r.db.QueryRowContext(ctx, query, groupID, senderID)

	var chatID uuid.NullUUID
	err = row.Scan(&chatID)
	if err != nil {
		if err == sql.ErrNoRows {
			return chat.Chat{}, repository.ErrNotFound
		}

		log.Error("Failed to get chat", zap.Error(err))
		return chat.Chat{}, fmt.Errorf("%s: %w", op, err)
	}

	if !chatID.Valid {
		log.Error("Invalid chat id",
			zap.String("chat_id", chatID.UUID.String()),
		)
		// TODO: handle this case
		return chat.Chat{}, fmt.Errorf("%s: Invalid chat id", op)
	}

	return r.GetChatByID(ctx, senderID, chatID.UUID)
}

func (r *PostgresMessageRepository) GetChatByUser(ctx context.Context, senderID, userID uuid.UUID) (chat.Chat, error) {
	const op = "repository.PostgresMessageRepository.GetChatByUser"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return chat.Chat{}, fmt.Errorf("%s: %w", op, err)
	}

	query := `SELECT id
	FROM chats
	WHERE user_id = $1 AND group_id IS NULL`

	row := r.db.QueryRowContext(ctx, query, userID)

	var chatID uuid.NullUUID
	err = row.Scan(&chatID)
	if err != nil {
		if err == sql.ErrNoRows {
			return chat.Chat{}, repository.ErrNotFound
		}

		log.Error("Failed to get chat", zap.Error(err))
		return chat.Chat{}, fmt.Errorf("%s: %w", op, err)
	}

	if !chatID.Valid {
		log.Error("Invalid chat id",
			zap.String("chat_id", chatID.UUID.String()),
		)
		// TODO: handle this case
		return chat.Chat{}, fmt.Errorf("%s: Invalid chat id", op)
	}

	return r.GetChatByID(ctx, senderID, chatID.UUID)
}

func (r *PostgresMessageRepository) GetChatsByUser(ctx context.Context, userID uuid.UUID) ([]chat.ChatDialog, error) {
	const op = "repository.PostgresMessageRepository.GetChatsByUser"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	query := `SELECT c.id, c.chat_name, c.group_id
	FROM chats AS c
	WHERE c.user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Error("Failed to get chats by user", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	chats := make([]chat.ChatDialog, 0)

	for rows.Next() {
		var (
			id              uuid.NullUUID
			name            string
			userID, groupID uuid.NullUUID
		)

		err := rows.Scan(&id, &name, &groupID)
		if err != nil {
			log.Error("Failed to scan row", zap.Error(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if !id.Valid {
			log.Error("Invalid chat id",
				zap.String("chat_id", id.UUID.String()),
				zap.String("user_id", userID.UUID.String()),
			)
			// TODO: handle this case
			continue
		}

		chats = append(chats, chat.ChatDialog{
			ID:   id.UUID,
			Name: name,
		})
	}

	return chats, nil
}

func (r *PostgresMessageRepository) IncrementUnreadCount(ctx context.Context, userID, chatID uuid.UUID, receiverID uuid.UUID) error {
	const op = "repository.PostgresMessageRepository.IncrementUnreadCount"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	query := `UPDATE chats
	SET unread_count = unread_count + 1
	WHERE id = $1 AND user_id = $2`

	_, err = r.db.ExecContext(ctx, query, chatID, receiverID)
	if err != nil {
		log.Error("Failed to increment unread count", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *PostgresMessageRepository) SetLastReadAt(ctx context.Context, userID, chatID uuid.UUID, lastReadAt time.Time) error {
	const op = "repository.PostgresMessageRepository.SetLastReadAt"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	query := `UPDATE chats
	SET last_read_at = $1
	WHERE id = $2 AND user_id = $3`

	_, err = r.db.ExecContext(ctx, query, lastReadAt, chatID, userID)
	if err != nil {
		log.Error("Failed to set last read at", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
