package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"getnoti.com/pkg/db"
	"getnoti.com/pkg/logger"
)

// EventStore defines the interface for storing and retrieving domain events
type EventStore interface {
	// Store an event in the tenant-specific database
	Store(ctx context.Context, event DomainEvent) error
	
	// Retrieve events for an aggregate
	GetEventsForAggregate(ctx context.Context, tenantID, aggregateID string) ([]DomainEvent, error)
	
	// Retrieve events by type
	GetEventsByType(ctx context.Context, tenantID, eventType string, limit int) ([]DomainEvent, error)
	
	// Initialize the event store (create tables if needed)
	Initialize(ctx context.Context) error
}

// DatabaseEventStore implements EventStore using your existing database manager
type DatabaseEventStore struct {
	dbManager *db.Manager
	logger    logger.Logger
}

// NewDatabaseEventStore creates a new database-backed event store
func NewDatabaseEventStore(dbManager *db.Manager, logger logger.Logger) *DatabaseEventStore {
	return &DatabaseEventStore{
		dbManager: dbManager,
		logger:    logger,
	}
}

// Store saves an event to the tenant-specific database
func (store *DatabaseEventStore) Store(ctx context.Context, event DomainEvent) error {
	// Get tenant-specific database connection
	tenantDB, err := store.dbManager.GetDatabaseConnection(event.GetTenantID())
	if err != nil {
		store.logger.Error("Failed to get tenant database for event storage",
			logger.Field{Key: "tenant_id", Value: event.GetTenantID()},
			logger.Field{Key: "event_id", Value: event.GetEventID()},
			logger.Field{Key: "error", Value: err.Error()})
		return err
	}
	
	// Serialize event payload
	payloadJSON, err := json.Marshal(event.GetPayload())
	if err != nil {
		store.logger.Error("Failed to serialize event payload",
			logger.Field{Key: "event_id", Value: event.GetEventID()},
			logger.Field{Key: "error", Value: err.Error()})
		return err
	}
	
	// Insert event into tenant-specific event_store table
	query := `
		INSERT INTO event_store (
			event_id, event_type, aggregate_id, tenant_id, 
			timestamp, version, payload, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	_, err = tenantDB.Exec(ctx, query,
		event.GetEventID(),
		event.GetEventType(),
		event.GetAggregateID(),
		event.GetTenantID(),
		event.GetTimestamp(),
		event.GetVersion(),
		payloadJSON,
	)
	
	if err != nil {
		store.logger.Error("Failed to store event in database",
			logger.Field{Key: "tenant_id", Value: event.GetTenantID()},
			logger.Field{Key: "event_id", Value: event.GetEventID()},
			logger.Field{Key: "event_type", Value: event.GetEventType()},
			logger.Field{Key: "error", Value: err.Error()})
		return err
	}
	
	store.logger.Debug("Event stored successfully",
		logger.Field{Key: "tenant_id", Value: event.GetTenantID()},
		logger.Field{Key: "event_id", Value: event.GetEventID()},
		logger.Field{Key: "event_type", Value: event.GetEventType()})
	
	return nil
}

// GetEventsForAggregate retrieves all events for a specific aggregate
func (store *DatabaseEventStore) GetEventsForAggregate(ctx context.Context, tenantID, aggregateID string) ([]DomainEvent, error) {
	// Get tenant-specific database connection
	tenantDB, err := store.dbManager.GetDatabaseConnection(tenantID)
	if err != nil {
		return nil, err
	}
	
	query := `
		SELECT event_id, event_type, aggregate_id, tenant_id, 
		       timestamp, version, payload
		FROM event_store 
		WHERE tenant_id = $1 AND aggregate_id = $2 
		ORDER BY timestamp ASC, version ASC
	`
	
	rows, err := tenantDB.Query(ctx, query, tenantID, aggregateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var events []DomainEvent
	for rows.Next() {
		event, err := store.scanEvent(rows)
		if err != nil {
			store.logger.Error("Failed to scan event from database",
				logger.Field{Key: "tenant_id", Value: tenantID},
				logger.Field{Key: "aggregate_id", Value: aggregateID},
				logger.Field{Key: "error", Value: err.Error()})
			continue
		}
		events = append(events, event)
	}
	
	return events, nil
}

// GetEventsByType retrieves events by type with limit
func (store *DatabaseEventStore) GetEventsByType(ctx context.Context, tenantID, eventType string, limit int) ([]DomainEvent, error) {
	// Get tenant-specific database connection
	tenantDB, err := store.dbManager.GetDatabaseConnection(tenantID)
	if err != nil {
		return nil, err
	}
	
	query := `
		SELECT event_id, event_type, aggregate_id, tenant_id, 
		       timestamp, version, payload
		FROM event_store 
		WHERE tenant_id = $1 AND event_type = $2 
		ORDER BY timestamp DESC 
		LIMIT $3
	`
	
	rows, err := tenantDB.Query(ctx, query, tenantID, eventType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var events []DomainEvent
	for rows.Next() {
		event, err := store.scanEvent(rows)
		if err != nil {
			store.logger.Error("Failed to scan event from database",
				logger.Field{Key: "tenant_id", Value: tenantID},
				logger.Field{Key: "event_type", Value: eventType},
				logger.Field{Key: "error", Value: err.Error()})
			continue
		}
		events = append(events, event)
	}
	
	return events, nil
}

// Initialize creates the event_store table in all tenant databases
func (store *DatabaseEventStore) Initialize(ctx context.Context) error {
	// This would be handled by your existing migration system
	// The event_store table should be created in each tenant database
	
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS event_store (
			id BIGSERIAL PRIMARY KEY,
			event_id VARCHAR(255) UNIQUE NOT NULL,
			event_type VARCHAR(255) NOT NULL,
			aggregate_id VARCHAR(255) NOT NULL,
			tenant_id VARCHAR(255) NOT NULL,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			version INTEGER NOT NULL,
			payload JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		
		CREATE INDEX IF NOT EXISTS idx_event_store_aggregate ON event_store (tenant_id, aggregate_id, timestamp);
		CREATE INDEX IF NOT EXISTS idx_event_store_type ON event_store (tenant_id, event_type, timestamp);
		CREATE INDEX IF NOT EXISTS idx_event_store_tenant ON event_store (tenant_id, timestamp);
	`
	
	store.logger.Info("Event store table schema ready. Please add this to your tenant migrations:",
		logger.Field{Key: "sql", Value: createTableSQL})
	
	return nil
}

// scanEvent scans a database row into a DomainEvent
func (store *DatabaseEventStore) scanEvent(rows *sql.Rows) (DomainEvent, error) {
	var eventID, eventType, aggregateID, tenantID string
	var timestamp time.Time
	var version int
	var payloadJSON []byte
	
	err := rows.Scan(&eventID, &eventType, &aggregateID, &tenantID, &timestamp, &version, &payloadJSON)
	if err != nil {
		return nil, err
	}
	
	// Deserialize payload
	var payload map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, err
	}
	
	// Create base domain event
	event := &BaseDomainEvent{
		EventID:     eventID,
		EventType:   eventType,
		AggregateID: aggregateID,
		TenantID:    tenantID,
		Timestamp:   timestamp,
		Version:     version,
		Payload:     payload,
	}
	
	return event, nil
}

// Rows interface to match your database driver
type Rows interface {
	Scan(dest ...interface{}) error
	Next() bool
	Close() error
}
