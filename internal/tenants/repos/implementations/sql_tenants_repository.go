package repos

import (
    "context"
    "fmt"
    "strconv"

    "getnoti.com/internal/tenants/domain"
    "getnoti.com/internal/tenants/repos"
    "getnoti.com/pkg/db"
   "getnoti.com/pkg/credentials"
)



type sqlTenantRepository struct {
    db          db.Database
    credManager *credentials.Manager
}

func NewTenantRepository(db db.Database, credManager *credentials.Manager) repository.TenantsRepository {
    return &sqlTenantRepository{
        db:          db,
        credManager: credManager,
    }
}

func (r *sqlTenantRepository) CreateTenant(ctx context.Context, tenant domain.Tenant) error {
    // Begin transaction
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Insert tenant
    tenantQuery := `INSERT INTO tenants (id, name) VALUES (?, ?)`
    _, err = tx.Exec(ctx, tenantQuery, tenant.ID, tenant.Name)
    if err != nil {
        return fmt.Errorf("failed to insert tenant data: %w", err)
    }

    // Store DB credentials if provided
    if tenant.DBConfig != nil {
        err = r.storeDBCredentials(tenant.ID, tenant.DBConfig)
        if err != nil {
            return fmt.Errorf("failed to store DB credentials: %w", err)
        }
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}

func (r *sqlTenantRepository) GetTenantByID(ctx context.Context, tenantID string) (domain.Tenant, error) {
    var tenant domain.Tenant

    metadataQuery := `SELECT id, name FROM tenants WHERE id = ?`
    err := r.db.QueryRow(ctx, metadataQuery, tenantID).Scan(&tenant.ID, &tenant.Name)
    if err != nil {
        return domain.Tenant{}, fmt.Errorf("failed to get tenant metadata: %w", err)
    }

    // Fetch DB credentials if available
    dbConfig, err := r.getDBCredentials(tenantID)
    if err == nil {
        tenant.DBConfig = dbConfig
    }

    return tenant, nil
}

func (r *sqlTenantRepository) Update(ctx context.Context, tenant domain.Tenant) error {
    // Begin transaction
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Update tenant data
    query := `UPDATE tenants SET name = ? WHERE id = ?`
    _, err = tx.Exec(ctx, query, tenant.Name, tenant.ID)
    if err != nil {
        return fmt.Errorf("failed to update tenant data: %w", err)
    }

    // Update DB credentials if provided
    if tenant.DBConfig != nil {
        err = r.storeDBCredentials(tenant.ID, tenant.DBConfig)
        if err != nil {
            return fmt.Errorf("failed to update DB credentials: %w", err)
        }
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}

func (r *sqlTenantRepository) GetAllTenants(ctx context.Context) ([]domain.Tenant, error) {
    query := `SELECT id, name FROM tenants ORDER BY name`
    rows, err := r.db.Query(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to query tenants: %w", err)
    }
    defer rows.Close()

    tenants := []domain.Tenant{}
    for rows.Next() {
        var tenant domain.Tenant
        err := rows.Scan(&tenant.ID, &tenant.Name)
        if err != nil {
            return nil, fmt.Errorf("failed to scan tenant row: %w", err)
        }
        tenants = append(tenants, tenant)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating tenant rows: %w", err)
    }

    return tenants, nil
}

func (r *sqlTenantRepository) storeDBCredentials(tenantID string, dbConfig *domain.DBCredentials) error {
    data := map[string]interface{}{
        "type":     dbConfig.Type,
        "host":     dbConfig.Host,
        "port":     dbConfig.Port,
        "username": dbConfig.Username,
        "password": dbConfig.Password,
        "dbname":   dbConfig.DBName,
        "dsn":      dbConfig.DSN,
    }

    // Use credential manager instead of vault directly
    err := r.credManager.StoreTenantDatabaseCredentials(tenantID, data)
    if err != nil {
        return fmt.Errorf("failed to store DB credentials: %w", err)
    }

    return nil
}

func (r *sqlTenantRepository) getDBCredentials(tenantID string) (*domain.DBCredentials, error) {
    // Use credential manager instead of vault directly
    data, err := r.credManager.GetTenantDatabaseCredentials(tenantID)
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve DB credentials: %w", err)
    }

    dbConfig := &domain.DBCredentials{}

    // Type is required
    if typeStr, ok := data["type"].(string); ok {
        dbConfig.Type = typeStr
    } else {
        return nil, fmt.Errorf("DB credentials missing required 'type' field")
    }

    // DSN is optional but preferred
    if dsn, ok := data["dsn"].(string); ok {
        dbConfig.DSN = dsn
    }

    // Individual fields
    if host, ok := data["host"].(string); ok {
        dbConfig.Host = host
    }

    if portVal, ok := data["port"]; ok {
        switch v := portVal.(type) {
        case int:
            dbConfig.Port = v
        case float64:
            dbConfig.Port = int(v)
        case string:
            port, err := strconv.Atoi(v)
            if err == nil {
                dbConfig.Port = port
            }
        }
    }

    if username, ok := data["username"].(string); ok {
        dbConfig.Username = username
    }

    if password, ok := data["password"].(string); ok {
        dbConfig.Password = password
    }

    if dbname, ok := data["dbname"].(string); ok {
        dbConfig.DBName = dbname
    }

    return dbConfig, nil
}

