-- Add retry and dead letter queues for workflow executions
CREATE TABLE workflow_dead_letter_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES workflow_executions(id),
    step_id VARCHAR(255),
    error_reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add retry policy configuration
ALTER TABLE workflows
ADD COLUMN retry_policy JSONB DEFAULT '{
    "maxAttempts": 3,
    "initialInterval": "30s",
    "maxInterval": "1h",
    "multiplier": 2,
    "maxElapsedTime": "24h"
}'::jsonb;
