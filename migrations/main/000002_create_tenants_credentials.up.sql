CREATE TABLE tenant_credentials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id TEXT NOT NULL,
    credential_type TEXT NOT NULL,
    name TEXT NOT NULL,
    encrypted_data TEXT NOT NULL,
    storage_type TEXT NOT NULL DEFAULT 'vault',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, credential_type, name)
);

CREATE TABLE tenant_encryption_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id TEXT NOT NULL UNIQUE,
    use_custom_key BOOLEAN NOT NULL DEFAULT FALSE,
    custom_key_hash TEXT,
    key_derivation_salt TEXT,
    encryption_version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tenant_credentials_tenant_id ON tenant_credentials(tenant_id);
CREATE INDEX idx_tenant_credentials_type ON tenant_credentials(credential_type);

-- Auto-update triggers for updated_at
CREATE TRIGGER update_tenant_credentials_updated_at 
    AFTER UPDATE ON tenant_credentials
    BEGIN
        UPDATE tenant_credentials SET updated_at = CURRENT_TIMESTAMP 
        WHERE id = NEW.id;
    END;

CREATE TRIGGER update_tenant_encryption_config_updated_at 
    AFTER UPDATE ON tenant_encryption_config
    BEGIN
        UPDATE tenant_encryption_config SET updated_at = CURRENT_TIMESTAMP 
        WHERE id = NEW.id;
    END;