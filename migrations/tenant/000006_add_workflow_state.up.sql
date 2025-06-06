-- Add state management columns to workflow_executions
ALTER TABLE workflow_executions
ADD COLUMN state_data JSONB,
ADD COLUMN checkpoint_version INT DEFAULT 0,
ADD COLUMN recovery_count INT DEFAULT 0,
ADD COLUMN retry_strategy JSONB,
ADD COLUMN next_retry_at TIMESTAMP;

-- Add index for state data
CREATE INDEX idx_workflow_executions_state ON workflow_executions(state_data) WHERE state_data IS NOT NULL;
CREATE INDEX idx_workflow_executions_next_retry ON workflow_executions(next_retry_at) WHERE next_retry_at IS NOT NULL;
