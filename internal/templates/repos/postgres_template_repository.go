package repos

import (
	"context"
	"errors"
	"fmt"

	"getnoti.com/internal/templates/domain/template"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTemplateRepository struct {
	db *pgxpool.Pool
}

func NewPostgresTemplateRepository(db *pgxpool.Pool) TemplateRepository {
	return &PostgresTemplateRepository{db: db}
}

func (r *PostgresTemplateRepository) CreateTemplate(ctx context.Context, tmpl *template.Template) error {
	_, err := r.db.Exec(ctx, "INSERT INTO templates (id, tenant_id, name, content, is_public, variables) VALUES ($1, $2, $3, $4, $5, $6)",
		tmpl.ID, tmpl.TenantID, tmpl.Name, tmpl.Content, tmpl.IsPublic, tmpl.Variables)
	if err != nil {
		return wrapError(err, ErrTemplateCreateFailed)
	}
	return nil
}

func (r *PostgresTemplateRepository) GetTemplateByID(ctx context.Context, templateID string) (*template.Template, error) {
	var tmpl template.Template
	err := r.db.QueryRow(ctx, "SELECT id, tenant_id, name, content, is_public, variables FROM templates WHERE id=$1",
		templateID).Scan(&tmpl.ID, &tmpl.TenantID, &tmpl.Name, &tmpl.Content, &tmpl.IsPublic, &tmpl.Variables)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTemplateNotFound
		}
		return nil, wrapError(err, ErrUnexpected)
	}
	return &tmpl, nil
}

func (r *PostgresTemplateRepository) UpdateTemplate(ctx context.Context, tmpl *template.Template) error {
	_, err := r.db.Exec(ctx, "UPDATE templates SET tenant_id=$1, name=$2, content=$3, is_public=$4, variables=$5 WHERE id=$6",
		tmpl.TenantID, tmpl.Name, tmpl.Content, tmpl.IsPublic, tmpl.Variables, tmpl.ID)
	if err != nil {
		return wrapError(err, ErrTemplateUpdateFailed)
	}
	return nil
}

// wrapError wraps a pgx error with a more generic error
func wrapError(err error, genericErr error) error {
	if err != nil {
		return fmt.Errorf("%w: %v", genericErr, err)
	}
	return genericErr
}
