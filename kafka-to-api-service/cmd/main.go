package main

import (
    "log"
    "kafka-to-api-service/internal/api"
    "kafka-to-api-service/internal/config"
    "kafka-to-api-service/internal/kafka"
)

func main() {
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("could not load config: %v", err)
    }

    kafkaConsumer := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
    apiClient := api.NewAPIClient(cfg.API.Endpoint)

    if err := kafkaConsumer.Consume(apiClient); err != nil {
        log.Fatalf("could not consume messages: %v", err)
    }
}