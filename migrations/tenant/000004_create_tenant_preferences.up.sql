-- Create tenant_preferences table in tenant databases
CREATE TABLE tenant_preferences (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    channel_preferences JSONB NOT NULL,
    category_preferences JSONB NOT NULL,
    digest_settings JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add a uniqueness constraint to ensure one preference set per tenant
CREATE UNIQUE INDEX idx_tenant_preferences_tenant_unique ON tenant_preferences(tenant_id);

-- Ensure tenant has default preferences
INSERT INTO tenant_preferences (id, tenant_id, enabled, channel_preferences, category_preferences, digest_settings)
SELECT 
    gen_random_uuid(), 
    id, 
    TRUE, 
    '{"email": true, "sms": true, "push": true, "web-push": true, "in-app": true}'::JSONB, 
    '{}'::JSONB,
    '{"enabled": false, "type": "none", "intervalMinutes": 60, "deliveryHour": 9, "preferredDayOfWeek": 1, "preferredChannel": "email"}'::JSONB
FROM tenants
WHERE id NOT IN (SELECT tenant_id FROM tenant_preferences);

-- Add a trigger to automatically update the updated_at field
CREATE TRIGGER update_tenant_preferences_updated_at
BEFORE UPDATE ON tenant_preferences
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
