# Noti - Unified Notification Service

## Overview
**Noti** is an enterprise-grade **Unified Notification Service** built in Go using Domain-Driven Design (DDD) principles. The service provides multi-tenant notification delivery across various communication channels with real-time capabilities, secure credential management, and extensible provider integrations. Each tenant operates with isolated databases and customizable notification preferences, making it ideal for SaaS platforms and enterprise applications.

### Worker Pool Management
- **Scalable Worker Pools**: Efficient processing for notifications and scheduled tasks
- **Circuit Breaker Pattern**: Automatic failure detection and recovery
- **Job Retry System**: Intelligent retry mechanism with exponential backoff for failed operations

## Key Features

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
- **Job Priority Management**: Priority-based job processing with queue management
- **Worker Pool Optimization**: Intelligent worker allocation for different task types

### Workflow Engine
- **Multi-Step Workflows**: Chain multiple notification steps with conditional logic
- **Event-Driven Triggers**: Start workflows based on events, schedules, or webhooks
- **Scheduled Workflow Execution**: CRON-based scheduling for automated workflow execution with per-tenant isolation
- **Conditional Branching**: Route workflow execution based on dynamic conditions
- **Delay Steps**: Time-based delays between workflow steps
- **Digest Support**: Batch notifications over configurable time periods
- **Channel Routing**: Smart routing to different channels based on user preferences
- **Template Integration**: Rich template support for each workflow step
- **Execution Tracking**: Complete audit trail of workflow executions and step results
- **Job Processing**: Robust job processing system with retry logic and failure management
- **State Management**: Comprehensive workflow state tracking and recovery mechanisms
- **Multi-tenant Scheduling**: Isolated workflow schedules per tenant with shared execution infrastructure

### Workflow Scheduling
- **CRON-Based Scheduling**: Schedule workflows to run automatically using flexible CRON expressions
- **Per-Tenant Isolation**: Each tenant can define their own workflow schedules without interference
- **Schedule Management**: Create, update, enable/disable, and delete workflow schedules through API
- **Execution History**: Track all scheduled workflow executions with detailed status and result information
- **Timezone Support**: Configure schedules with specific timezone handling for global deployments
- **Integration with Task Scheduler**: Seamless integration with the general task scheduler for unified job processing
- **Error Handling**: Comprehensive error handling and retry logic for failed scheduled executions
- **Resource Management**: Efficient resource allocation and cleanup for scheduled workflow jobs

### Monitoring & Reliability
- **Comprehensive Logging**: Structured logging with contextual error tracking
- **Health Monitoring**: Deep health checks across all infrastructure components
- **Worker Pool Management**: Scalable worker pools for notification processing
- **Circuit Breaker Pattern**: Automatic failure detection and recovery

## Getting Started

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- HashiCorp Vault (optional - database credential storage fallback available)
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
- **Provider Credentials**: Third-party service API keys (stored in Vault or encrypted in database)
- **Webhook Endpoints**: Client webhook URLs and authentication
- **Real-time Settings**: SSE connection limits and timeouts
- **Queue Configuration**: Optional message queue for async processing
- **Scheduler Settings**: CRON expressions and execution parameters for scheduled tasks
- **Workflow Scheduling**: Configuration for automated workflow execution schedules

## Architecture

### Domain Structure
```
internal/
├── domain/          # Core business logic
│   ├── notification/
│   ├── tenant/
│   ├── user/
│   ├── webhook/
│   └── workflows/   # Workflow management
├── shared/          # Shared components
│   ├── scheduler/   # Task scheduling system
│   ├── events/      # Event management
│   └── utils/       # Utility functions
├── usecase/         # Application services
├── infrastructure/  # External integrations
├── container/       # Dependency injection
└── server/          # HTTP layer
```

Built with ❤️ using Go, Domain-Driven Design, and modern cloud-native patterns.