package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"getnoti.com/internal/tenants/domain"
	repository "getnoti.com/internal/tenants/repos"
	"getnoti.com/pkg/db"
)

type sqlUserPreferenceRepository struct {
	db db.Database
}

func NewUserPreferenceRepository(db db.Database) repository.UserPreferenceRepository {
	return &sqlUserPreferenceRepository{db: db}
}

func (r *sqlUserPreferenceRepository) CreateUserPreference(ctx context.Context, preference domain.UserPreference) error {
	now := time.Now()
	preference.CreatedAt = now
	preference.UpdatedAt = now
	
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
	
	// Execute insert query
	query := `INSERT INTO user_preferences (id, user_id, enabled, channel_preferences, category_preferences, digest_settings, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err = r.db.Exec(ctx, query, preference.ID, preference.UserID, preference.Enabled, 
		channelPrefs, categoryPrefs, digestSettings, preference.CreatedAt, preference.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create user preference: %w", err)
	}
	
	return nil
}

func (r *sqlUserPreferenceRepository) GetUserPreferenceByUserID(ctx context.Context, userID string) (domain.UserPreference, error) {
	query := `SELECT id, user_id, enabled, channel_preferences, category_preferences, digest_settings, created_at, updated_at
              FROM user_preferences WHERE user_id = ?`
	
	row := r.db.QueryRow(ctx, query, userID)
	
	var preference domain.UserPreference
	var channelPrefs, categoryPrefs, digestSettings []byte
	
	err := row.Scan(
		&preference.ID,
		&preference.UserID,
		&preference.Enabled,
		&channelPrefs,
		&categoryPrefs,
		&digestSettings,
		&preference.CreatedAt,
		&preference.UpdatedAt,
	)
	
	if err != nil {
		return domain.UserPreference{}, fmt.Errorf("failed to get user preference: %w", err)
	}
	
	// Unmarshal JSON fields
	if err := json.Unmarshal(channelPrefs, &preference.ChannelPrefs); err != nil {
		return domain.UserPreference{}, fmt.Errorf("failed to unmarshal channel preferences: %w", err)
	}
	
	if err := json.Unmarshal(categoryPrefs, &preference.CategoryPrefs); err != nil {
		return domain.UserPreference{}, fmt.Errorf("failed to unmarshal category preferences: %w", err)
	}
	
	if err := json.Unmarshal(digestSettings, &preference.DigestSettings); err != nil {
		return domain.UserPreference{}, fmt.Errorf("failed to unmarshal digest settings: %w", err)
	}
	
	return preference, nil
}

func (r *sqlUserPreferenceRepository) UpdateUserPreference(ctx context.Context, preference domain.UserPreference) error {
	preference.UpdatedAt = time.Now()
	
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
	
	// Execute update query
	query := `UPDATE user_preferences SET 
              enabled = ?, 
              channel_preferences = ?, 
              category_preferences = ?, 
              digest_settings = ?, 
              updated_at = ?
              WHERE user_id = ?`
	
	_, err = r.db.Exec(ctx, query, 
		preference.Enabled, channelPrefs, categoryPrefs, 
		digestSettings, preference.UpdatedAt, preference.UserID)
	
	if err != nil {
		return fmt.Errorf("failed to update user preference: %w", err)
	}
	
	return nil
}

func (r *sqlUserPreferenceRepository) GetUserPreferencesByCategory(ctx context.Context, category string) ([]domain.UserPreference, error) {
	// This is more complex as we need to query JSONB data
	// The query checks if the category exists within the category_preferences field
	query := `SELECT id, user_id, enabled, channel_preferences, category_preferences, digest_settings, created_at, updated_at
              FROM user_preferences 
              WHERE category_preferences ? ?`
	
	rows, err := r.db.Query(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to query user preferences by category: %w", err)
	}
	defer rows.Close()
	
	var preferences []domain.UserPreference
	
	for rows.Next() {
		var preference domain.UserPreference
		var channelPrefs, categoryPrefs, digestSettings []byte
		
		err := rows.Scan(
			&preference.ID,
			&preference.UserID,
			&preference.Enabled,
			&channelPrefs,
			&categoryPrefs,
			&digestSettings,
			&preference.CreatedAt,
			&preference.UpdatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan user preference row: %w", err)
		}
		
		// Unmarshal JSON fields
		if err := json.Unmarshal(channelPrefs, &preference.ChannelPrefs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal channel preferences: %w", err)
		}
		
		if err := json.Unmarshal(categoryPrefs, &preference.CategoryPrefs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal category preferences: %w", err)
		}
		
		if err := json.Unmarshal(digestSettings, &preference.DigestSettings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal digest settings: %w", err)
		}
		
		preferences = append(preferences, preference)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user preference rows: %w", err)
	}
	
	return preferences, nil
}

func (r *sqlUserPreferenceRepository) GetUsersForDigest(ctx context.Context, digestType domain.DigestType, dayOfWeek, hour int) ([]domain.UserPreference, error) {
	// This query finds users with matching digest settings
	query := `SELECT id, user_id, enabled, channel_preferences, category_preferences, digest_settings, created_at, updated_at
              FROM user_preferences 
              WHERE enabled = true 
              AND digest_settings->>'enabled' = 'true'
              AND digest_settings->>'type' = ?`
	
	var args []interface{}
	args = append(args, string(digestType))
	
	// Add additional conditions based on digest type
	switch digestType {
	case domain.DigestTypeDaily:
		query += ` AND (digest_settings->>'deliveryHour')::int = ?`
		args = append(args, hour)
	case domain.DigestTypeWeekly:
		query += ` AND (digest_settings->>'preferredDayOfWeek')::int = ? AND (digest_settings->>'deliveryHour')::int = ?`
		args = append(args, dayOfWeek, hour)
	}
	
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users for digest: %w", err)
	}
	defer rows.Close()
	
	var preferences []domain.UserPreference
	
	for rows.Next() {
		var preference domain.UserPreference
		var channelPrefs, categoryPrefs, digestSettings []byte
		
		err := rows.Scan(
			&preference.ID,
			&preference.UserID,
			&preference.Enabled,
			&channelPrefs,
			&categoryPrefs,
			&digestSettings,
			&preference.CreatedAt,
			&preference.UpdatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan user preference row: %w", err)
		}
		
		// Unmarshal JSON fields
		if err := json.Unmarshal(channelPrefs, &preference.ChannelPrefs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal channel preferences: %w", err)
		}
		
		if err := json.Unmarshal(categoryPrefs, &preference.CategoryPrefs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal category preferences: %w", err)
		}
		
		if err := json.Unmarshal(digestSettings, &preference.DigestSettings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal digest settings: %w", err)
		}
		
		preferences = append(preferences, preference)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user preference rows: %w", err)
	}
	
	return preferences, nil
}
