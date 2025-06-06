-- Create schedules table
CREATE TABLE IF NOT EXISTS schedules (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    cron_expression VARCHAR(100) NOT NULL,
    config TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    last_execution_at TIMESTAMP NULL,
    next_execution_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_schedules_tenant_id (tenant_id),
    INDEX idx_schedules_active (is_active),
    INDEX idx_schedules_next_execution (next_execution_at),
    INDEX idx_schedules_type (type),
    INDEX idx_schedules_tenant_active (tenant_id, is_active)
);

-- Create schedule_executions table
CREATE TABLE IF NOT EXISTS schedule_executions (
    id VARCHAR(36) PRIMARY KEY,
    schedule_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) NOT NULL,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    error_message TEXT NULL,
    result TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (schedule_id) REFERENCES schedules(id) ON DELETE CASCADE,
    INDEX idx_executions_schedule_id (schedule_id),
    INDEX idx_executions_status (status),
    INDEX idx_executions_created_at (created_at),
    INDEX idx_executions_schedule_status (schedule_id, status)
);
