package repository

import (
	"context"
	"database/sql"
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert into providers table
	_, err = tx.Exec(ctx, "INSERT INTO providers (id, name) VALUES (?, ?)", provider.ID, provider.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to insert provider: %w", err)
	}

	// Insert channels
	for _, channel := range provider.Channels {
		_, err = tx.Exec(ctx, `
            INSERT INTO provider_channels (provider_id, channel_type, enabled, priority)
            VALUES (?, ?, ?, ?)
        `, provider.ID, channel.Type, channel.Enabled, channel.Priority)
		if err != nil {
			return nil, fmt.Errorf("failed to insert channel: %w", err)
		}
	}

	// TODO: Store credentials
	// Implement credential storage here. Consider using a secure storage solution like a vault.
	// Example:
	// if provider.Credentials != nil {
	//     err = storeCredentialsInVault(provider.ID, provider.Credentials)
	//     if err != nil {
	//         return nil, fmt.Errorf("failed to store credentials: %w", err)
	//     }
	// }

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return provider, nil
}

func (r *sqlProviderRepository) GetProviderByID(ctx context.Context, id string) (*domain.Provider, error) {
	provider := &domain.Provider{ID: id}

	err := r.db.QueryRow(ctx, "SELECT name FROM providers WHERE id = ?", id).Scan(&provider.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	rows, err := r.db.Query(ctx, `
        SELECT channel_type, enabled, priority
        FROM provider_channels
        WHERE provider_id = ?
        ORDER BY priority
    `, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider channels: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var channel domain.PrioritizedChannel
		err := rows.Scan(&channel.Type, &channel.Enabled, &channel.Priority)
		if err != nil {
			return nil, fmt.Errorf("failed to scan channel: %w", err)
		}
		provider.Channels = append(provider.Channels, channel)
	}

	return provider, nil
}

func (r *sqlProviderRepository) GetProvidersByChannel(ctx context.Context, channelName string) ([]*domain.Provider, error) {
	query := `
        SELECT p.id, p.name, pc.channel_type, pc.enabled, pc.priority
        FROM providers p
        JOIN provider_channels pc ON p.id = pc.provider_id
        WHERE pc.channel_type = ? AND pc.enabled = true
        ORDER BY pc.priority ASC
    `
	rows, err := r.db.Query(ctx, query, channelName)
	if err != nil {
		return nil, fmt.Errorf("failed to query providers by channel: %w", err)
	}
	defer rows.Close()

	var providers []*domain.Provider
	for rows.Next() {
		var provider domain.Provider
		var channel domain.PrioritizedChannel

		err := rows.Scan(
			&provider.ID, &provider.Name, &channel.Type, &channel.Enabled, &channel.Priority,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider row: %w", err)
		}

		provider.Channels = append(provider.Channels, channel)
		providers = append(providers, &provider)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over provider rows: %w", err)
	}

	if len(providers) == 0 {
		return nil, nil // No providers found for this channel
	}

	return providers, nil
}

func (r *sqlProviderRepository) GetProviderByChannel(ctx context.Context, channelName string) (*domain.Provider, error) {
	query := `
        SELECT p.id, p.name, pc.channel_type, pc.enabled, pc.priority
        FROM providers p
        JOIN provider_channels pc ON p.id = pc.provider_id
        WHERE pc.channel_type = ? AND pc.enabled = true
        ORDER BY pc.priority ASC
        LIMIT 1
    `
	var provider domain.Provider
	var channel domain.PrioritizedChannel

	err := r.db.QueryRow(ctx, query, channelName).Scan(
		&provider.ID, &provider.Name, &channel.Type, &channel.Enabled, &channel.Priority,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No provider found for this channel
		}
		return nil, fmt.Errorf("failed to get provider by channel: %w", err)
	}

	provider.Channels = []domain.PrioritizedChannel{channel}
	return &provider, nil
}

func (r *sqlProviderRepository) GetProviders(ctx context.Context) ([]*domain.Provider, error) {
	rows, err := r.db.Query(ctx, "SELECT id, name FROM providers")
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}
	defer rows.Close()

	var providers []*domain.Provider
	for rows.Next() {
		var provider domain.Provider
		err := rows.Scan(&provider.ID, &provider.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider: %w", err)
		}
		providers = append(providers, &provider)
	}

	for _, provider := range providers {
		channelRows, err := r.db.Query(ctx, `
            SELECT channel_type, enabled, priority
            FROM provider_channels
            WHERE provider_id = ?
            ORDER BY priority
        `, provider.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get provider channels: %w", err)
		}
		defer channelRows.Close()

		for channelRows.Next() {
			var channel domain.PrioritizedChannel
			err := channelRows.Scan(&channel.Type, &channel.Enabled, &channel.Priority)
			if err != nil {
				return nil, fmt.Errorf("failed to scan channel: %w", err)
			}
			provider.Channels = append(provider.Channels, channel)
		}
	}

	return providers, nil
}

func (r *sqlProviderRepository) UpdateProvider(ctx context.Context, provider *domain.Provider) (*domain.Provider, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update provider name
	_, err = tx.Exec(ctx, "UPDATE providers SET name = ? WHERE id = ?", provider.Name, provider.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update provider: %w", err)
	}

	// Delete existing channels
	_, err = tx.Exec(ctx, "DELETE FROM provider_channels WHERE provider_id = ?", provider.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing channels: %w", err)
	}

	// Insert updated channels
	for _, channel := range provider.Channels {
		_, err = tx.Exec(ctx, `
            INSERT INTO provider_channels (provider_id, channel_type, enabled, priority)
            VALUES (?, ?, ?, ?)
        `, provider.ID, channel.Type, channel.Enabled, channel.Priority)
		if err != nil {
			return nil, fmt.Errorf("failed to insert updated channel: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return provider, nil
}

func (r *sqlProviderRepository) GetNextAvailablePriority(ctx context.Context, channelName string) (int, error) {
	var maxPriority int
	err := r.db.QueryRow(ctx, `
        SELECT COALESCE(MAX(priority), 0)
        FROM provider_channels
        WHERE channel_type = ?
    `, channelName).Scan(&maxPriority)
	if err != nil {
		return 0, fmt.Errorf("failed to get max priority: %w", err)
	}
	return maxPriority + 1, nil
}
