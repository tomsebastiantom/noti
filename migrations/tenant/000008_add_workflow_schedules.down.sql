-- Drop trigger for updated_at
DROP TRIGGER IF EXISTS update_workflow_schedules_updated_at ON workflow_schedules;

-- Drop indexes
DROP INDEX IF EXISTS idx_workflow_schedules_next_execution;
DROP INDEX IF EXISTS idx_workflow_schedules_workflow_id;
DROP INDEX IF EXISTS idx_workflow_schedules_tenant_id;

-- Drop workflow schedules table
DROP TABLE IF EXISTS workflow_schedules;
