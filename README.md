# Noti - Unified Notification Service

## Overview
This project, **Noti**, is a **Unified Notification Service** written in Go using Domain-Driven Design (DDD). The service aims to unify the sending of notifications across various tenants, each with a unique database. Database credentials are securely stored in HashiCorp Vault, and connections are cached for efficiency. Notifications are routed to various providers such as Twilio, Amazon SES, and others, and can differ for each tenant. The service supports sending push notifications, SMS, calls, and other types of notifications. Providers can be added as needed, and each client receives notifications tied to their specific database.

## Features
- **Multi-Tenant Support**: Each tenant has a unique database.
- **Secure Credential Storage**: Database credentials are stored in HashiCorp Vault.
- **Connection Caching**: Database connections are cached for performance.
- **Provider Flexibility**: Supports multiple notification providers (e.g., Twilio, Amazon SES).
- **Notification Types**: Supports push notifications, SMS, calls, and more.
- **Template Management**: Templates can be stored and customized per provider and tenant.
- **Database Support**: Currently supports SQLite, MySQL, and PostgreSQL, with plans to add more.
- **Queue Management**: Notifications are queued for efficient processing.
- **Monitoring and Analytics**: Plans to add monitoring and analytics for open rates and tracking.
- **Microservice Architecture**: Designed to run in an authenticated environment, recommended to use Keycloak or Hydra for authentication.

## Getting Started

### Prerequisites
- Go 1.16+
- Docker (for running databases and other services)
- HashiCorp Vault
- Keycloak or Hydra (for authentication)

### Installation
1. **Clone the repository**:
    ```bash
    git clone https://github.com/yourusername/noti.git
    cd noti
    ```

2. **Set up environment variables**:
    Create a `.env` file in the root directory and add the necessary environment variables:
    ```env
    VAULT_ADDR=http://127.0.0.1:8200
    VAULT_TOKEN=myroot
    ```

3. **Run the services**:
    ```bash
    docker-compose up -d
    ```

4. **Run the application**:
    ```bash
    go run main.go
    ```

### Configuration
- **Database Configuration**: Configure your database settings in `config.toml`.
- **Provider Configuration**: Add your provider credentials and settings in `providers.toml`.

### Usage
- **API Endpoints**: The API endpoints are documented using Swagger. Access the documentation at `http://localhost:8000/swagger/index.html`.
- **Sending Notifications**: Use the `/send` endpoint to send notifications. Example payload:
    ```json
    {
        "tenant_id": "tenant123",
        "notification_type": "email",
        "provider": "twilio",
        "message": "Hello, this is a test notification!"
    }
    ```

## Contributing
We welcome contributions! Please read our [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments
- Inspired by various DDD resources and Go projects.
- Special thanks to the contributors and the open-source community.

---

Feel free to reach out if you have any questions or need further assistance. Happy coding!
