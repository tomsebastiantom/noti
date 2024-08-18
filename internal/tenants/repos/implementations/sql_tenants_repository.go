package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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

	// Begin transaction on mainDB
	txMain, err := r.mainDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction on mainDB: %w", err)
	}
	defer txMain.Rollback()

	// Begin transaction on tenantDB
	txTenant, err := r.tenantDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction on tenantDB: %w", err)
	}
	defer txTenant.Rollback()

	// Insert into mainDB
	tenantQuery := `INSERT INTO tenants (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`
	_, err = txMain.Exec(ctx, tenantQuery, tenant.ID, tenant.Name, now, now)
	if err != nil {
		return fmt.Errorf("failed to insert tenant data: %w", err)
	}

	// Insert into tenantDB
	metadataQuery := `INSERT INTO tenant_metadata (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`
	_, err = txTenant.Exec(ctx, metadataQuery, tenant.ID, tenant.Name, now, now)
	if err != nil {
		return fmt.Errorf("failed to insert tenant metadata: %w", err)
	}

	// Store DB credentials in Vault
	if tenant.DBConfig != nil {
		err = r.storeDBCredentialsInVault(tenant.ID, tenant.DBConfig)
		if err != nil {
			return fmt.Errorf("failed to store DB credentials in Vault: %w", err)
		}
	}

	// Commit both transactions
	if err := txMain.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction on mainDB: %w", err)
	}
	if err := txTenant.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction on tenantDB: %w", err)
	}

	return nil
}

func (r *sqlTenantRepository) GetTenantByID(ctx context.Context, tenantID string) (domain.Tenant, error) {
	var tenant domain.Tenant

	metadataQuery := `SELECT id, name FROM tenants WHERE id = ?`
	err := r.mainDB.QueryRow(ctx, metadataQuery, tenantID).Scan(&tenant.ID, &tenant.Name)
	if err != nil {
		return domain.Tenant{}, fmt.Errorf("failed to get tenant metadata: %w", err)
	}

	// Fetch DB credentials from Vault
	dbConfig, err := r.getDBCredentialsFromVault(tenantID)
	if err != nil {
		return domain.Tenant{}, fmt.Errorf("failed to get DB credentials from Vault: %w", err)
	}
	tenant.DBConfig = dbConfig

	return tenant, nil
}

func (r *sqlTenantRepository) Update(ctx context.Context, tenant domain.Tenant) error {
	now := time.Now()

	// Begin transaction on mainDB
	txMain, err := r.mainDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction on mainDB: %w", err)
	}
	defer txMain.Rollback()

	// Begin transaction on tenantDB
	txTenant, err := r.tenantDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction on tenantDB: %w", err)
	}
	defer txTenant.Rollback()

	// Update tenant data in mainDB
	tenantQuery := `UPDATE tenants SET name = ?, updated_at = ? WHERE id = ?`
	_, err = txMain.Exec(ctx, tenantQuery, tenant.Name, now, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to update tenant data: %w", err)
	}

	// Update tenant metadata in tenantDB
	metadataQuery := `UPDATE tenant_metadata SET name = ?, updated_at = ? WHERE id = ?`
	_, err = txTenant.Exec(ctx, metadataQuery, tenant.Name, now, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to update tenant metadata: %w", err)
	}

	// Update DB credentials in Vault
	if tenant.DBConfig != nil {
		err = r.updateDBCredentialsInVault(tenant.ID, tenant.DBConfig)
		if err != nil {
			return fmt.Errorf("failed to update DB credentials in Vault: %w", err)
		}
	}

	// Commit transactions
	if err := txMain.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction on mainDB: %w", err)
	}
	if err := txTenant.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction on tenantDB: %w", err)
	}

	return nil
}

func (r *sqlTenantRepository) storeDBCredentialsInVault(tenantID string, dbConfig *domain.DBCredentials) error {

	data := map[string]interface{}{
		"type":     dbConfig.Type,
		"host":     dbConfig.Host,
		"port":     dbConfig.Port,
		"username": dbConfig.Username,
		"password": dbConfig.Password,
		"dbname":   dbConfig.DBName,
		"dsn":      dbConfig.DSN,
	}

	err := vault.CreateCredential(tenantID, vault.DBCredential, "", data)
	if err != nil {
		return fmt.Errorf("failed to store DB credentials in Vault for database %s", err)
	}

	return nil
}

func (r *sqlTenantRepository) updateDBCredentialsInVault(tenantID string, dbConfig *domain.DBCredentials) error {
	data := map[string]interface{}{
		"type":     dbConfig.Type,
		"host":     dbConfig.Host,
		"port":     dbConfig.Port,
		"username": dbConfig.Username,
		"password": dbConfig.Password,
		"dbname":   dbConfig.DBName,
		"dsn":      dbConfig.DSN,
	}

	err := vault.UpdateCredential(tenantID, vault.DBCredential, "", data)
	if err != nil {
		return fmt.Errorf("failed to update DB credentials in Vault for database %s: %w", dbConfig.DBName, err)
	}

	return nil
}

func (r *sqlTenantRepository) getDBCredentialsFromVault(tenantID string) (*domain.DBCredentials, error) {
	// Retrieve credentials from Vault
	dbConfigData, err := vault.GetClientCredentials(tenantID, vault.DBCredential, "")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve DB credentials from Vault: %w", err)
	}

	// Check if credentials map is empty
	if len(dbConfigData) == 0 {
		return nil, fmt.Errorf("no DB credentials found in Vault for tenant %s", tenantID)
	}
	portStr := dbConfigData["port"].(json.Number).String()
	port, err := strconv.Atoi(portStr)

	if err != nil {
		return nil, fmt.Errorf("port error for tenant %s", tenantID)
	}
	// Extract and assign credential data
	dbConfig := &domain.DBCredentials{
		Type:     dbConfigData["type"].(string),
		Host:     dbConfigData["host"].(string),
		Port:     port,
		Username: dbConfigData["username"].(string),
		Password: dbConfigData["password"].(string),
		DBName:   dbConfigData["dbname"].(string),
		DSN:      dbConfigData["dsn"].(string),
	}

	// Return the credentials
	return dbConfig, nil
}

type sqlTenantsRepository struct {
	mainDB db.Database
}

func NewTenantsRepository(mainDB db.Database) repository.TenantsRepository {
	return &sqlTenantsRepository{mainDB: mainDB}
}

func (r *sqlTenantsRepository) GetAllTenants(ctx context.Context) ([]domain.Tenant, error) {
	query := `SELECT id, name FROM tenants`
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
