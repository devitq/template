package kafka

import (
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaConfig struct {
	Addr          string
	ReaderGroupID string
	Topic         string

	// Writer settings
	Balancer               kafka.Balancer
	MaxAttempts            int
	WriteBatchSize         int
	WriteBatchTimeout      time.Duration
	WriteRequiredAcks      kafka.RequiredAcks
	AllowAutoTopicCreation bool

	// Reader settings
	ReadMinBytes       int
	ReadMaxBytes       int
	ReadCommitInterval time.Duration
}

func NewKafkaWriter(cfg KafkaConfig) *kafka.Writer {
	if cfg.Balancer == nil {
		cfg.Balancer = &kafka.LeastBytes{}
	}
	if cfg.WriteBatchSize == 0 {
		cfg.WriteBatchSize = 100
	}
	if cfg.WriteBatchTimeout == 0 {
		cfg.WriteBatchTimeout = 1 * time.Second
	}
	if cfg.WriteRequiredAcks == 0 {
		cfg.WriteRequiredAcks = kafka.RequireAll
	}

	return &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Addr),
		Balancer:               cfg.Balancer,
		MaxAttempts:            cfg.MaxAttempts,
		BatchSize:              cfg.WriteBatchSize,
		BatchTimeout:           cfg.WriteBatchTimeout,
		RequiredAcks:           cfg.WriteRequiredAcks,
		AllowAutoTopicCreation: cfg.AllowAutoTopicCreation,
	}
}

func NewKafkaReader(cfg KafkaConfig) *kafka.Reader {
	if cfg.ReadMinBytes == 0 {
		cfg.ReadMinBytes = 10e3 // 10KB
	}
	if cfg.ReadMaxBytes == 0 {
		cfg.ReadMaxBytes = 10e6 // 10MB
	}
	if cfg.ReadCommitInterval == 0 {
		cfg.ReadCommitInterval = 1 * time.Second
	}

	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{cfg.Addr},
		GroupID:        cfg.ReaderGroupID,
		GroupTopics:    []string{cfg.Topic},
		MinBytes:       cfg.ReadMinBytes,
		MaxBytes:       cfg.ReadMaxBytes,
		CommitInterval: cfg.ReadCommitInterval,
	})
}
