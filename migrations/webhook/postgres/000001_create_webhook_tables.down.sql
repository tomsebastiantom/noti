-- Drop triggers
DROP TRIGGER IF EXISTS update_webhook_deliveries_updated_at ON webhook_deliveries;
DROP TRIGGER IF EXISTS update_webhooks_updated_at ON webhooks;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_webhook_events_created_at;
DROP INDEX IF EXISTS idx_webhook_events_event_type;
DROP INDEX IF EXISTS idx_webhook_events_tenant_id;
DROP INDEX IF EXISTS idx_webhook_events_webhook_id;
DROP INDEX IF EXISTS idx_webhook_deliveries_next_retry;
DROP INDEX IF EXISTS idx_webhook_deliveries_status;
DROP INDEX IF EXISTS idx_webhook_deliveries_tenant_id;
DROP INDEX IF EXISTS idx_webhook_deliveries_webhook_id;
DROP INDEX IF EXISTS idx_webhooks_tenant_active;
DROP INDEX IF EXISTS idx_webhooks_tenant_id;

-- Drop tables
DROP TABLE IF EXISTS webhook_events;
DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS webhooks;
