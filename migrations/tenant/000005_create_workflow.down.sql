-- Drop triggers
DROP TRIGGER IF EXISTS update_step_executions_updated_at ON step_executions;
DROP TRIGGER IF EXISTS update_workflow_executions_updated_at ON workflow_executions;
DROP TRIGGER IF EXISTS update_workflows_updated_at ON workflows;

-- Drop tables in reverse order (child tables first)
DROP TABLE IF EXISTS step_executions;
DROP TABLE IF EXISTS workflow_executions;
DROP TABLE IF EXISTS workflows;
