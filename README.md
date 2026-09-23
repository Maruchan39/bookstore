# Bookstore

## Local development with Docker MySQL

The Go application runs on the host and connects to MySQL running in Docker.

### Prerequisites

- Go
- Docker

### Configure and start MySQL

Copy the example configuration, set a local JWT secret and root password.

```sh
cp .env.example .env
# Edit .env and replace the JWT_SECRET and MYSQL_ROOT_PASSWORD values.
docker compose up -d
```

### Run the application

Build and start the Go server:

```sh
go build . && go run .
```
