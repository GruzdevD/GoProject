package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/segmentio/kafka-go"
)

// Config представляет структуру конфигурации Kafka
type Config struct {
	Brokers  []string `json:"brokers"`
	GroupID  string   `json:"group_id"`
	Topic    string   `json:"topic"`
	MinBytes int      `json:"min_bytes"`
	MaxBytes int      `json:"max_bytes"`
}

// Consumer представляет Kafka-консьюмера
type Consumer struct {
	reader *kafka.Reader
}

// LoadConfig загружает конфигурацию из JSON-файла
func LoadConfig(filePath string) (*Config, error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("could not unmarshal config: %w", err)
	}

	return &config, nil
}

// NewConsumer создает новый Kafka-консьюмера
func NewConsumer(config *Config) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.Brokers,
		Topic:    config.Topic,
		GroupID:  config.GroupID,
		MinBytes: config.MinBytes, // Минимальный размер сообщения
		MaxBytes: config.MaxBytes, // Максимальный размер сообщения
	})

	return &Consumer{
		reader: r,
	}
}

// Consume читает сообщения из Kafka и обрабатывает их
func (c *Consumer) Consume(ctx context.Context) {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("could not read message: %v", err)
			continue
		}

		log.Printf("Message received: Topic=%s, Partition=%d, Offset=%d, Value=%s",
			m.Topic, m.Partition, m.Offset, string(m.Value))
	}
}

// Close закрывает Kafka-консьюмера
func (c *Consumer) Close() error {
	return c.reader.Close()
}
