# go-auth

`go-auth` is a microservice built with Go and the [Echo](https://echo.labstack.com/) framework, responsible for handling user authentication and authorization within the larger architecture. 

## Features

- **Authentication:** JWT-based user authentication.
- **Database:** Uses GORM for ORM with PostgreSQL (and SQLite for testing/local development).
- **Security:** Built-in password hashing and security measures using `golang.org/x/crypto`.
- **Metrics:** Exposes Prometheus metrics (`github.com/prometheus/client_golang`).
- **REST API:** High-performance HTTP server using Echo.

## Prerequisites

- Go 1.26+
- Python 3.x (for the build script)

## Build and Run

Like other services in this repository, `go-auth` uses `Makefile.py` for operations.

```sh
# Run tests
python Makefile.py test

# Run linter
python Makefile.py lint

# Build binary
python Makefile.py linux
```

## Architecture Context

`go-auth` is part of a larger microservices architecture. It provides the central authentication mechanism that other services rely on. See the root `README.md` and `projects/ecom-shop/README.md` for more details.
