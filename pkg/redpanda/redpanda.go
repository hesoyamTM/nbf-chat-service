// Package redpanda is a wrapper for the redpanda client.
package redpanda

import (
	"github.com/IBM/sarama"
)

func NewSaramaConfig(cfg RedpandaConfig) *sarama.Config {
	config := sarama.NewConfig()

	config.Version = sarama.V2_8_0_0

	return config
}

func NewSaramaAsyncProducer(saramaCfg *sarama.Config, brokers []string) (*sarama.AsyncProducer, error) {
	producer, err := sarama.NewAsyncProducer(brokers, saramaCfg)
	if err != nil {
		return nil, err
	}

	return &producer, nil
}

func NewSaramaConsumer(saramaCfg *sarama.Config, brokers []string, groupID string) (sarama.ConsumerGroup, error) {
	consumer, err := sarama.NewConsumerGroup(brokers, groupID, saramaCfg)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}
