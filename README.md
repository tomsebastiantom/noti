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

### Notification Providers
- **Multi-Provider Support**: Extensible framework supporting SMS (Twilio), Email (SES), Push notifications, and Voice calls
- **Provider Flexibility**: Easy integration of new notification providers through standardized interfaces
- **Per-Tenant Provider Configuration**: Different providers and settings for each tenant

### Database & Performance
- **Multi-Database Support**: SQLite, MySQL, and PostgreSQL with automated migration system
- **Connection Pooling**: Optimized database connections with up to 40% performance improvement

### User & Tenant Management
- **User Preferences**: Granular notification preferences per user (email, SMS, push, frequency)
- **Tenant Preferences**: Organization-level notification settings and provider configurations
- **Template Management**: Customizable notification templates per provider and tenant
- **Role-Based Access**: Fine-grained permission system for tenant administrators

### Monitoring & Reliability
- **Comprehensive Logging**: Structured logging with contextual error tracking
- **Health Monitoring**: Deep health checks across all infrastructure components
- **Queue Management**: Asynchronous notification processing with worker pools
- **Circuit Breaker Pattern**: Automatic failure detection and recovery

## Getting Started

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- HashiCorp Vault
- Redis (for caching and real-time features)
- Database (PostgreSQL recommended for production)

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
    docker-compose up -d vault redis postgres
    ```

4. **Initialize Vault** (if needed):
    ```bash
    # Vault setup commands here
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

## API Documentation

### Core Endpoints
- `POST /api/v1/notifications/send` - Send notifications
- `GET /api/v1/notifications/sse/{tenant_id}` - SSE stream for real-time updates
- `POST /api/v1/webhooks/{tenant_id}` - Webhook delivery endpoint
- `GET /api/v1/users/{user_id}/preferences` - User notification preferences
- `PUT /api/v1/tenants/{tenant_id}/preferences` - Tenant configuration

### Real-Time Features
```javascript
// SSE Connection Example
const eventSource = new EventSource('/api/v1/notifications/sse/tenant123');
eventSource.onmessage = function(event) {
    const notification = JSON.parse(event.data);
    // Handle real-time notification
};
```

### Webhook Integration
```json
{
    "webhook_url": "https://your-app.com/webhooks/notifications",
    "secret": "your-webhook-secret",
    "events": ["notification.sent", "notification.failed"],
    "retry_config": {
        "max_attempts": 3,
        "backoff_seconds": [1, 5, 15]
    }
}
```

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

### Multi-Tenant Data Flow
1. Request arrives with tenant context
2. Tenant-specific database connection retrieved
3. User preferences and tenant settings loaded
4. Notification routed to configured providers
5. Real-time updates sent via SSE
6. Webhook delivery attempted with retries
7. Results logged and tracked

## Deployment

### Production Considerations
- Use PostgreSQL for production databases
- Configure Vault for credential management
- Set up Redis cluster for caching
- Enable TLS for all external communications
- Configure proper logging and monitoring

### Docker Deployment
```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes Support
Helm charts and Kubernetes manifests available in `/deploy` directory.

## Contributing

We welcome contributions! Please read our [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Code style guidelines
- Testing requirements
- Pull request process
- Development setup

## Roadmap

- [ ] GraphQL API support
- [ ] Advanced analytics and reporting
- [ ] A/B testing for notification templates
- [ ] Machine learning for delivery optimization
- [ ] Additional provider integrations
- [ ] Mobile SDKs

## License

This project is licensed under the MIT License - see [LICENSE](LICENSE) file for details.

## Support

- 📖 [Documentation](https://docs.getnoti.com)
- 💬 [Discord Community](https://discord.gg/noti)
- 🐛 [Issue Tracker](https://github.com/yourusername/noti/issues)
- 📧 [Email Support](mailto:support@getnoti.com)

---

Built with ❤️ using Go, Domain-Driven Design, and modern cloud-native patterns.