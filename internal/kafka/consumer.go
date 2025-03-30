package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/segmentio/kafka-go"
)

// Config представляет структуру конфигурации Kafka
type Config struct {
	Brokers      []string `json:"brokers"`
	GroupID      string   `json:"group_id"`
	Topic        string   `json:"topic"`
	MinBytes     int      `json:"min_bytes"`
	MaxBytes     int      `json:"max_bytes"`
	MessageLimit int      `json:"message_limit"` // Лимит сообщений
	PostURL      string   `json:"post_url"`      // URL для POST-запроса
}

// Consumer представляет Kafka-консьюмера
type Consumer struct {
	reader  *kafka.Reader
	limit   int
	postURL string
	file    *os.File
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
		MinBytes: config.MinBytes,
		MaxBytes: config.MaxBytes,
	})

	// Создаем файл для хранения сообщений
	filePath := fmt.Sprintf("%s_messages.txt", config.Topic)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to create file for topic %s: %v", config.Topic, err)
	}

	return &Consumer{
		reader:  r,
		limit:   config.MessageLimit,
		postURL: config.PostURL,
		file:    file,
	}
}

// sendMessages отправляет сообщения POST-запросом
func (c *Consumer) sendMessages(ctx context.Context) error {
	data, err := os.ReadFile(c.file.Name())
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.postURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("received non-OK response: %d, body: %s", resp.StatusCode, string(body))
	}

	// Очищаем файл после успешной отправки
	if err := c.file.Close(); err != nil {
		return fmt.Errorf("failed to close file before truncating: %w", err)
	}
	if err := os.Truncate(c.file.Name(), 0); err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}
	file, err := os.OpenFile(c.file.Name(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to reopen file: %w", err)
	}
	c.file = file

	log.Printf("Messages successfully sent and file cleared: %s", c.file.Name())
	return nil
}

// Consume читает сообщения из Kafka и обрабатывает их
func (c *Consumer) Consume(ctx context.Context) {
	messageCount := 0

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("could not read message: %v", err)
			if ctx.Err() != nil {
				log.Println("Context canceled, stopping consumer...")
				return
			}
			continue
		}

		log.Printf("Message received: Topic=%s, Partition=%d, Offset=%d, Value=%s",
			m.Topic, m.Partition, m.Offset, string(m.Value))

		// Сохраняем сообщение в файл
		if _, err := c.file.WriteString(string(m.Value) + "\n"); err != nil {
			log.Printf("Failed to write message to file: %v", err)
			continue
		}

		messageCount++

		// Если достигнут лимит, отправляем сообщения
		if messageCount >= c.limit {
			log.Println("Message limit reached, sending messages...")
			if err := c.sendMessages(ctx); err != nil {
				log.Printf("Failed to send messages: %v", err)
			} else {
				messageCount = 0
			}
		}
	}
}

// Close закрывает Kafka-консьюмера и файл
func (c *Consumer) Close() error {
	var errs []error

	if err := c.file.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close file: %w", err))
	}
	if err := c.reader.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close Kafka reader: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("multiple errors: %v", errs)
	}
	return nil
}
