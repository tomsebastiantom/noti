package repos

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"getnoti.com/internal/templates/domain"
	"getnoti.com/internal/templates/repos"
	"getnoti.com/pkg/db"
)

type sqlTemplateRepository struct {
	db db.Database
}

// NewTemplateRepository creates a new instance of sqlTemplateRepository
func NewTemplateRepository(db db.Database) repos.TemplateRepository {
	return &sqlTemplateRepository{db: db}
}

// CreateTemplate inserts a new template into the database
func (r *sqlTemplateRepository) CreateTemplate(ctx context.Context, tmpl *domain.Template) error {
	query := `INSERT INTO templates (id, tenant_id, name, content, is_public, variables) VALUES (?, ?, ?, ?, ?, ?)`
	variables, err := json.Marshal(tmpl.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}
	_, err = r.db.Exec(ctx, query, tmpl.ID, tmpl.TenantID, tmpl.Name, tmpl.Content, tmpl.IsPublic, variables)
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}
	return nil
}

// GetTemplateByID retrieves a template by its ID
func (r *sqlTemplateRepository) GetTemplateByID(ctx context.Context, templateID string) (*domain.Template, error) {
	query := `SELECT id, tenant_id, name, content, is_public, variables FROM templates WHERE id = ?`
	row := r.db.QueryRow(ctx, query, templateID)
	tmpl := &domain.Template{}
	var variables []byte
	err := row.Scan(&tmpl.ID, &tmpl.TenantID, &tmpl.Name, &tmpl.Content, &tmpl.IsPublic, &variables)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	err = json.Unmarshal(variables, &tmpl.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
	}

	return tmpl, nil
}

// UpdateTemplate updates an existing template in the database
func (r *sqlTemplateRepository) UpdateTemplate(ctx context.Context, tmpl *domain.Template) error {
	query := `UPDATE templates SET tenant_id = ?, name = ?, content = ?, is_public = ?, variables = ? WHERE id = ?`
	variables, err := json.Marshal(tmpl.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}
	_, err = r.db.Exec(ctx, query, tmpl.TenantID, tmpl.Name, tmpl.Content, tmpl.IsPublic, variables, tmpl.ID)
	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}
	return nil
}

// GetTemplatesByTenantID retrieves templates by tenant ID
func (r *sqlTemplateRepository) GetTemplatesByTenantID(ctx context.Context, tenantID string) ([]domain.Template, error) {
	query := `SELECT id, tenant_id, name, content, is_public, variables FROM templates WHERE tenant_id = ?`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query templates: %w", err)
	}
	defer rows.Close()

	var templates []domain.Template
	for rows.Next() {
		var tmpl domain.Template
		var variables []byte
		err := rows.Scan(&tmpl.ID, &tmpl.TenantID, &tmpl.Name, &tmpl.Content, &tmpl.IsPublic, &variables)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}

		err = json.Unmarshal(variables, &tmpl.Variables)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}

		templates = append(templates, tmpl)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating template rows: %w", err)
	}
	return templates, nil
}

// wrapError wraps a database error with a more generic error
func wrapError(err error, genericErr error) error {
	if err != nil {
		return fmt.Errorf("%w: %v", genericErr, err)
	}
	return genericErr
}
