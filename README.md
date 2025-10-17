# Service Manager API

A RESTful API built with Go and Gin for managing and monitoring background services. This application allows you to register, start, stop, and remove long-running processes, as well as stream their logs in real-time.

## Features

- Register and manage background services.
- Start, stop, and remove services via API calls.
- Persists service configurations to a JSON file.
- Real-time `stdout` and `stderr` log streaming.
- View service status and resource metrics (CPU/RAM).
- Automatic API documentation with Swagger.

## Getting Started

### Prerequisites

- Go (version 1.21 or later recommended).

### Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/your-username/service-manager-go.git
   ```
2. Navigate to the project directory:
    ```sh
    cd service-manager-go
    ```
3. Install dependencies:
    ```sh
    go mod tidy
    ```

### Configuration

Create a `.env` file in the root directory with the following variables:

```
HOST=0.0.0.0
PORT=8080
LOGS_DIR=data/logs
SERVICES_DATA=data/services_data.json
```

### Running the Application

```sh
go run ./cmd/server/main.go
```

The server will start on the configured `HOST` and `PORT` (defaulting to `0.0.0.0:8080`).

## API Endpoints

| Method   | Endpoint                   | Description                        | Payload Example                                                                                             |
| :------- | :------------------------- | :--------------------------------- | :---------------------------------------------------------------------------------------------------------- |
| `POST`   | `/manager/register`        | Register a new service.            | `{"name": "My App", "command": "python", "args": ["-u", "main.py"], "directory": "/path/to/your/app"}` |
| `GET`    | `/manager/services`        | Get a list of all registered services. | N/A                                                                                                         |
| `POST`   | `/manager/start`           | Start a registered service.        | `{"id": "your-service-id"}`                                                                                 |
| `POST`   | `/manager/stop`            | Stop a running service.            | `{"id": "your-service-id"}`                                                                                 |
| `DELETE` | `/manager/remove`          | Remove a stopped service.          | `{"id": "your-service-id"}`                                                                                 |
| `POST`   | `/manager/metrics`         | Get CPU and RAM usage for a service. | `{"id": "your-service-id"}`                                                                                 |
| `POST`   | `/manager/network`         | Get network info for a service.    | `{"id": "your-service-id"}`                                                                                 |
| `GET`    | `/stream/stdout/:serviceID`| Stream stdout logs for a service.  | N/A                                                                                                         |
| `GET`    | `/stream/stderr/:serviceID`| Stream stderr logs for a service.  | N/A                                                                                                         |

## API Documentation

This project uses Swagger for automatic API documentation. Once the server is running, you can access the interactive Swagger UI at:

[http://localhost:8080/docs/index.html](http://localhost:8080/docs/index.html)
