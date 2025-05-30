DROP TRIGGER IF EXISTS update_tenant_encryption_config_updated_at;
DROP TRIGGER IF EXISTS update_tenant_credentials_updated_at;
DROP INDEX IF EXISTS idx_tenant_credentials_type;
DROP INDEX IF EXISTS idx_tenant_credentials_tenant_id;
DROP TABLE IF EXISTS tenant_encryption_config;
DROP TABLE IF EXISTS tenant_credentials;