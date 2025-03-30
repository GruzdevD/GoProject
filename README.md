# Kafka to API Service

This project is a Go service that reads messages from a Kafka topic and sends them to a specified API endpoint. It is designed to be simple and efficient, leveraging Go's concurrency features to handle message consumption and API requests. тест

## Project Structure

```
kafka-to-api-service
├── cmd
│   └── main.go          # Entry point of the application
├── internal
│   ├── kafka
│   │   └── consumer.go  # Kafka consumer implementation
│   ├── api
│   │   └── client.go    # API client for sending messages
│   └── config
│       └── config.go    # Configuration loading
├── pkg
│   └── logger
│       └── logger.go     # Logger implementation
├── go.mod                # Go module file
├── go.sum                # Go module checksums
└── README.md             # Project documentation
```

## Installation

To install the necessary dependencies, run:

```
go mod tidy
```

## Configuration

The service requires a configuration file to specify the Kafka connection details and the API endpoint. The configuration can be loaded using the `LoadConfig` function defined in `internal/config/config.go`.

## Running the Service

To run the service, execute the following command:

```
go run cmd/main.go
```

This will start the Kafka consumer, which will begin reading messages from the configured Kafka topic and sending them to the specified API endpoint.

## Logging

The service uses a custom logger defined in `pkg/logger/logger.go` to log messages at different levels (Info, Error, Debug). Ensure that logging is properly configured in your application.

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.