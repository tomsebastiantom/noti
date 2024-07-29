package repos

import (
	"context"

	"fmt"
	"time"

	"getnoti.com/internal/tenants/domain"
	repository "getnoti.com/internal/tenants/repos"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/vault"
)

type sqlTenantRepository struct {
	mainDB   db.Database
	tenantDB db.Database
}

func NewTenantRepository(mainDB, tenantDB db.Database) repository.TenantRepository {
	return &sqlTenantRepository{mainDB: mainDB, tenantDB: tenantDB}
}

func (r *sqlTenantRepository) CreateTenant(ctx context.Context, tenant domain.Tenant) error {
	now := time.Now()

	tx, err := r.mainDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	metadataQuery := `INSERT INTO tenant_metadata (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`
	_, err = tx.Exec(ctx, metadataQuery, tenant.ID, tenant.Name, now, now)
	if err != nil {
		return fmt.Errorf("failed to insert tenant metadata: %w", err)
	}

	tenantQuery := `INSERT INTO tenants (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`
	_, err = r.tenantDB.Exec(ctx, tenantQuery, tenant.ID, tenant.Name, now, now)
	if err != nil {
		return fmt.Errorf("failed to insert tenant data: %w", err)
	}

	// Store DB credentials in Vault
	if len(tenant.DBConfigs) > 0 {
		err = r.storeDBCredentialsInVault(tenant.ID, tenant.DBConfigs)
		if err != nil {
			return fmt.Errorf("failed to store DB credentials in Vault: %w", err)
		}
	}

	return tx.Commit()
}

func (r *sqlTenantRepository) GetTenantByID(ctx context.Context, tenantID string) (domain.Tenant, error) {
	var tenant domain.Tenant

	metadataQuery := `SELECT id, name FROM tenant_metadata WHERE id = ?`
	err := r.mainDB.QueryRow(ctx, metadataQuery, tenantID).Scan(&tenant.ID, &tenant.Name)
	if err != nil {
		return domain.Tenant{}, fmt.Errorf("failed to get tenant metadata: %w", err)
	}

	// Fetch DB credentials from Vault
	dbConfigs, err := r.getDBCredentialsFromVault(tenantID)
	if err != nil {
		return domain.Tenant{}, fmt.Errorf("failed to get DB credentials from Vault: %w", err)
	}
	tenant.DBConfigs = dbConfigs

	return tenant, nil
}

func (r *sqlTenantRepository) Update(ctx context.Context, tenant domain.Tenant) error {
	now := time.Now()

	tx, err := r.mainDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	metadataQuery := `UPDATE tenant_metadata SET name = ?, updated_at = ? WHERE id = ?`
	_, err = tx.Exec(ctx, metadataQuery, tenant.Name, now, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to update tenant metadata: %w", err)
	}

	tenantQuery := `UPDATE tenants SET name = ?, updated_at = ? WHERE id = ?`
	_, err = r.tenantDB.Exec(ctx, tenantQuery, tenant.Name, now, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to update tenant data: %w", err)
	}

	// Update DB credentials in Vault
	if len(tenant.DBConfigs) > 0 {
		err = r.updateDBCredentialsInVault(tenant.ID, tenant.DBConfigs)
		if err != nil {
			return fmt.Errorf("failed to update DB credentials in Vault: %w", err)
		}
	}

	return tx.Commit()
}

func (r *sqlTenantRepository) storeDBCredentialsInVault(tenantID string, dbConfigs map[string]*domain.DBCredentials) error {
	err := vault.RefreshToken(tenantID)
	if err != nil {
		return fmt.Errorf("failed to refresh Vault token: %w", err)
	}

	for dbName, dbConfig := range dbConfigs {
		data := map[string]interface{}{
			"type":     dbConfig.Type,
			"host":     dbConfig.Host,
			"port":     dbConfig.Port,
			"username": dbConfig.Username,
			"password": dbConfig.Password,
			"dbname":   dbConfig.DBName,
			"dsn":      dbConfig.DSN,
		}

		err = vault.CreateCredential(tenantID, vault.DBCredential, dbName, data)
		if err != nil {
			return fmt.Errorf("failed to store DB credentials in Vault for database %s: %w", dbName, err)
		}
	}

	return nil
}

func (r *sqlTenantRepository) updateDBCredentialsInVault(tenantID string, dbConfigs map[string]*domain.DBCredentials) error {
	err := vault.RefreshToken(tenantID)
	if err != nil {
		return fmt.Errorf("failed to refresh Vault token: %w", err)
	}

	existingCredentials, err := vault.GetClientCredentials(tenantID, vault.DBCredential, "")
	if err != nil {
		return fmt.Errorf("failed to get existing DB credentials from Vault: %w", err)
	}

	for dbName, dbConfig := range dbConfigs {
		data := map[string]interface{}{
			"type":     dbConfig.Type,
			"host":     dbConfig.Host,
			"port":     dbConfig.Port,
			"username": dbConfig.Username,
			"password": dbConfig.Password,
			"dbname":   dbConfig.DBName,
			"dsn":      dbConfig.DSN,
		}

		if _, exists := existingCredentials[dbName]; exists {
			err = vault.UpdateCredential(tenantID, vault.DBCredential, dbName, data)
		} else {
			err = vault.CreateCredential(tenantID, vault.DBCredential, dbName, data)
		}

		if err != nil {
			return fmt.Errorf("failed to update/create DB credentials in Vault for database %s: %w", dbName, err)
		}
	}

	for existingDbName := range existingCredentials {
		if _, exists := dbConfigs[existingDbName]; !exists {
			err = vault.UpdateCredential(tenantID, vault.DBCredential, existingDbName, nil)
			if err != nil {
				return fmt.Errorf("failed to delete DB credentials from Vault for database %s: %w", existingDbName, err)
			}
		}
	}

	return nil
}

func (r *sqlTenantRepository) getDBCredentialsFromVault(tenantID string) (map[string]*domain.DBCredentials, error) {
	credentials, err := vault.GetClientCredentials(tenantID, vault.DBCredential, "")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve DB credentials from Vault: %w", err)
	}

	dbConfigs := make(map[string]*domain.DBCredentials)
	for dbName, cred := range credentials {
		dbConfigData, ok := cred.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid DB config data in Vault for database %s", dbName)
		}

		dbConfig := &domain.DBCredentials{
			Type:     dbConfigData["type"].(string),
			Host:     dbConfigData["host"].(string),
			Port:     int(dbConfigData["port"].(float64)),
			Username: dbConfigData["username"].(string),
			Password: dbConfigData["password"].(string),
			DBName:   dbConfigData["dbname"].(string),
			DSN:      dbConfigData["dsn"].(string),
		}

		dbConfigs[dbName] = dbConfig
	}

	return dbConfigs, nil
}

type sqlTenantsRepository struct {
	mainDB db.Database
}

func NewTenantsRepository(mainDB db.Database) repository.TenantsRepository {
	return &sqlTenantsRepository{mainDB: mainDB}
}

func (r *sqlTenantsRepository) GetAllTenants(ctx context.Context) ([]domain.Tenant, error) {
	query := `SELECT id, name FROM tenant_metadata`
	rows, err := r.mainDB.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenant metadata: %w", err)
	}
	defer rows.Close()

	var tenants []domain.Tenant
	for rows.Next() {
		var tenant domain.Tenant
		err := rows.Scan(&tenant.ID, &tenant.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant metadata: %w", err)
		}
		tenants = append(tenants, tenant)
	}

	return tenants, nil
}
