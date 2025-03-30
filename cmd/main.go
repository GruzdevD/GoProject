package main

import (
	"context"
	"encoding/json"
	"kafka-to-api-service/internal/kafka"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// App представляет приложение с Kafka-консьюмером и конфигурацией
type App struct {
	Consumer *kafka.Consumer
	Config   *kafka.Config
	Ctx      context.Context
	Cancel   context.CancelFunc
	Running  bool
	mu       sync.Mutex // Для управления состоянием приложения
}

func (app *App) startConsumerHandler(w http.ResponseWriter, r *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	if app.Running {
		http.Error(w, "Consumer is already running", http.StatusBadRequest)
		return
	}

	app.Ctx, app.Cancel = context.WithCancel(context.Background())
	go app.Consumer.Consume(app.Ctx)
	app.Running = true

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Consumer started"))
}

func (app *App) stopConsumerHandler(w http.ResponseWriter, r *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	if !app.Running {
		http.Error(w, "Consumer is not running", http.StatusBadRequest)
		return
	}

	app.Cancel()
	app.Running = false

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Consumer stopped"))
}

func (app *App) getConfigHandler(w http.ResponseWriter, r *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app.Config)
}

func (app *App) updateConfigHandler(w http.ResponseWriter, r *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	var newConfig kafka.Config
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, "Invalid config format", http.StatusBadRequest)
		return
	}

	// Сохраняем изменения в config.json
	configPath := "c:/GoProject/internal/config/config.json"
	file, err := os.Create(configPath)
	if err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(newConfig); err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	// Обновляем конфигурацию в приложении
	app.Config = &newConfig
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Config updated"))
}

func main() {
	// Загружаем конфигурацию
	configPath := "c:/GoProject/internal/config/config.json"
	config, err := kafka.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаём консьюмера
	consumer := kafka.NewConsumer(config)

	app := &App{
		Consumer: consumer,
		Config:   config,
		Running:  false,
	}

	// HTTP-обработчики
	http.HandleFunc("/start", app.startConsumerHandler)
	http.HandleFunc("/stop", app.stopConsumerHandler)
	http.HandleFunc("/config", app.getConfigHandler)
	http.HandleFunc("/config/update", app.updateConfigHandler)

	// Обслуживание статических файлов (HTML)
	webDir := "c:/GoProject/web"
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	// Запуск HTTP-сервера
	go func() {
		log.Println("Starting HTTP server on :8080...")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Обработка сигналов для корректного завершения
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	<-signals
	log.Println("Shutting down...")
	if app.Running {
		app.Cancel()
	}
	if err := consumer.Close(); err != nil {
		log.Printf("Failed to close consumer: %v", err)
	}
}
