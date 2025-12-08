package redpanda

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/IBM/sarama"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"github.com/hesoyamTM/nbf-chat-service/pkg/redpanda"
	"go.uber.org/zap"
)

type RedPandaApp struct {
	consumerGroup sarama.ConsumerGroup

	topics      []string
	chatService ChatService
	saramaCfg   *sarama.Config

	wg *sync.WaitGroup
}

func NewRedPandaApp(ctx context.Context, chatService ChatService, config redpanda.RedpandaConfig) *RedPandaApp {
	l, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		log.Fatal("failed to get logger from context")
	}

	l.Debug("redpanda host", zap.Any("host", config.Brokers))

	saramaCfg := redpanda.NewSaramaConfig(config)
	consumerGroup, err := redpanda.NewSaramaConsumer(saramaCfg,
		config.Brokers,
		config.GroupID,
	)
	if err != nil {
		l.Error("Error creating RedPanda producer", zap.Error(err))
	}

	return &RedPandaApp{
		consumerGroup: consumerGroup,

		topics:      config.Topics,
		chatService: chatService,

		wg: &sync.WaitGroup{},
	}
}

func (r *RedPandaApp) Run(ctx context.Context) error {
	l, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		log.Fatal("failed to get logger from context")
	}

	l.Info("Starting RedPanda consumer")

	consumer := &Consumer{
		chatService:   r.chatService,
		topicHandlers: map[string]TopicHandler{},
		log:           l,
		Ready:         make(chan struct{}),
	}

	r.wg.Go(func() {
		for {
			if err := r.consumerGroup.Consume(ctx, r.topics, consumer); err != nil {
				l.Error("Error consuming from RedPanda", zap.Error(err))
				return
			}
			if ctx.Err() != nil {
				return
			}

			consumer.Ready = make(chan struct{})
		}
	})

	<-consumer.Ready
	l.Info("Consumer group ready")

	return nil
}

func (r *RedPandaApp) Stop(ctx context.Context) error {
	const op = "redpanda.Stop"

	l, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		log.Fatal("failed to get logger from context")
	}

	if err := r.consumerGroup.Close(); err != nil {
		l.Error(op, zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	r.wg.Wait()
	return nil
}
