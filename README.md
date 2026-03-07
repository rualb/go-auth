# go-auth

An authentication microservice written in Go. This service provides a centralized identity management system supporting multi-mode authentication (Email and Telephone), multi-step verification flows, and built-in security features.

Designed to be decoupled and scalable, `go-auth` can serve as the identity provider for various distributed applications.

## 🚀 Features

- **Dual-Mode Authentication**: Supports both Email and Telephone-based identity.
- **Multi-Step Verification**: Secure flows for Signup and Forgot Password using OTP (TOTP) codes and temporary tokens.
- **JWT-Based Sessions**: Stateless authentication using JSON Web Tokens with support for automatic rotation.
- **Bot & Rate Limiting**: Built-in protection against brute-force and automated attacks using an in-memory limit manager (IP and Account based).
- **Internationalization (i18n)**: Flexible translation system with support for multi-language responses.
- **Database Agnostic**: Built on GORM; supports PostgreSQL and SQLite.
- **Security First**: 
    - Password hashing using Bcrypt.
    - Graceful shutdown for clean resource management.
    - Vault system for managing cryptographic keys (Keychain support).
- **Monitoring**: Integrated Prometheus metrics and health check endpoints.

## 🛠 Tech Stack

- **Language**: Go 1.22+
- **Web Framework**: [Echo v4](https://github.com/labstack/echo)
- **ORM**: [GORM](https://gorm.io/)
- **Authentication**: JWT (v5), TOTP
- **Logging**: Structured logging using `slog`

## 📂 Project Structure

- `cmd/`: Application entry point.
- `internal/cmd/`: Logic for server startup and graceful shutdown.
- `internal/controller/`: HTTP handlers organized by functionality (Account, Auth, Email, Tel).
- `internal/service/`: Core business logic (Account management, Sign-in flows, Vault/Key management).
- `internal/repository/`: Data access layer.
- `internal/token/`: JWT and OTP generation/validation logic.
- `internal/util/`: Helper packages for crypto, string manipulation, and bot limiting.
- `web/`: Embedded static assets and index templates.

## ⚙️ Configuration

The service uses a flexible configuration system that reads from JSON files, environment variables, and CLI flags.

### Environment Variables
Environment variables should be prefixed with `APP_`. For example:
- `APP_ENV`: Environment (development, testing, production).
- `APP_DB_DIALECT`: `postgres` or `sqlite`.
- `APP_DB_HOST`: Database host.
- `APP_IDENTITY_IS_AUTH_TEL`: Enable/Disable phone authentication (`true`/`false`).
- `APP_IDENTITY_IS_AUTH_EMAIL`: Enable/Disable email authentication (`true`/`false`).
- `APP_HTTP_SERVER_LISTEN`: Address to listen on (e.g., `127.0.0.1:30280`).

### Configuration Files
The service looks for `config.{env}.json` in directories specified by the `-config` flag or `APP_CONFIG` environment variable.

## 🚀 Getting Started

### Prerequisites
- Go 1.22 or higher
- PostgreSQL (optional, defaults to SQLite if configured)

### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/your-repo/go-auth.git
   cd go-auth
   ```

2. Download dependencies:
   ```bash
   go mod download
   ```

3. Run the application:
   ```bash
   go run cmd/go-auth/main.go -env development
   ```

### Running Tests
To run the full test suite (including end-to-end tests):
```bash
go test ./...
```

## 🛣 API Endpoints (Summary)

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/auth/api/signin/email` | POST | Sign in using email and password |
| `/auth/api/signup/tel` | POST | Multi-step phone registration |
| `/auth/api/forgot-password` | POST | Password reset flow |
| `/auth/api/status` | GET | Check current authentication status |
| `/auth/api/signout` | POST | Clear session cookies |
| `/sys/api/metrics` | GET | Prometheus metrics (protected by API Key) |

## 🛡 Security Note

Ensure that `APP_HTTPSERVER_SYSAPIKEY` is set to a strong, random string in production. This key protects sensitive monitoring and metrics endpoints. Cryptographic keys for JWT and OTP are managed via the `Vault` service and should be rotated according to your security policy.

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.