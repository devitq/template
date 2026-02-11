package kafka

import (
	"github.com/segmentio/kafka-go"
)

type KafkaConfig struct {
	Addr          string
	ReaderGroupID string
	Topic         string
}

func NewKafkaWriter(addr string) *kafka.Writer {
	return &kafka.Writer{
		Addr:                   kafka.TCP(addr),
		AllowAutoTopicCreation: true,
	}
}

func NewKafkaReader(cfg KafkaConfig) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{cfg.Addr},
		GroupID:     cfg.ReaderGroupID,
		GroupTopics: []string{cfg.Topic},
	})
}
