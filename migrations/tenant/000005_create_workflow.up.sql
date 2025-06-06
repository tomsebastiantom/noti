-- Create workflows table
CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    trigger JSONB NOT NULL,
    steps JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT workflows_status_check CHECK (status IN ('draft', 'active', 'paused', 'archived')),
    CONSTRAINT workflows_tenant_name_unique UNIQUE (tenant_id, name)
);

-- Create index on tenant_id for performance
CREATE INDEX idx_workflows_tenant_id ON workflows(tenant_id);

-- Create index on trigger identifier for lookup
CREATE INDEX idx_workflows_trigger_identifier ON workflows USING GIN ((trigger->>'identifier'));

-- Create index on status for filtering
CREATE INDEX idx_workflows_status ON workflows(status);

-- Create workflow_executions table
CREATE TABLE workflow_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    tenant_id VARCHAR(255) NOT NULL,
    trigger_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    payload JSONB,
    context JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    
    CONSTRAINT workflow_executions_status_check CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled', 'paused'))
);

-- Create indexes for workflow_executions
CREATE INDEX idx_workflow_executions_workflow_id ON workflow_executions(workflow_id);
CREATE INDEX idx_workflow_executions_tenant_id ON workflow_executions(tenant_id);
CREATE INDEX idx_workflow_executions_status ON workflow_executions(status);
CREATE INDEX idx_workflow_executions_trigger_id ON workflow_executions(trigger_id);
CREATE INDEX idx_workflow_executions_created_at ON workflow_executions(created_at);

-- Create step_executions table
CREATE TABLE step_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES workflow_executions(id) ON DELETE CASCADE,
    step_id VARCHAR(255) NOT NULL,
    step_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    result JSONB,
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    delay_until TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT step_executions_status_check CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled', 'paused')),
    CONSTRAINT step_executions_step_type_check CHECK (step_type IN ('email', 'sms', 'push', 'webhook', 'delay', 'digest', 'condition'))
);

-- Create indexes for step_executions
CREATE INDEX idx_step_executions_execution_id ON step_executions(execution_id);
CREATE INDEX idx_step_executions_status ON step_executions(status);
CREATE INDEX idx_step_executions_step_type ON step_executions(step_type);
CREATE INDEX idx_step_executions_delay_until ON step_executions(delay_until) WHERE delay_until IS NOT NULL;
CREATE INDEX idx_step_executions_pending ON step_executions(created_at) WHERE status = 'pending';

-- Create updated_at trigger function if it doesn't exist
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_workflows_updated_at 
    BEFORE UPDATE ON workflows 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_workflow_executions_updated_at 
    BEFORE UPDATE ON workflow_executions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_step_executions_updated_at 
    BEFORE UPDATE ON step_executions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
