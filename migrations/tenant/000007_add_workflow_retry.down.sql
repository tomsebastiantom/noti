-- Remove retry policy configuration from workflows table
ALTER TABLE workflows DROP COLUMN IF EXISTS retry_policy;

-- Drop workflow dead letter queue table
DROP TABLE IF EXISTS workflow_dead_letter_queue;
