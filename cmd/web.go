package main

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	isConsumerRunning bool
	mu                sync.Mutex
)

func main() {
	// Создаем новый роутер Gin
	router := gin.Default()

	// Главная страница
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Добро пожаловать в Kafka-to-API Service!",
		})
	})

	// Эндпоинт для запуска Kafka-консьюмера
	router.POST("/start-consumer", func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()

		if isConsumerRunning {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Kafka consumer уже запущен",
			})
			return
		}

		// Здесь вы можете добавить логику для запуска Kafka-консьюмера
		isConsumerRunning = true
		c.JSON(http.StatusOK, gin.H{
			"status": "Kafka consumer запущен",
		})
	})

	// Эндпоинт для остановки Kafka-консьюмера
	router.POST("/stop-consumer", func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()

		if !isConsumerRunning {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Kafka consumer не запущен",
			})
			return
		}

		// Здесь вы можете добавить логику для остановки Kafka-консьюмера
		isConsumerRunning = false
		c.JSON(http.StatusOK, gin.H{
			"status": "Kafka consumer остановлен",
		})
	})

	// Эндпоинт для проверки статуса Kafka-консьюмера
	router.GET("/status", func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()

		status := "остановлен"
		if isConsumerRunning {
			status = "запущен"
		}

		c.JSON(http.StatusOK, gin.H{
			"consumer_status": status,
		})
	})

	// Запуск веб-сервера
	router.Run(":8080") // Сервер будет доступен по адресу http://localhost:8080
}
