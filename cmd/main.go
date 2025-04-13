package main

import (
	"context"       // Пакет для работы с контекстами, которые помогают управлять временем жизни операций.
	"encoding/json" // Пакет для работы с JSON (кодирование и декодирование).
	"fmt"
	"kafka-to-api-service/internal/kafka" // Импорт внутреннего пакета для работы с Kafka.
	"log"                                 // Пакет для логирования сообщений.
	"net/http"                            // Пакет для создания HTTP-сервера.
	"os"                                  // Пакет для работы с файловой системой и операционной системой.
	"os/signal"                           // Пакет для обработки системных сигналов (например, Ctrl+C).
	"sync"                                // Пакет для работы с мьютексами (синхронизация потоков).
	"syscall"                             // Пакет для работы с системными вызовами (например, завершение программы).

	"github.com/gorilla/websocket" // Библиотека для работы с WebSocket-соединениями.
)

// Структура App представляет приложение с Kafka-консьюмером и конфигурацией.
type App struct {
	Consumer *kafka.Consumer    // Kafka-консьюмер для чтения сообщений.
	Config   *kafka.Config      // Конфигурация приложения.
	Ctx      context.Context    // Контекст для управления временем жизни операций.
	Cancel   context.CancelFunc // Функция для отмены контекста.
	Running  bool               // Флаг, указывающий, запущен ли консьюмер.
	mu       sync.Mutex         // Мьютекс для синхронизации доступа к состоянию приложения.
}

// Обработчик для запуска Kafka-консьюмера.
func (app *App) startConsumerHandler(w http.ResponseWriter, r *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	// Загружаем актуальную конфигурацию из файла.
	configPath := "c:/GoProject/internal/config/config.json"
	newConfig, err := kafka.LoadConfig(configPath)
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		logToChannel("Failed to load config: " + err.Error())
		return
	}

	// Добавляем логирование сразу после загрузки новой конфигурации.
	logToChannel(fmt.Sprintf("Loaded New Config - Topic: %s, GroupID: %s, Brokers: %v", newConfig.Topic, newConfig.GroupID, newConfig.Brokers))

	// Добавляем логирование текущей конфигурации перед сравнением.
	logToChannel(fmt.Sprintf("Current App Config - Topic: %s, GroupID: %s, Brokers: %v", app.Config.Topic, app.Config.GroupID, app.Config.Brokers))

	// Проверяем, изменились ли параметры (topic, group_id, brokers).
	if app.Config.Topic != newConfig.Topic || app.Config.GroupID != newConfig.GroupID || !equalStringSlices(app.Config.Brokers, newConfig.Brokers) {
		// Если параметры изменились, пересоздаём Kafka-консьюмера.
		logToChannel("Configuration changed. Recreating Kafka consumer...")
		if err := app.Consumer.Close(); err != nil {
			logToChannel("Failed to close existing consumer: " + err.Error())
		}

		// Создаём новый Kafka-консьюмер с обновлённой конфигурацией.
		app.Consumer = kafka.NewConsumer(newConfig)
		app.Config = newConfig // Обновляем конфигурацию в приложении.
		logToChannel("Kafka consumer recreated with new configuration.")
	} else {
		logToChannel("Configuration unchanged. Using existing Kafka consumer.")
	}

	// Если консьюмер уже запущен, останавливаем его перед запуском.
	if app.Running {
		app.Cancel()
		app.Running = false
	}

	// Создаём новый контекст для управления временем жизни консьюмера.
	app.Ctx, app.Cancel = context.WithCancel(context.Background())
	go app.Consumer.Consume(app.Ctx) // Запускаем консьюмер в отдельной горутине.
	app.Running = true

	logToChannel("Consumer started")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Consumer started"))
}

// Обработчик для остановки Kafka-консьюмера.
func (app *App) stopConsumerHandler(w http.ResponseWriter, r *http.Request) {
	app.mu.Lock()         // Блокируем доступ к состоянию приложения.
	defer app.mu.Unlock() // Разблокируем после завершения функции.

	// Если консьюмер не запущен, возвращаем ошибку.
	if !app.Running {
		http.Error(w, "Consumer is not running", http.StatusBadRequest)
		return
	}

	// Отменяем контекст, чтобы остановить консьюмер.
	app.Cancel()        // Отменяем текущий контекст.
	app.Running = false // Устанавливаем флаг, что консьюмер остановлен.

	// Возвращаем успешный ответ.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Consumer stopped"))
}

// Обработчик для получения текущей конфигурации.
func (app *App) getConfigHandler(w http.ResponseWriter, r *http.Request) {
	app.mu.Lock()         // Блокируем доступ к состоянию приложения.
	defer app.mu.Unlock() // Разблокируем после завершения функции.

	// Устанавливаем заголовок ответа как JSON.
	w.Header().Set("Content-Type", "application/json")
	// Кодируем текущую конфигурацию в JSON и отправляем в ответ.
	json.NewEncoder(w).Encode(app.Config)
}

// Обработчик для обновления конфигурации.
func (app *App) updateConfigHandler(w http.ResponseWriter, r *http.Request) {
	app.mu.Lock()         // Блокируем доступ к состоянию приложения.
	defer app.mu.Unlock() // Разблокируем после завершения функции.

	var newConfig kafka.Config // Создаем переменную для новой конфигурации.
	// Декодируем JSON из тела запроса в структуру newConfig.
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, "Invalid config format", http.StatusBadRequest)
		return
	}

	// Сохраняем новую конфигурацию в файл config.json.
	configPath := "c:/GoProject/internal/config/config.json"
	file, err := os.Create(configPath) // Создаем (или перезаписываем) файл.
	if err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}
	defer file.Close() // Закрываем файл после завершения функции.

	// Кодируем новую конфигурацию в JSON и записываем в файл.
	if err := json.NewEncoder(file).Encode(newConfig); err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	// Обновляем конфигурацию в приложении.
	//app.Config = &newConfig
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Config updated"))
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var logChannel = make(chan string, 100) // Канал для передачи логов

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil) // Обновляем соединение до WebSocket
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()

	log.Println("WebSocket connection established")

	// Отправляем логи через WebSocket
	for logMessage := range logChannel {
		err := conn.WriteMessage(websocket.TextMessage, []byte(logMessage))
		if err != nil {
			log.Println("Failed to send log message:", err)
			break
		}
	}
}

// Функция для отправки логов в канал.
func logToChannel(message string) {
	select {
	case logChannel <- message:
	default:
		log.Println("Log channel is full, dropping message:", message)
	}
}

// logWriterFunc позволяет перенаправить стандартный логгер в кастомную функцию.
type logWriterFunc func(p []byte) (n int, err error)

func (f logWriterFunc) Write(p []byte) (n int, err error) {
	message := string(p)
	logToChannel(message) // Отправляем сообщение в WebSocket через logToChannel
	return len(p), nil
}

// Функция для сравнения двух срезов строк.
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Главная функция приложения.
func main() {

	// Перенаправляем стандартный логгер в logToChannel
	log.SetOutput(logWriterFunc(func(p []byte) (n int, err error) {
		message := string(p)
		logToChannel(message) // Отправляем сообщение в WebSocket
		return len(p), nil
	}))

	// Загружаем конфигурацию из файла config.json.
	configPath := "c:/GoProject/internal/config/config.json"
	config, err := kafka.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err) // Завершаем приложение, если конфигурация не загружена.
	}

	// Создаем Kafka-консьюмера с загруженной конфигурацией.
	consumer := kafka.NewConsumer(config)

	// Создаем экземпляр приложения.
	app := &App{
		Consumer: consumer,
		Config:   config,
		Running:  false,
	}

	// Регистрируем HTTP-обработчики.
	http.HandleFunc("/api/start", app.startConsumerHandler)        // Запуск консьюмера.
	http.HandleFunc("/api/stop", app.stopConsumerHandler)          // Остановка консьюмера.
	http.HandleFunc("/api/config", app.getConfigHandler)           // Получение конфигурации.
	http.HandleFunc("/api/config/update", app.updateConfigHandler) // Обновление конфигурации.

	// Обслуживание статических файлов (например, HTML-страницы).
	webDir := "c:/GoProject/web"
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	http.HandleFunc("/ws", handleWebSocket)

	log.Println("Starting HTTP server on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))

	// Запуск HTTP-сервера в отдельной горутине.
	go func() {
		log.Println("Starting HTTP server on :" + config.PortApp)
		if err := http.ListenAndServe(":"+config.PortApp, nil); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Обработка системных сигналов (например, Ctrl+C).
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	<-signals                       // Ожидаем сигнал завершения.
	log.Println("Shutting down...") // Логируем завершение приложения.

	// Если консьюмер запущен, останавливаем его.
	if app.Running {
		app.Cancel()
	}
	// Закрываем Kafka-консьюмера.
	if err := consumer.Close(); err != nil {
		log.Printf("Failed to close consumer: %v", err)
	}

	// Закрытие WebSocket-соединений
	close(logChannel)
}
