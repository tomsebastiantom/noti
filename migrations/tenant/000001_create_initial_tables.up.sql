
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20),
    device_id VARCHAR(255),
    web_push_token VARCHAR(255),
    consents JSONB NOT NULL,
    preferred_mode VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE templates (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    content TEXT NOT NULL,
    is_public BOOLEAN NOT NULL,
    variables JSONB NOT NULL
);

CREATE TABLE providers (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE provider_channels (
    provider_id UUID,
    channel_type INT,
    enabled BOOLEAN,
    priority INT,
    PRIMARY KEY (provider_id, channel_type),
    FOREIGN KEY (provider_id) REFERENCES providers(id)
);

CREATE TABLE tenant_metadata (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    channel VARCHAR(50) NOT NULL,
    template_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    provider_id UUID,
    variables JSONB NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (template_id) REFERENCES templates(id),
    FOREIGN KEY (provider_id) REFERENCES providers(id)
);
