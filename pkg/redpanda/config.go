package redpanda

type RedpandaConfig struct {
	Brokers []string `env:"REDPANDA_BROKERS" env-required:"true" yaml:"brokers"`
	GroupID string   `env:"REDPANDA_GROUP_ID" env-required:"true" yaml:"group_id"`
	Topics  []string `env:"REDPANDA_TOPICS" env-required:"true" yaml:"topics"`
}
