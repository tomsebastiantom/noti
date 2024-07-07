package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"getnoti.com/internal/templates/domain"
	"getnoti.com/pkg/db"
	"getnoti.com/internal/templates/repos"
)


type PostgresTemplateRepository struct {
	db db.Database
}

// NewPostgresTemplateRepository creates a new instance of PostgresTemplateRepository
func NewPostgresTemplateRepository(db db.Database) repos.TemplateRepository {
	return &PostgresTemplateRepository{db: db}
}

// CreateTemplate inserts a new template into the database
func (r *PostgresTemplateRepository) CreateTemplate(ctx context.Context, tmpl *domain.Template) error {
	query := `INSERT INTO templates (id, tenant_id, name, content, is_public, variables) VALUES (\$1, \$2, \$3, \$4, \$5, \$6)`
	_, err := r.db.Exec(ctx, query, tmpl.ID, tmpl.TenantID, tmpl.Name, tmpl.Content, tmpl.IsPublic, tmpl.Variables)
	if err != nil {
		return wrapError(err, ErrTemplateCreateFailed)
	}
	return nil
}

// GetTemplateByID retrieves a template by its ID
func (r *PostgresTemplateRepository) GetTemplateByID(ctx context.Context, templateID string) (*domain.Template, error) {
	query := `SELECT id, tenant_id, name, content, is_public, variables FROM templates WHERE id=\$1`
	row := r.db.QueryRow(ctx, query, templateID)
	tmpl := &domain.Template{}
	var variables []byte
	err := row.Scan(&tmpl.ID, &tmpl.TenantID, &tmpl.Name, &tmpl.Content, &tmpl.IsPublic, &variables)
	if err != nil {
		if errors.Is(err, db.ErrNoRows) {
			return nil, ErrTemplateNotFound
		}
		return nil, wrapError(err, ErrUnexpected)
	}

	err = json.Unmarshal(variables, &tmpl.Variables)
	if err != nil {
		return nil, wrapError(err, ErrUnexpected)
	}

	return tmpl, nil
}

// UpdateTemplate updates an existing template in the database
func (r *PostgresTemplateRepository) UpdateTemplate(ctx context.Context, tmpl *domain.Template) error {
	query := `UPDATE templates SET tenant_id=\$1, name=\$2, content=\$3, is_public=\$4, variables=\$5 WHERE id=\$6`
	_, err := r.db.Exec(ctx, query, tmpl.TenantID, tmpl.Name, tmpl.Content, tmpl.IsPublic, tmpl.Variables, tmpl.ID)
	if err != nil {
		return wrapError(err, ErrTemplateUpdateFailed)
	}
	return nil
}

// GetTemplatesByTenantID retrieves templates by tenant ID
func (r *PostgresTemplateRepository) GetTemplatesByTenantID(ctx context.Context, tenantID string) ([]domain.Template, error) {
	query := `SELECT id, tenant_id, name, content, is_public, variables FROM templates WHERE tenant_id=\$1`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, wrapError(err, ErrUnexpected)
	}
	defer rows.Close()

	var templates []domain.Template
	for rows.Next() {
		var tmpl domain.Template
		var variables []byte
		if err := rows.Scan(&tmpl.ID, &tmpl.TenantID, &tmpl.Name, &tmpl.Content, &tmpl.IsPublic, &variables); err != nil {
			return nil, wrapError(err, ErrUnexpected)
		}
		if err := json.Unmarshal(variables, &tmpl.Variables); err != nil {
			return nil, wrapError(err, ErrUnexpected)
		}
		templates = append(templates, tmpl)
	}

	if err := rows.Err(); err != nil {
		return nil, wrapError(err, ErrUnexpected)
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
