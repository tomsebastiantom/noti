package repository

import (
    "context"
    "fmt"
    "encoding/json"

    "getnoti.com/internal/providers/domain"
    repository "getnoti.com/internal/providers/repos"
    "getnoti.com/pkg/db"
)

type sqlProviderRepository struct {
    db db.Database
}

func NewProviderRepository(db db.Database) repository.ProviderRepository {
    return &sqlProviderRepository{db: db}
}

// CreateProvider inserts a new provider into the database
func (r *sqlProviderRepository) CreateProvider(ctx context.Context, provider *domain.Provider) (*domain.Provider, error) {
    query := `INSERT INTO providers (id, name, channels, enabled) 
              VALUES (?, ?, ?, ?)`
    channelsJSON, err := json.Marshal(provider.Channels)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal channels: %w", err)
    }
    _, err = r.db.Exec(ctx, query, provider.ID, provider.Name, channelsJSON, provider.Enabled)
    if err != nil {
        return nil, fmt.Errorf("failed to create provider: %w", err)
    }
    return provider, nil
}

// GetProviderByID retrieves a provider by its ID
func (r *sqlProviderRepository) GetProviderByID(ctx context.Context, id string) (*domain.Provider, error) {
    query := `SELECT id, name, channels, enabled FROM providers WHERE id = ?`
    row := r.db.QueryRow(ctx, query, id)
    provider := &domain.Provider{}
    var channelsJSON []byte
    err := row.Scan(&provider.ID, &provider.Name, &channelsJSON, &provider.Enabled)
    if err != nil {
        return nil, fmt.Errorf("failed to get provider: %w", err)
    }
    err = json.Unmarshal(channelsJSON, &provider.Channels)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal channels: %w", err)
    }
    return provider, nil
}

// GetProviders retrieves all providers
func (r *sqlProviderRepository) GetProviders(ctx context.Context) ([]*domain.Provider, error) {
    query := `SELECT id, name, channels, enabled FROM providers`
    rows, err := r.db.Query(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to get providers: %w", err)
    }
    defer rows.Close()

    var providers []*domain.Provider
    for rows.Next() {
        provider := &domain.Provider{}
        var channelsJSON []byte
        err := rows.Scan(&provider.ID, &provider.Name, &channelsJSON, &provider.Enabled)
        if err != nil {
            return nil, fmt.Errorf("failed to scan provider: %w", err)
        }
        err = json.Unmarshal(channelsJSON, &provider.Channels)
        if err != nil {
            return nil, fmt.Errorf("failed to unmarshal channels: %w", err)
        }
        providers = append(providers, provider)
    }
    return providers, nil
}

// UpdateProvider updates an existing provider in the database
func (r *sqlProviderRepository) UpdateProvider(ctx context.Context, provider *domain.Provider) (*domain.Provider, error) {
    query := `UPDATE providers SET name = ?, channels = ?, enabled = ? WHERE id = ?`
    channelsJSON, err := json.Marshal(provider.Channels)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal channels: %w", err)
    }
    _, err = r.db.Exec(ctx, query, provider.Name, channelsJSON, provider.Enabled, provider.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to update provider: %w", err)
    }
    return provider, nil
}

// GetNextAvailablePriority gets the next available priority for a given channel
func (r *sqlProviderRepository) GetNextAvailablePriority(ctx context.Context, channelName string) (int, error) {
    query := `SELECT COALESCE(MAX(JSON_EXTRACT(channels, CONCAT('$."', ?, '".priority'))), 0) FROM providers`
    var highestPriority int
    err := r.db.QueryRow(ctx, query, channelName).Scan(&highestPriority)
    if err != nil {
        return 0, fmt.Errorf("failed to get highest priority: %w", err)
    }
    return highestPriority + 1, nil
}
