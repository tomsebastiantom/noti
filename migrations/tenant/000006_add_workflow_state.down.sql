-- Remove state management columns from workflow_executions
DROP INDEX IF EXISTS idx_workflow_executions_state;
DROP INDEX IF EXISTS idx_workflow_executions_next_retry;

ALTER TABLE workflow_executions
DROP COLUMN IF EXISTS state_data,
DROP COLUMN IF EXISTS checkpoint_version,
DROP COLUMN IF EXISTS recovery_count,
DROP COLUMN IF EXISTS retry_strategy,
DROP COLUMN IF EXISTS next_retry_at;
