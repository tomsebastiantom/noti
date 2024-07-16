package repository

import (
	"context"
	"fmt"

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
	query := `INSERT INTO providers (id, name, channels, tenant_id) 
              VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(ctx, query, provider.ID, provider.Name, provider.Channels, provider.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}
	return provider, nil
}

// GetProviderByID retrieves a provider by its ID
func (r *sqlProviderRepository) GetProviderByID(ctx context.Context, id string) (*domain.Provider, error) {
	query := `SELECT id, name, channels, tenant_id FROM providers WHERE id = ?`
	row := r.db.QueryRow(ctx, query, id)
	provider := &domain.Provider{}
	err := row.Scan(&provider.ID, &provider.Name, &provider.Channels, &provider.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}
	return provider, nil
}

// GetProvidersByTenantID retrieves all providers for a given tenant ID
func (r *sqlProviderRepository) GetProvidersByTenantID(ctx context.Context, tenantID string) ([]*domain.Provider, error) {
	query := `SELECT id, name, channels, tenant_id FROM providers WHERE tenant_id = ?`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}
	defer rows.Close()

	var providers []*domain.Provider
	for rows.Next() {
		provider := &domain.Provider{}
		err := rows.Scan(&provider.ID, &provider.Name, &provider.Channels, &provider.TenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider: %w", err)
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

// UpdateProvider updates an existing provider in the database
func (r *sqlProviderRepository) UpdateProvider(ctx context.Context, provider *domain.Provider) (*domain.Provider, error) {
	query := `UPDATE providers SET name = ?, channels = ?, tenant_id = ? WHERE id = ?`
	_, err := r.db.Exec(ctx, query, provider.Name, provider.Channels, provider.TenantID, provider.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update provider: %w", err)
	}
	return provider, nil
}
