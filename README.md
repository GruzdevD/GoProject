# Kafka to API Service

Этот проект представляет собой Go-сервис, который читает сообщения из Kafka-топика и отправляет их на указанный API-эндпоинт. Он разработан для простоты и эффективности, используя возможности конкурентности Go для обработки сообщений и выполнения API-запросов. Также предоставляется удобный веб-интерфейс для управления приложением.

---

## Структура проекта

```
kafka-to-api-service
├── cmd
│   └── main.go          # Точка входа в приложение
├── internal
│   ├── kafka
│   │   └── consumer.go  # Реализация Kafka-консьюмера
│   ├── api
│   │   └── client.go    # API-клиент для отправки сообщений
│   └── config
│       └── config.go    # Загрузка конфигурации
├── pkg
│   └── logger
│       └── logger.go     # Реализация логгера
├── web
│   ├── index.html       # Веб-интерфейс для управления приложением
│   ├── script.js        # Логика фронтенда
│   ├── styles.css       # Стили для веб-интерфейса
│   ├── icons            # SVG-иконки для интерфейса
│   └── site.webmanifest # Манифест для PWA
├── go.mod                # Файл модулей Go
├── go.sum                # Контрольные суммы модулей Go
└── README.md             # Документация проекта
```

---

## Установка

1.  Клонируйте репозиторий (если еще не сделано):
    ```bash
    git clone <your-repo-url>
    cd <your-repo-directory>
    ```
2.  Загрузите зависимости:
    ```bash
    go mod tidy
    # или
    go mod download
    ```

## Конфигурация

Приложение настраивается с помощью JSON-файла. Путь к файлу указывается при запуске с помощью флага `-config`. По умолчанию используется `./config.json`.

**Пример `config.json`:**

```json
{
    "brokers": ["localhost:9092"],
    "group_id": "my-group",
    "topic": "test-topic",
    "min_bytes": 10000,
    "max_bytes": 10000000,
    "message_limit": 10,
    "post_url": "http://localhost:8080/api/messages",
    "port": 8080
}
```

- **`brokers`**: Список брокеров Kafka.
- **`group_id`**: Группа консьюмера.
- **`topic`**: Топик Kafka для вычитки сообщений.
- **`min_bytes` и `max_bytes`**: Минимальный и максимальный размер данных для чтения.
- **`message_limit`**: Лимит сообщений для отправки.
- **`post_url`**: URL для отправки сообщений.
- **`port`**: Порт для запуска веб-сервера.

---

## Запуск приложения

Для запуска приложения выполните:

```bash
go run cmd/main.go
```

После запуска приложение будет доступно по адресу: [http://localhost:8080](http://localhost:8080).

---

## Веб-интерфейс

Приложение предоставляет удобный веб-интерфейс для управления Kafka-консьюмером и конфигурацией. Откройте [http://localhost:8080](http://localhost:8080) в браузере.

### Возможности веб-интерфейса:

1. **Управление консьюмером**:
   - **Запуск**: Нажмите кнопку "Запустить консьюмер", чтобы начать вычитку сообщений из Kafka.
   - **Остановка**: Нажмите кнопку "Остановить консьюмер", чтобы остановить вычитку.

2. **Редактирование конфигурации**:
   - Измените параметры, такие как `group_id`, `topic`, `brokers`, и нажмите "Обновить", чтобы сохранить изменения.

3. **Просмотр логов**:
   - Логи отображаются в реальном времени через WebSocket.

4. **Темная тема**:
   - Переключение между светлой и темной темой интерфейса.

---

## API (опционально)

Вы также можете взаимодействовать с приложением через API:

- **Запуск консьюмера**:
  ```bash
  curl -X POST http://localhost:8080/start
  ```

- **Остановка консьюмера**:
  ```bash
  curl -X POST http://localhost:8080/stop
  ```

- **Получение текущей конфигурации**:
  ```bash
  curl -X GET http://localhost:8080/config
  ```

- **Обновление конфигурации**:
  ```bash
  curl -X POST -H "Content-Type: application/json" -d '{"brokers":["localhost:9092"],"group_id":"my-group","topic":"test-topic","min_bytes":10000,"max_bytes":10000000,"message_limit":10,"post_url":"http://localhost:8080/api/messages","port":8080}' http://localhost:8080/config/update
  ```

---

## Логирование

Приложение использует кастомный логгер, который перенаправляет все сообщения в:
1. Терминал.
2. WebSocket для отображения логов в веб-интерфейсе.

---

## Вклад в проект

Мы приветствуем ваши предложения и улучшения! Вы можете отправить pull request или открыть issue для исправления ошибок или добавления новых функций.

---

## Лицензия

Этот проект распространяется под лицензией MIT. Подробнее см. в файле `LICENSE`.