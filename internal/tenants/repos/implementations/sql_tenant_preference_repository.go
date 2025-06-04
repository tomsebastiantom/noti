package repos

import (
	"context"
	"encoding/json"
	"fmt"

	"getnoti.com/internal/tenants/domain"
	repository "getnoti.com/internal/tenants/repos"
	"getnoti.com/pkg/db"
)

type sqlTenantPreferenceRepository struct {
	db db.Database
}

// NewTenantPreferenceRepository creates a new SQL-based tenant preference repository
func NewTenantPreferenceRepository(db db.Database) repository.TenantPreferenceRepository {
	return &sqlTenantPreferenceRepository{db: db}
}

// CreateTenantPreference creates a new tenant preference record
func (r *sqlTenantPreferenceRepository) CreateTenantPreference(ctx context.Context, preference domain.TenantPreference) error {
	// Marshal JSON fields
	channelPrefs, err := json.Marshal(preference.ChannelPrefs)
	if err != nil {
		return fmt.Errorf("failed to marshal channel preferences: %w", err)
	}

	categoryPrefs, err := json.Marshal(preference.CategoryPrefs)
	if err != nil {
		return fmt.Errorf("failed to marshal category preferences: %w", err)
	}

	digestSettings, err := json.Marshal(preference.DigestSettings)
	if err != nil {
		return fmt.Errorf("failed to marshal digest settings: %w", err)
	}

	// Insert into the database
	query := `
		INSERT INTO tenant_preferences (id, tenant_id, enabled, channel_preferences, category_preferences, digest_settings)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.Exec(ctx, query,
		preference.ID,
		preference.TenantID,
		preference.Enabled,
		channelPrefs,
		categoryPrefs,
		digestSettings,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create tenant preference: %w", err)
	}

	return nil
}

// GetTenantPreferenceByTenantID gets tenant preferences by tenant ID
func (r *sqlTenantPreferenceRepository) GetTenantPreferenceByTenantID(ctx context.Context, tenantID string) (domain.TenantPreference, error) {
	var preference domain.TenantPreference
	var channelPrefsJSON, categoryPrefsJSON, digestSettingsJSON []byte

	query := `
		SELECT id, tenant_id, enabled, channel_preferences, category_preferences, digest_settings, created_at, updated_at
		FROM tenant_preferences
		WHERE tenant_id = ?
	`
	
	err := r.db.QueryRow(ctx, query, tenantID).Scan(
		&preference.ID,
		&preference.TenantID,
		&preference.Enabled,
		&channelPrefsJSON,
		&categoryPrefsJSON,
		&digestSettingsJSON,
		&preference.CreatedAt,
		&preference.UpdatedAt,
	)
	
	if err != nil {
		return domain.TenantPreference{}, fmt.Errorf("failed to get tenant preference: %w", err)
	}

	// Unmarshal JSON fields
	preference.ChannelPrefs = make(map[domain.ChannelType]bool)
	if err = json.Unmarshal(channelPrefsJSON, &preference.ChannelPrefs); err != nil {
		return domain.TenantPreference{}, fmt.Errorf("failed to unmarshal channel preferences: %w", err)
	}

	preference.CategoryPrefs = make(map[string]domain.CategoryPreference)
	if err = json.Unmarshal(categoryPrefsJSON, &preference.CategoryPrefs); err != nil {
		return domain.TenantPreference{}, fmt.Errorf("failed to unmarshal category preferences: %w", err)
	}

	if err = json.Unmarshal(digestSettingsJSON, &preference.DigestSettings); err != nil {
		return domain.TenantPreference{}, fmt.Errorf("failed to unmarshal digest settings: %w", err)
	}

	return preference, nil
}

// UpdateTenantPreference updates an existing tenant preference record
func (r *sqlTenantPreferenceRepository) UpdateTenantPreference(ctx context.Context, preference domain.TenantPreference) error {
	// Marshal JSON fields
	channelPrefs, err := json.Marshal(preference.ChannelPrefs)
	if err != nil {
		return fmt.Errorf("failed to marshal channel preferences: %w", err)
	}

	categoryPrefs, err := json.Marshal(preference.CategoryPrefs)
	if err != nil {
		return fmt.Errorf("failed to marshal category preferences: %w", err)
	}

	digestSettings, err := json.Marshal(preference.DigestSettings)
	if err != nil {
		return fmt.Errorf("failed to marshal digest settings: %w", err)
	}

	// Update in the database
	query := `
		UPDATE tenant_preferences
		SET enabled = ?, channel_preferences = ?, category_preferences = ?, digest_settings = ?
		WHERE tenant_id = ?
	`
	
	_, err = r.db.Exec(ctx, query,
		preference.Enabled,
		channelPrefs,
		categoryPrefs,
		digestSettings,
		preference.TenantID,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update tenant preference: %w", err)
	}

	return nil
}
