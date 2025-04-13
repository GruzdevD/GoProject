package kafka

import (
	"bytes"         // Пакет для работы с буферами.
	"context"       // Пакет для работы с контекстами, которые помогают управлять временем жизни операций.
	"encoding/json" // Пакет для работы с JSON (кодирование и декодирование).
	"fmt"           // Пакет для форматирования строк.
	"io/ioutil"     // Пакет для работы с файлами (устаревший, но используется в этом коде).
	"log"           // Пакет для логирования сообщений.
	"net/http"      // Пакет для отправки HTTP-запросов.
	"os"            // Пакет для работы с файловой системой и операционной системой.

	"github.com/segmentio/kafka-go" // Библиотека для работы с Kafka.
)

// Config представляет структуру конфигурации Kafka.
type Config struct {
	Brokers      []string `json:"brokers"`       // Список брокеров Kafka.
	GroupID      string   `json:"group_id"`      // Группа консьюмера.
	Topic        string   `json:"topic"`         // Топик Kafka для чтения сообщений.
	MinBytes     int      `json:"min_bytes"`     // Минимальный размер данных для чтения.
	MaxBytes     int      `json:"max_bytes"`     // Максимальный размер данных для чтения.
	MessageLimit int      `json:"message_limit"` // Лимит сообщений для отправки.
	PostURL      string   `json:"post_url"`      // URL для отправки сообщений через POST-запрос.
	PortApp      string   `json:"portApp"`       // порт публикации приложения
}

// Consumer представляет Kafka-консьюмера.
type Consumer struct {
	reader  *kafka.Reader // Kafka Reader для чтения сообщений.
	limit   int           // Лимит сообщений перед отправкой.
	postURL string        // URL для отправки сообщений.
	file    *os.File      // Файл для временного хранения сообщений.
}

// LoadConfig загружает конфигурацию из JSON-файла.
func LoadConfig(filePath string) (*Config, error) {
	// Читаем содержимое файла конфигурации.
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var config Config
	// Декодируем JSON в структуру Config.
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("could not unmarshal config: %w", err)
	}

	return &config, nil
}

// NewConsumer создает новый Kafka-консьюмера.
func NewConsumer(config *Config) *Consumer {
	// Создаем Kafka Reader с параметрами из конфигурации.
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.Brokers,
		Topic:    config.Topic,
		GroupID:  config.GroupID,
		MinBytes: config.MinBytes,
		MaxBytes: config.MaxBytes,
	})

	// Создаем файл для временного хранения сообщений.
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

// sendMessages отправляет сообщения POST-запросом.
func (c *Consumer) sendMessages(ctx context.Context) error {
	// Читаем содержимое файла с сообщениями.
	data, err := os.ReadFile(c.file.Name())
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Создаем HTTP-запрос с данными из файла.
	req, err := http.NewRequestWithContext(ctx, "POST", c.postURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Отправляем запрос.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа.
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("received non-OK response: %d, body: %s", resp.StatusCode, string(body))
	}

	// Очищаем файл после успешной отправки.
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

// logToChannel логирует сообщение в канал (заглушка для демонстрации).
func logToChannel(message string) {
	// Здесь можно реализовать отправку сообщения в канал.
	log.Println("Log to channel:", message)
}

// Consume читает сообщения из Kafka и обрабатывает их.
func (c *Consumer) Consume(ctx context.Context) {
	messageCount := 0 // Счетчик сообщений.

	for {
		// Проверяем, был ли контекст отменён.
		select {
		case <-ctx.Done():
			log.Printf("Context canceled before reading message (%v), stopping consumer...", ctx.Err())
			logToChannel("Consumer stopped due to context cancellation.")
			return
		default:
			// Контекст не отменён, продолжаем выполнение цикла.
		}

		// Читаем сообщение из Kafka.
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("could not read message: %v", err)
			if ctx.Err() != nil {
				log.Println("Context canceled, stopping consumer...")
				logToChannel("Consumer stopped due to context cancellation.")
				return
			}
			if err == kafka.ErrGroupClosed {
				log.Println("Kafka group closed, stopping consumer...")
				logToChannel("Consumer stopped: Kafka group closed.")
				return
			}
			log.Println("Critical error, stopping consumer...")
			logToChannel("Consumer stopped due to critical error.")
			return
		}

		// Логируем полученное сообщение.
		log.Printf("Message received: Topic=%s, Partition=%d, Offset=%d, Value=%s",
			m.Topic, m.Partition, m.Offset, string(m.Value))

		// Сохраняем сообщение в файл.
		if _, err := c.file.WriteString(string(m.Value) + "\n"); err != nil {
			log.Printf("Failed to write message to file: %v", err)
			continue
		}

		messageCount++ // Увеличиваем счётчик сообщений.

		// Если достигнут лимит, отправляем сообщения.
		if messageCount >= c.limit {
			log.Println("Message limit reached, sending messages...")
			if err := c.sendMessages(ctx); err != nil {
				log.Printf("Failed to send messages: %v", err)
			} else {
				messageCount = 0 // Сбрасываем счётчик после успешной отправки.
			}
		}
	}
}

// Close закрывает Kafka-консьюмера и файл.
func (c *Consumer) Close() error {
	var errs []error // Список ошибок.

	// Закрываем файл.
	if err := c.file.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close file: %w", err))
	}
	// Закрываем Kafka Reader.
	if err := c.reader.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close Kafka reader: %w", err))
	}

	// Если есть ошибки, возвращаем их.
	if len(errs) > 0 {
		return fmt.Errorf("multiple errors: %v", errs)
	}
	return nil
}
