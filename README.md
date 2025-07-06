# Testing Trainer

A Go-based backend service for managing and tracking testing training sessions and progress.

## Project Structure

The project follows a clean architecture pattern with the following structure:

```
.
├── cmd/            # Application entry points
├── internal/       # Private application code
│   ├── app/       # Application setup and configuration
│   ├── entities/  # Domain entities
│   ├── storage/   # Data storage implementations
│   └── usecase/   # Business logic
├── pkg/           # Public libraries
├── config/        # Configuration files
├── etc/          # Additional configuration and resources
├── middlewares/   # HTTP middleware components
├── scripts/      # Utility scripts
└── utils/        # Utility functions
```

## Technologies

- **Go 1.23.1**: Main programming language
- **Gin**: Web framework
- **PostgreSQL**: Database
- **Docker**: Containerization
- **JWT**: Authentication
- **gRPC**: RPC framework
- **Swagger**: API documentation

## Prerequisites

- Go 1.23.1 or higher
- Docker and Docker Compose
- PostgreSQL (if running locally)

## Getting Started

1. Clone the repository:
```bash
git clone <repository-url>
cd testing_trainer
```

2. Start the services using Docker Compose. Two configurations are available and can run at the same time:
```bash
# development configuration
docker-compose -f docker-compose.yaml up -d
# production-like configuration
docker-compose -f docker-compose.prod.yaml up -d
```

`docker-compose.yaml` exposes the database on port `5432` and the API on `7001`.
`docker-compose.prod.yaml` exposes the database on port `5433` and the API on `7002`.

Each compose file loads its own environment settings:
`docker-compose.yaml` mounts `./etc/config.env` while `docker-compose.prod.yaml`
mounts `./etc/prod.env`.

## Environment Variables

The application uses the following environment variables:

- `TOKEN_MINUTE_LIFESPAN`: JWT token expiration time in minutes (default: 24)
- `REFRESH_HOUR_LIFESPAN`: Refresh token expiration time in hours (default: 48)
- `API_SECRET`: Secret key for API JWT tokens
- `REFRESH_SECRET`: Secret key for refresh tokens

## API Documentation

The API documentation is available through Swagger UI at `/swagger/index.html` when the application is running.

## Development

### Building the Application

```bash
go build -o app ./cmd/api
```

### Running Tests

```bash
go test ./...
```

### Database Migrations

The project uses Goose for database migrations. To run migrations:

```bash
goose -dir ./migrations up
```

## Docker

The project includes Docker configuration for both the application and database:

- `Dockerfile`: Application container configuration
- `docker-compose.yaml`: Multi-container setup with PostgreSQL

## License

[Add your license information here]

## Contributing

[Add contribution guidelines here] 