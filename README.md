# Noti - Unified Notification Service

## Overview
**Noti** is an enterprise-grade **Unified Notification Service** built in Go using Domain-Driven Design (DDD) principles. The service provides multi-tenant notification delivery across various communication channels with real-time capabilities, secure credential management, and extensible provider integrations. Each tenant operates with isolated databases and customizable notification preferences, making it ideal for SaaS platforms and enterprise applications.

## Key Features

### Core Architecture
- **Multi-Tenant Support**: Complete tenant isolation with per-tenant databases and secure credential management
- **Domain-Driven Design**: Clean architecture with separated domains, use cases, and infrastructure layers
- **Dependency Injection**: Centralized service container for efficient resource management and testing
- **Graceful Shutdown**: Production-ready lifecycle management with timeout-based cleanup

### Security & Credentials
- **HashiCorp Vault Integration**: Secure credential storage with database fallback for high availability
- **Per-Tenant Credential Isolation**: Automated secret management across multiple client environments
- **Zero Credential Exposure**: Production-safe credential handling with encrypted storage

### Real-Time Communication
- **Server-Sent Events (SSE)**: Real-time notification streaming to web clients with automatic reconnection
- **Connection Optimization**: Efficient connection pooling and caching for improved performance
- **Live Dashboard Support**: Real-time status updates and notification tracking

### Webhook System
- **Robust Webhook Delivery**: Reliable webhook delivery with retry logic and circuit breaker patterns
- **Per-Tenant Webhook Configuration**: Customizable webhook endpoints and authentication per client
- **Delivery Guarantees**: Comprehensive error handling and failure recovery mechanisms
- **Security Manager**: Built-in webhook security validation and signature verification

### Notification Providers
- **Multi-Provider Support**: Extensible framework supporting SMS (Twilio), Email (SES), Push notifications, and Voice calls
- **Provider Flexibility**: Easy integration of new notification providers through standardized interfaces
- **Per-Tenant Provider Configuration**: Different providers and settings for each tenant

### Database & Performance
- **Multi-Database Support**: SQLite, MySQL, and PostgreSQL with automated migration system
- **Connection Pooling**: Optimized database connections with up to 40% performance improvement
- **Queue Management**: Optional queue support for asynchronous notification processing

### User & Tenant Management
- **User Preferences**: Granular notification preferences per user (email, SMS, push, frequency)
- **Tenant Preferences**: Organization-level notification settings and provider configurations
- **Template Management**: Customizable notification templates per provider and tenant
- **Repository Factory**: Dynamic repository creation for multi-tenant data access

### Monitoring & Reliability
- **Comprehensive Logging**: Structured logging with contextual error tracking
- **Health Monitoring**: Deep health checks across all infrastructure components
- **Worker Pool Management**: Scalable worker pools for notification processing
- **Circuit Breaker Pattern**: Automatic failure detection and recovery

## Getting Started

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- HashiCorp Vault
- Database (PostgreSQL recommended for production)
- Message Queue (optional - for async processing)

### Quick Start
1. **Clone the repository**:
    ```bash
    git clone https://github.com/yourusername/noti.git
    cd noti
    ```

2. **Environment Setup**:
    ```bash
    cp .env.example .env
    # Edit .env with your configuration
    ```

3. **Start Infrastructure**:
    ```bash
    docker-compose up -d vault postgres
    ```

4. **Initialize Vault** (if needed):
    ```bash
    export VAULT_ADDR=http://127.0.0.1:8200
    export VAULT_TOKEN=your-token
    ```

5. **Run the Application**:
    ```bash
    go run cmd/main.go
    ```

### Configuration

The service supports multiple configuration methods:
- Environment variables
- Configuration files (`config/config.yaml`)
- HashiCorp Vault for sensitive data

Key configuration areas:
- **Database Settings**: Per-tenant database connections
- **Provider Credentials**: Third-party service API keys (stored in Vault)
- **Webhook Endpoints**: Client webhook URLs and authentication
- **Real-time Settings**: SSE connection limits and timeouts
- **Queue Configuration**: Optional message queue for async processing

## Architecture

### Domain Structure
```
internal/
├── domain/          # Core business logic
│   ├── notification/
│   ├── tenant/
│   ├── user/
│   └── webhook/
├── usecase/         # Application services
├── infrastructure/  # External integrations
├── container/       # Dependency injection
└── server/         # HTTP layer
```

Built with ❤️ using Go, Domain-Driven Design, and modern cloud-native patterns.