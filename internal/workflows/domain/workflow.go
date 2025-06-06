package domain

import (
	"time"

	"github.com/google/uuid"
)

type WorkflowStatus string

const (
	WorkflowStatusDraft    WorkflowStatus = "draft"
	WorkflowStatusActive   WorkflowStatus = "active"
	WorkflowStatusPaused   WorkflowStatus = "paused"
	WorkflowStatusArchived WorkflowStatus = "archived"
)

type Workflow struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	TenantID    string          `json:"tenant_id" db:"tenant_id"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Status      WorkflowStatus  `json:"status" db:"status"`
	Trigger     WorkflowTrigger `json:"trigger" db:"trigger"`
	Steps       []WorkflowStep  `json:"steps" db:"steps"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

type WorkflowTrigger struct {
	Type       string                 `json:"type"`       // event, schedule, webhook
	Identifier string                 `json:"identifier"` // unique trigger identifier
	Config     map[string]interface{} `json:"config"`
}

type StepType string

const (
	StepTypeEmail     StepType = "email"
	StepTypeSMS       StepType = "sms"
	StepTypePush      StepType = "push"
	StepTypeWebhook   StepType = "webhook"
	StepTypeDelay     StepType = "delay"
	StepTypeDigest    StepType = "digest"
	StepTypeCondition StepType = "condition"
)

type WorkflowStep struct {
	ID         string                 `json:"id"`
	Type       StepType               `json:"type"`
	Name       string                 `json:"name"`
	Config     map[string]interface{} `json:"config"`
	Conditions []Condition            `json:"conditions,omitempty"`
	NextSteps  []string               `json:"next_steps,omitempty"`
	Position   int                    `json:"position"`
	Enabled    bool                   `json:"enabled"`
}

type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, contains, in, not_in
	Value    interface{} `json:"value"`
}

// NewWorkflow creates a new workflow
func NewWorkflow(tenantID, name, description string) *Workflow {
	return &Workflow{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		Status:      WorkflowStatusDraft,
		Steps:       []WorkflowStep{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// AddStep adds a new step to the workflow
func (w *Workflow) AddStep(step WorkflowStep) {
	step.ID = uuid.New().String()
	step.Position = len(w.Steps) + 1
	step.Enabled = true
	w.Steps = append(w.Steps, step)
	w.UpdatedAt = time.Now()
}

// Activate sets the workflow status to active
func (w *Workflow) Activate() error {
	if len(w.Steps) == 0 {
		return ErrWorkflowNoSteps
	}
	w.Status = WorkflowStatusActive
	w.UpdatedAt = time.Now()
	return nil
}

// Pause sets the workflow status to paused
func (w *Workflow) Pause() {
	w.Status = WorkflowStatusPaused
	w.UpdatedAt = time.Now()
}

// IsActive returns true if the workflow is active
func (w *Workflow) IsActive() bool {
	return w.Status == WorkflowStatusActive
}
