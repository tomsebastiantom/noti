package domain

import "errors"

var (
	// Workflow errors
	ErrWorkflowNotFound     = errors.New("workflow not found")
	ErrWorkflowNoSteps      = errors.New("workflow must have at least one step")
	ErrWorkflowInactive     = errors.New("workflow is not active")
	ErrWorkflowAlreadyExists = errors.New("workflow with this identifier already exists")
	
	// Execution errors
	ErrExecutionNotFound    = errors.New("execution not found")
	ErrExecutionAlreadyRunning = errors.New("execution is already running")
	ErrExecutionCompleted   = errors.New("execution is already completed")
	ErrExecutionFailed      = errors.New("execution has failed")
	
	// Step errors
	ErrStepNotFound         = errors.New("step not found")
	ErrStepProcessorNotFound = errors.New("step processor not found")
	ErrStepValidation       = errors.New("step validation failed")
	ErrStepConditionFailed  = errors.New("step condition failed")
	
	// General errors
	ErrInvalidTenantID      = errors.New("invalid tenant ID")
	ErrInvalidPayload       = errors.New("invalid payload")
	ErrTriggerNotFound      = errors.New("trigger not found")
)
