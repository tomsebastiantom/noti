# Noti - Unified Notification Service

## Overview
**Noti** is an enterprise-grade **Unified Notification Service** built in Go using Domain-Driven Design (DDD) principles. The service provides multi-tenant notification delivery across various communication channels with real-time capabilities, secure credential management, and extensible provider integrations. Each tenant operates with isolated databases and customizable notification preferences, making it ideal for SaaS platforms and enterprise applications.

## 🚀 Quick Start (Development)

### Prerequisites
- **Go 1.21+** (required)
- **Git** (required)
- **Docker** (optional, for external services)

### 1. Clone and Setup
```bash
git clone <repository-url>
cd noti
```

### 2. Environment Setup
```bash
# Copy environment template
copy .env.example .env

# Edit .env file and set REQUIRED variables:
# NOTI_ENCRYPTION_KEY=your-32-character-encryption-key
```

### 3. Run Development Server
```bash
# Quick start (uses SQLite, no external dependencies)
go run cmd/main.go

# Or use Makefile
make run
```

### 4. Access Application
- **API**: http://localhost:8072
- **SSE Events**: http://localhost:8072/v1/events/stream (requires tenant auth)

## 📋 Configuration

### Required Environment Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `NOTI_ENCRYPTION_KEY` | ✅ | 32-character encryption key for tenant credentials | `dev-encryption-key-32-bytes-long` |

### Optional Environment Variables

| Variable | Default | Description | Example |
|----------|---------|-------------|---------|
| `NOTI_HTTP_PORT` | `8072` | HTTP server port | `3000` |
| `NOTI_LOGGER_LOG_LEVEL` | `debug` (dev), `info` (prod) | Log level | `debug`, `info`, `warn`, `error` |
| `NOTI_DATABASE_DSN` | `./data/noti.db` | Database connection string | `postgres://user:pass@localhost:5432/noti` |
| `NOTI_VAULT_ADDRESS` | - | Vault server address (if enabled) | `http://localhost:8200` |
| `NOTI_VAULT_TOKEN` | - | Vault access token (if enabled) | `hvs.your-vault-token` |
| `NOTI_QUEUE_URL` | - | Message queue URL (if enabled) | `amqp://guest:guest@localhost:5672/` |
| `NOTI_CONFIG_DEBUG` | `false` | Enable config debugging | `true` |

### Configuration Files

Noti uses YAML configuration files with environment-specific overrides:

| File | Purpose | External Services |
|------|---------|-------------------|
| `config/config.yaml` | Base development config | ❌ Disabled (SQLite only) |
| `config/config.development.yaml` | Development overrides | ❌ Disabled (easy setup) |
| `config/config.production.yaml` | Production settings | ✅ Enabled (Vault + Queue) |
| `config/config.test.yaml` | Test configuration | ❌ Disabled (in-memory) |

### Configuration Flags

Control external services via configuration:

```yaml
# Enable/disable external services
queue:
  enabled: false  # Set to true to enable message queue
  url: ""        # Will use NOTI_QUEUE_URL if set

vault:
  enabled: false  # Set to true to enable Vault
  address: ""    # Will use NOTI_VAULT_ADDRESS if set

database:
  type: "sqlite"           # sqlite, postgres, mysql
  migrate_on_start: true   # Auto-run migrations
```

## 🛠️ Development Setup

### Minimal Setup (Recommended)
```bash
# 1. Set encryption key only
echo "NOTI_ENCRYPTION_KEY=dev-encryption-key-32-bytes-long" > .env

# 2. Run application
go run cmd/main.go
```

This gives you:
- ✅ SQLite database (auto-created)
- ✅ All APIs working
- ✅ No external dependencies
- ✅ Mock notification providers (log output)

### Full Development Setup (Optional)
```bash
# 1. Start external services
docker-compose up -d vault rabbitmq

# 2. Configure environment
cat > .env << EOF
NOTI_ENCRYPTION_KEY=dev-encryption-key-32-bytes-long
NOTI_VAULT_ADDRESS=http://localhost:8200
NOTI_VAULT_TOKEN=your-vault-token
NOTI_QUEUE_URL=amqp://guest:guest@localhost:5672/
EOF

# 3. Enable services in config
# Edit config/config.development.yaml:
# queue:
#   enabled: true
# vault:
#   enabled: true
```

### Testing Setup
```bash
# Set test encryption key
echo "NOTI_TEST_ENCRYPTION_KEY=test-encryption-key-32-bytes-abc" > .env.test

# Run tests
go test ./...
```

## 🏗️ Production Setup

### Required Production Variables
```bash
# Database (required)
NOTI_DATABASE_DSN=postgres://user:password@localhost:5432/noti?sslmode=disable

# Encryption key (required)
NOTI_ENCRYPTION_KEY=your-production-encryption-key-32-chars

# Vault (recommended)
NOTI_VAULT_ADDRESS=https://vault.yourcompany.com
NOTI_VAULT_TOKEN=your-production-vault-token

# Queue (recommended for scale)
NOTI_QUEUE_URL=amqp://user:password@rabbitmq.yourcompany.com:5672/
```

### Production Configuration
Use `config/config.production.yaml` which automatically:
- ✅ Enables Vault for secure credential storage
- ✅ Enables message queue for scalability
- ✅ Uses PostgreSQL database
- ✅ Sets production logging levels
- ✅ Disables auto-migration (manual control)

## 📊 Architecture Overview

### Worker Pool Management
- **Scalable Worker Pools**: Efficient processing for notifications and scheduled tasks
- **Circuit Breaker Pattern**: Automatic failure detection and recovery
- **Job Retry System**: Intelligent retry mechanism with exponential backoff for failed operations

## Key Features

### Monitoring & Observability
- **Service Health Checks**: Built-in health monitoring for database, queue, worker pools, and credential manager
- **Internal Metrics**: Event bus metrics, queue metrics, and worker pool performance tracking
- **Logging**: Comprehensive structured logging with configurable levels

### Core Architecture
- **Multi-Tenant Support**: Complete tenant isolation with per-tenant databases and secure credential management
- **Domain-Driven Design**: Clean architecture with separated domains, use cases, and infrastructure layers
- **Dependency Injection**: Centralized service container for efficient resource management and testing
- **Graceful Shutdown**: Production-ready lifecycle management with timeout-based cleanup
- **Task Scheduler**: Robust CRON-based scheduling system for recurring tasks with failure recovery

### Security & Credentials
- **HashiCorp Vault Integration**: Secure credential storage with database fallback for high availability
- **Per-Tenant Credential Isolation**: Automated secret management across multiple client environments
- **Zero Credential Exposure**: Production-safe credential handling with encrypted storage
- **Flexible Credential Storage**: Option to store encrypted credentials directly in database when Vault is unavailable

### Real-Time Communication
- **Server-Sent Events (SSE)**: Real-time notification streaming to web clients with automatic reconnection
- **Connection Optimization**: Efficient connection pooling and caching for improved performance
- **Live Dashboard Support**: Real-time status updates and notification tracking

### Webhook System
- **Robust Webhook Delivery**: Reliable webhook delivery with retry logic and circuit breaker patterns
- **Per-Tenant Webhook Configuration**: Customizable webhook endpoints and authentication per client
- **Delivery Guarantees**: Comprehensive error handling and failure recovery mechanisms
- **Security Manager**: Built-in webhook security validation and signature verification
- **Scheduled Webhooks**: CRON-based webhook scheduling for time-triggered integrations

### Notification Providers
- **Multi-Provider Support**: Extensible framework supporting SMS (Twilio), Email (SES), Push notifications, and Voice calls
- **Provider Flexibility**: Easy integration of new notification providers through standardized interfaces
- **Per-Tenant Provider Configuration**: Different providers and settings for each tenant
- **Scheduled Notifications**: CRON-based scheduling for delayed and recurring notifications

### Database & Performance
- **Multi-Database Support**: SQLite, MySQL, and PostgreSQL with automated migration system
- **Connection Pooling**: Optimized database connections with up to 40% performance improvement
- **Queue Management**: Optional queue support for asynchronous notification processing
- **Data Retention**: Automated cleanup of historical data based on configurable retention policies

### User & Tenant Management
- **User Preferences**: Granular notification preferences per user (email, SMS, push, frequency)
- **Tenant Preferences**: Organization-level notification settings and provider configurations
- **Template Management**: Customizable notification templates per provider and tenant
- **Repository Factory**: Dynamic repository creation for multi-tenant data access

### Task Scheduler
- **CRON-based Scheduling**: Flexible scheduling using cron expressions for recurring tasks
- **Multiple Task Types**: Support for notification, webhook, cleanup, report generation, and workflow execution tasks
- **Execution Tracking**: Detailed execution history with status tracking and result storage
- **Failure Management**: Comprehensive error handling with configurable retry policies
- **Multi-tenant Isolation**: Per-tenant schedule management and execution
- **Workflow Integration**: Seamless integration with workflow engine for scheduled workflow executions

## 🔧 Common Development Tasks

### Generate Encryption Key
```bash
# Linux/macOS (generates Base64-encoded 32-byte key)
openssl rand -base64 32

# Windows PowerShell (generates Base64-encoded 32-byte key)
[System.Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Maximum 256 }))
```

### Switch Database Types
```bash
# SQLite (default)
# No configuration needed

# PostgreSQL
NOTI_DATABASE_DSN=postgres://user:pass@localhost:5432/noti?sslmode=disable

# MySQL
NOTI_DATABASE_DSN=user:pass@tcp(localhost:3306)/noti?parseTime=true
```

### Enable External Services
```yaml
# In config/config.development.yaml
queue:
  enabled: true
vault:
  enabled: true
```

### Run with Different Environments
```bash
# Development (default)
go run cmd/main.go

# Test environment
NOTI_ENV=test go run cmd/main.go

# Production
NOTI_ENV=production go run cmd/main.go
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Built with ❤️ using Go, Domain-Driven Design, and modern cloud-native patterns.
