package kafka

import (
    "context"
    "encoding/json"
    "github.com/segmentio/kafka-go"
    "log"
    "time"
)

type Consumer struct {
    reader *kafka.Reader
    apiClient *APIClient
}

func NewConsumer(brokers []string, topic string, apiClient *APIClient) *Consumer {
    r := kafka.NewReader(kafka.ReaderConfig{
        Brokers:  brokers,
        Topic:    topic,
        GroupID:  "my-group",
        MinBytes: 10e3, // 10KB
        MaxBytes: 10e6, // 10MB
    })

    return &Consumer{
        reader:   r,
        apiClient: apiClient,
    }
}

func (c *Consumer) Consume(ctx context.Context) {
    for {
        m, err := c.reader.ReadMessage(ctx)
        if err != nil {
            log.Printf("could not read message: %v", err)
            continue
        }

        var messageData map[string]interface{}
        if err := json.Unmarshal(m.Value, &messageData); err != nil {
            log.Printf("could not unmarshal message: %v", err)
            continue
        }

        if err := c.apiClient.SendMessage(ctx, messageData); err != nil {
            log.Printf("could not send message to API: %v", err)
        }
    }
}

func (c *Consumer) Close() error {
    return c.reader.Close()
}