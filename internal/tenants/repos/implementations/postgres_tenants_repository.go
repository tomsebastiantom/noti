package postgres

import (
    "context"
    "encoding/json"
    "getnoti.com/internal/tenants/domain"
    "getnoti.com/internal/tenants/repos"
    "getnoti.com/pkg/db"
    "time"
)

type postgresTenantRepository struct {
    db db.Database
}

func NewPostgresTenantRepository(db db.Database) repository.TenantRepository {
    return &postgresTenantRepository{db: db}
}

func (r *postgresTenantRepository) CreateTenant(ctx context.Context, tenant domain.Tenant) error {
    now := time.Now()
    preferences, err := json.Marshal(tenant.Preferences)
    if err != nil {
        return err
    }

    query := `INSERT INTO tenants (id, name, default_channel, preferences, created_at, updated_at) 
              VALUES (\$1, \$2, \$3, \$4, \$5, \$6)`
    _, err = r.db.Exec(ctx, query, tenant.ID, tenant.Name, preferences, now, now)
    return err
}

func (r *postgresTenantRepository) GetTenantByID(ctx context.Context, tenantid string) (domain.Tenant, error) {
    query := `SELECT id, name, default_channel, preferences, created_at, updated_at FROM tenants WHERE id = \$1`
    row := r.db.QueryRow(ctx, query, tenantid)
    var tenant domain.Tenant
    var preferences []byte
	err := row.Scan(&tenant.ID, &tenant.Name, &preferences)
    //err := row.Scan(&tenant.ID, &tenant.Name, &tenant.DefaultChannel, &preferences, &tenant.CreatedAt, &tenant.UpdatedAt)
    if err != nil {
        return domain.Tenant{}, err
    }

    err = json.Unmarshal(preferences, &tenant.Preferences)
    if err != nil {
        return domain.Tenant{}, err
    }

    return tenant, nil
}

func (r *postgresTenantRepository) Update(ctx context.Context, tenant domain.Tenant) error {
    now := time.Now()
    preferences, err := json.Marshal(tenant.Preferences)
    if err != nil {
        return err
    }

    query := `UPDATE tenants SET name = \$1, default_channel = \$2, preferences = \$3, updated_at = \$4 WHERE id = \$5`
    _, err = r.db.Exec(ctx, query, tenant.Name, preferences, now, tenant.ID)
    return err
}

func (r *postgresTenantRepository) GetAllTenants(ctx context.Context) ([]domain.Tenant, error) {
    query := `SELECT id, name, default_channel, preferences, created_at, updated_at FROM tenants`
    rows, err := r.db.Query(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tenants []domain.Tenant
    for rows.Next() {
        var tenant domain.Tenant
        var preferences []byte
        err := rows.Scan(&tenant.ID, &tenant.Name, &preferences)
		//err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.DefaultChannel, &preferences, &tenant.CreatedAt, &tenant.UpdatedAt)
        if err != nil {
            return nil, err
        }

        err = json.Unmarshal(preferences, &tenant.Preferences)
        if err != nil {
            return nil, err
        }

        tenants = append(tenants, tenant)
    }
    return tenants, nil
}

func (r *postgresTenantRepository) GetPreferenceByChannel(ctx context.Context, tenantID string, channel string) (map[string]string, error) {
    query := `SELECT preferences FROM tenants WHERE id = \$1`
    row := r.db.QueryRow(ctx, query, tenantID)
    var preferences []byte
    err := row.Scan(&preferences)
    if err != nil {
        return nil, err
    }

    var prefs map[string]map[string]string
    err = json.Unmarshal(preferences, &prefs)
    if err != nil {
        return nil, err
    }

    return prefs[channel], nil
}