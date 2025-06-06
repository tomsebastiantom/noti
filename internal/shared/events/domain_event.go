package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event that occurred within the system
type DomainEvent interface {
	// Event identification
	GetEventID() string
	GetEventType() string
	GetAggregateID() string
	GetTenantID() string
	
	// Event metadata
	GetTimestamp() time.Time
	GetVersion() int
	
	// Event data
	GetPayload() interface{}
	ToJSON() ([]byte, error)
}

// BaseDomainEvent provides common implementation for domain events
type BaseDomainEvent struct {
	EventID     string                 `json:"event_id"`
	EventType   string                 `json:"event_type"`
	AggregateID string                 `json:"aggregate_id"`
	TenantID    string                 `json:"tenant_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     int                    `json:"version"`
	Payload     map[string]interface{} `json:"payload"`
}

// NewBaseDomainEvent creates a new base domain event
func NewBaseDomainEvent(eventType, aggregateID, tenantID string, payload map[string]interface{}) *BaseDomainEvent {
	return &BaseDomainEvent{
		EventID:     uuid.New().String(),
		EventType:   eventType,
		AggregateID: aggregateID,
		TenantID:    tenantID,
		Timestamp:   time.Now(),
		Version:     1,
		Payload:     payload,
	}
}

// Implement DomainEvent interface
func (e *BaseDomainEvent) GetEventID() string     { return e.EventID }
func (e *BaseDomainEvent) GetEventType() string  { return e.EventType }
func (e *BaseDomainEvent) GetAggregateID() string { return e.AggregateID }
func (e *BaseDomainEvent) GetTenantID() string   { return e.TenantID }
func (e *BaseDomainEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *BaseDomainEvent) GetVersion() int       { return e.Version }
func (e *BaseDomainEvent) GetPayload() interface{} { return e.Payload }

func (e *BaseDomainEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
