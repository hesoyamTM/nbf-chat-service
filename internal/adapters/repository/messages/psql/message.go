// Package psql provides postgres adapters for messages.
package psql

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"github.com/hesoyamTM/nbf-chat-service/internal/domain/message"
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

	log.Info("Save message", zap.String("message_id", message.ID.String()))

	query := `INSERT INTO messages (id, user_id, group_id, message, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err = r.db.ExecContext(
		ctx,
		query,
		message.ID,
		message.User.ID,
		message.GroupID,
		message.Text,
		message.CreatedAt,
	)
	if err != nil {
		log.Error("Failed to save message", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Save message", zap.String("message_id", message.ID.String()))

	return nil
}

func (r *PostgresMessageRepository) GetAll(ctx context.Context, groupID uuid.UUID) ([]message.Message, error) {
	const op = "repository.PostgresMessageRepository.GetAll"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Get all messages for group", zap.String("group_id", groupID.String()))

	query := `SELECT id, user_id, group_id, message, created_at FROM messages WHERE group_id = $1`
	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		log.Error("Failed to get messages", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	messages := make([]message.Message, 0)
	for rows.Next() {
		var message message.Message
		err := rows.Scan(
			&message.ID,
			&message.User.ID,
			&message.GroupID,
			&message.Text,
			&message.CreatedAt,
		)
		if err != nil {
			log.Error("Failed to scan message", zap.Error(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		messages = append(messages, message)
	}

	log.Info("Get all messages for group",
		zap.String("group_id", groupID.String()),
		zap.Int("count", len(messages)),
	)

	return messages, nil
}
