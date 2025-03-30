package main

import (
	"context"
	"kafka-to-api-service/internal/kafka"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Загружаем конфигурацию
	// configPath := filepath.Join("c:", "GoProject", "internal", "config", "config.json")
	configPath := "c:/GoProject/internal/config/config.json"
	config, err := kafka.LoadConfig(configPath) // Используем прямые слэши
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаём консьюмера
	consumer := kafka.NewConsumer(config)
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("Failed to close consumer: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка сигналов для корректного завершения
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signals
		cancel()
	}()

	log.Println("Starting consumer...")
	consumer.Consume(ctx)
}
