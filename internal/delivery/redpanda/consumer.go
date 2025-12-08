// Package redpanda provides the implementation of the redpanda consumer.
package redpanda

import (
	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type Consumer struct {
	chatService   ChatService
	topicHandlers map[string]TopicHandler
	log           *zap.Logger
	Ready         chan struct{}
}

func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	close(c.Ready)
	c.log.Info("consumer group setup")
	return nil
}

func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	c.log.Info("consumer group cleanup")
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	c.topicHandlers = map[string]TopicHandler{
		"matcher.group.accept": c.HandleMatcherGroupRequestAccepted,
		"matcher.group.leave":  c.HandleMatcherGroupUserLeft,
		"matcher.group.kick":   c.HandleMatcherGroupUserKicked,
	}

	for message := range claim.Messages() {
		c.log.Info("message received",
			zap.String("topic", message.Topic),
			zap.Int32("partition", message.Partition),
			zap.Int64("offset", message.Offset),
		)

		topicHandler, ok := c.topicHandlers[message.Topic]
		if !ok {
			c.log.Error("no handler for topic",
				zap.String("topic", message.Topic),
			)
			continue
		}

		if err := topicHandler(session.Context(), c.log, message); err != nil {
			c.log.Error("failed to handle message",
				zap.String("topic", message.Topic),
				zap.Error(err),
			)
		}

		session.MarkMessage(message, "")
	}
	return nil
}
