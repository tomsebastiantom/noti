-- Create workflow schedules table
CREATE TABLE workflow_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    tenant_id VARCHAR(255) NOT NULL,
    cron_expression VARCHAR(100) NOT NULL,
    payload JSONB,
    is_active BOOLEAN DEFAULT true,
    last_execution_at TIMESTAMP,
    next_execution_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for workflow schedules
CREATE INDEX idx_workflow_schedules_tenant_id ON workflow_schedules(tenant_id);
CREATE INDEX idx_workflow_schedules_workflow_id ON workflow_schedules(workflow_id);
CREATE INDEX idx_workflow_schedules_next_execution ON workflow_schedules(next_execution_at) 
    WHERE is_active = true;

-- Add trigger for updated_at
CREATE TRIGGER update_workflow_schedules_updated_at 
    BEFORE UPDATE ON workflow_schedules 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
