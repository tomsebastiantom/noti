// // package repos

// // import (
// // 	"context"
// // 	"encoding/json"
// // 	"fmt"
// // 	"strconv"
// // 	"time"

// // 	"getnoti.com/internal/tenants/domain"
// // 	repository "getnoti.com/internal/tenants/repos"
// // 	"getnoti.com/pkg/db"
// // 	"getnoti.com/pkg/vault"
// // )

// // type sqlTenantRepository struct {
// // 	mainDB   db.Database
// // 	tenantDB db.Database
// // }

// // func NewTenantRepository(mainDB, tenantDB db.Database) repository.TenantRepository {
// // 	return &sqlTenantRepository{mainDB: mainDB, tenantDB: tenantDB}
// // }

// // func (r *sqlTenantRepository) CreateTenant(ctx context.Context, tenant domain.Tenant) error {
// // 	now := time.Now()

// // 	// Begin transaction on mainDB
// // 	txMain, err := r.mainDB.BeginTx(ctx, nil)
// // 	if err != nil {
// // 		return fmt.Errorf("failed to begin transaction on mainDB: %w", err)
// // 	}
// // 	defer txMain.Rollback()

// // 	// Begin transaction on tenantDB
// // 	txTenant, err := r.tenantDB.BeginTx(ctx, nil)
// // 	if err != nil {
// // 		return fmt.Errorf("failed to begin transaction on tenantDB: %w", err)
// // 	}
// // 	defer txTenant.Rollback()

// // 	// Insert into mainDB
// // 	tenantQuery := `INSERT INTO tenants (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`
// // 	_, err = txMain.Exec(ctx, tenantQuery, tenant.ID, tenant.Name, now, now)
// // 	if err != nil {
// // 		return fmt.Errorf("failed to insert tenant data: %w", err)
// // 	}

// // 	// Insert into tenantDB
// // 	metadataQuery := `INSERT INTO tenant_metadata (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`
// // 	_, err = txTenant.Exec(ctx, metadataQuery, tenant.ID, tenant.Name, now, now)
// // 	if err != nil {
// // 		return fmt.Errorf("failed to insert tenant metadata: %w", err)
// // 	}

// // 	// Store DB credentials in Vault
// // 	if tenant.DBConfig != nil {
// // 		err = r.storeDBCredentialsInVault(tenant.ID, tenant.DBConfig)
// // 		if err != nil {
// // 			return fmt.Errorf("failed to store DB credentials in Vault: %w", err)
// // 		}
// // 	}

// // 	// Commit both transactions
// // 	if err := txMain.Commit(); err != nil {
// // 		return fmt.Errorf("failed to commit transaction on mainDB: %w", err)
// // 	}
// // 	if err := txTenant.Commit(); err != nil {
// // 		return fmt.Errorf("failed to commit transaction on tenantDB: %w", err)
// // 	}

// // 	return nil
// // }

// // func (r *sqlTenantRepository) GetTenantByID(ctx context.Context, tenantID string) (domain.Tenant, error) {
// // 	var tenant domain.Tenant

// // 	metadataQuery := `SELECT id, name FROM tenants WHERE id = ?`
// // 	err := r.mainDB.QueryRow(ctx, metadataQuery, tenantID).Scan(&tenant.ID, &tenant.Name)
// // 	if err != nil {
// // 		return domain.Tenant{}, fmt.Errorf("failed to get tenant metadata: %w", err)
// // 	}

// // 	// Fetch DB credentials from Vault
// // 	dbConfig, err := r.getDBCredentialsFromVault(tenantID)
// // 	if err != nil {
// // 		return domain.Tenant{}, fmt.Errorf("failed to get DB credentials from Vault: %w", err)
// // 	}
// // 	tenant.DBConfig = dbConfig

// // 	return tenant, nil
// // }

// // func (r *sqlTenantRepository) Update(ctx context.Context, tenant domain.Tenant) error {
// // 	now := time.Now()

// // 	// Begin transaction on mainDB
// // 	txMain, err := r.mainDB.BeginTx(ctx, nil)
// // 	if err != nil {
// // 		return fmt.Errorf("failed to begin transaction on mainDB: %w", err)
// // 	}
// // 	defer txMain.Rollback()

// // 	// Begin transaction on tenantDB
// // 	txTenant, err := r.tenantDB.BeginTx(ctx, nil)
// // 	if err != nil {
// // 		return fmt.Errorf("failed to begin transaction on tenantDB: %w", err)
// // 	}
// // 	defer txTenant.Rollback()

// // 	// Update tenant data in mainDB
// // 	tenantQuery := `UPDATE tenants SET name = ?, updated_at = ? WHERE id = ?`
// // 	_, err = txMain.Exec(ctx, tenantQuery, tenant.Name, now, tenant.ID)
// // 	if err != nil {
// // 		return fmt.Errorf("failed to update tenant data: %w", err)
// // 	}

// // 	// Update tenant metadata in tenantDB
// // 	metadataQuery := `UPDATE tenant_metadata SET name = ?, updated_at = ? WHERE id = ?`
// // 	_, err = txTenant.Exec(ctx, metadataQuery, tenant.Name, now, tenant.ID)
// // 	if err != nil {
// // 		return fmt.Errorf("failed to update tenant metadata: %w", err)
// // 	}

// // 	// Update DB credentials in Vault
// // 	if tenant.DBConfig != nil {
// // 		err = r.updateDBCredentialsInVault(tenant.ID, tenant.DBConfig)
// // 		if err != nil {
// // 			return fmt.Errorf("failed to update DB credentials in Vault: %w", err)
// // 		}
// // 	}

// // 	// Commit transactions
// // 	if err := txMain.Commit(); err != nil {
// // 		return fmt.Errorf("failed to commit transaction on mainDB: %w", err)
// // 	}
// // 	if err := txTenant.Commit(); err != nil {
// // 		return fmt.Errorf("failed to commit transaction on tenantDB: %w", err)
// // 	}

// // 	return nil
// // }

// // func (r *sqlTenantRepository) storeDBCredentialsInVault(tenantID string, dbConfig *domain.DBCredentials) error {

// // 	data := map[string]interface{}{
// // 		"type":     dbConfig.Type,
// // 		"host":     dbConfig.Host,
// // 		"port":     dbConfig.Port,
// // 		"username": dbConfig.Username,
// // 		"password": dbConfig.Password,
// // 		"dbname":   dbConfig.DBName,
// // 		"dsn":      dbConfig.DSN,
// // 	}

// // 	err := vault.CreateCredential(tenantID, vault.DBCredential, "", data)
// // 	if err != nil {
// // 		return fmt.Errorf("failed to store DB credentials in Vault for database %s", err)
// // 	}

// // 	return nil
// // }

// // func (r *sqlTenantRepository) updateDBCredentialsInVault(tenantID string, dbConfig *domain.DBCredentials) error {
// // 	data := map[string]interface{}{
// // 		"type":     dbConfig.Type,
// // 		"host":     dbConfig.Host,
// // 		"port":     dbConfig.Port,
// // 		"username": dbConfig.Username,
// // 		"password": dbConfig.Password,
// // 		"dbname":   dbConfig.DBName,
// // 		"dsn":      dbConfig.DSN,
// // 	}

// // 	err := vault.UpdateCredential(tenantID, vault.DBCredential, "", data)
// // 	if err != nil {
// // 		return fmt.Errorf("failed to update DB credentials in Vault for database %s: %w", dbConfig.DBName, err)
// // 	}

// // 	return nil
// // }

// // func (r *sqlTenantRepository) getDBCredentialsFromVault(tenantID string) (*domain.DBCredentials, error) {
// // 	// Retrieve credentials from Vault
// // 	dbConfigData, err := vault.GetClientCredentials(tenantID, vault.DBCredential, "")
// // 	if err != nil {
// // 		return nil, fmt.Errorf("failed to retrieve DB credentials from Vault: %w", err)
// // 	}

// // 	// Check if credentials map is empty
// // 	if len(dbConfigData) == 0 {
// // 		return nil, fmt.Errorf("no DB credentials found in Vault for tenant %s", tenantID)
// // 	}
// // 	portStr := dbConfigData["port"].(json.Number).String()
// // 	port, err := strconv.Atoi(portStr)

// // 	if err != nil {
// // 		return nil, fmt.Errorf("port error for tenant %s", tenantID)
// // 	}
// // 	// Extract and assign credential data
// // 	dbConfig := &domain.DBCredentials{
// // 		Type:     dbConfigData["type"].(string),
// // 		Host:     dbConfigData["host"].(string),
// // 		Port:     port,
// // 		Username: dbConfigData["username"].(string),
// // 		Password: dbConfigData["password"].(string),
// // 		DBName:   dbConfigData["dbname"].(string),
// // 		DSN:      dbConfigData["dsn"].(string),
// // 	}

// // 	// Return the credentials
// // 	return dbConfig, nil
// // }

// // type sqlTenantsRepository struct {
// // 	mainDB db.Database
// // }

// // func NewTenantsRepository(mainDB db.Database) repository.TenantsRepository {
// // 	return &sqlTenantsRepository{mainDB: mainDB}
// // }

// // func (r *sqlTenantsRepository) GetAllTenants(ctx context.Context) ([]domain.Tenant, error) {
// // 	query := `SELECT id, name FROM tenants`
// // 	rows, err := r.mainDB.Query(ctx, query)
// // 	if err != nil {
// // 		return nil, fmt.Errorf("failed to query tenant metadata: %w", err)
// // 	}
// // 	defer rows.Close()

// // 	var tenants []domain.Tenant
// // 	for rows.Next() {
// // 		var tenant domain.Tenant
// // 		err := rows.Scan(&tenant.ID, &tenant.Name)
// // 		if err != nil {
// // 			return nil, fmt.Errorf("failed to scan tenant metadata: %w", err)
// // 		}
// // 		tenants = append(tenants, tenant)
// // 	}

// // 	return tenants, nil
// // }

// package repos

// import (
//     "context"
//     "encoding/json"
//     "fmt"
//     "strconv"
//     "time"

//     "getnoti.com/internal/tenants/domain"
//     repository "getnoti.com/internal/tenants/repos"
//     "getnoti.com/pkg/credentials"
//     "getnoti.com/pkg/db"
// )

// type sqlTenantRepository struct {
//     mainDB            db.Database
//     tenantDB          db.Database
//     credentialManager *credentials.Manager
// }

// func NewTenantRepository(mainDB, tenantDB db.Database, credentialManager *credentials.Manager) repository.TenantRepository {
//     return &sqlTenantRepository{
//         mainDB:            mainDB,
//         tenantDB:          tenantDB,
//         credentialManager: credentialManager,
//     }
// }

// func (r *sqlTenantRepository) getDBCredentialsFromCredentialManager(tenantID string) (*domain.DBCredentials, error) {
//     // Use credential manager instead of direct vault calls
//     dbConfigData, err := r.credentialManager.GetCredentials(tenantID, credentials.DBCredential, "default")
//     if err != nil {
//         return nil, fmt.Errorf("failed to get database credentials for tenant %s: %w", tenantID, err)
//     }

//     // Check if credentials map is empty
//     if len(dbConfigData) == 0 {
//         return nil, fmt.Errorf("empty database credentials for tenant %s", tenantID)
//     }

//     // Extract port safely
//     var port int
//     if portVal, exists := dbConfigData["port"]; exists {
//         switch v := portVal.(type) {
//         case int:
//             port = v
//         case string:
//             if p, err := strconv.Atoi(v); err == nil {
//                 port = p
//             }
//         case float64:
//             port = int(v)
//         }
//     }

//     // Extract and assign credential data with proper type assertions
//     dbConfig := &domain.DBCredentials{}
    
//     if v, ok := dbConfigData["type"].(string); ok {
//         dbConfig.Type = v
//     }
//     if v, ok := dbConfigData["host"].(string); ok {
//         dbConfig.Host = v
//     }
//     dbConfig.Port = port
//     if v, ok := dbConfigData["username"].(string); ok {
//         dbConfig.Username = v
//     }
//     if v, ok := dbConfigData["password"].(string); ok {
//         dbConfig.Password = v
//     }
//     if v, ok := dbConfigData["dbname"].(string); ok {
//         dbConfig.DBName = v
//     }
//     if v, ok := dbConfigData["dsn"].(string); ok {
//         dbConfig.DSN = v
//     }

//     return dbConfig, nil
// }

// type sqlTenantsRepository struct {
//     mainDB            db.Database
//     credentialManager *credentials.Manager
// }

// func NewTenantsRepository(mainDB db.Database, credentialManager *credentials.Manager) repository.TenantsRepository {
//     return &sqlTenantsRepository{
//         mainDB:            mainDB,
//         credentialManager: credentialManager,
//     }
// }

// func (r *sqlTenantsRepository) GetAllTenants(ctx context.Context) ([]domain.Tenant, error) {
//     query := `SELECT id, name, created_at, updated_at FROM tenants`
//     rows, err := r.mainDB.Query(ctx, query)
//     if err != nil {
//         return nil, fmt.Errorf("failed to query tenants: %w", err)
//     }
//     defer rows.Close()

//     var tenants []domain.Tenant
//     for rows.Next() {
//         var tenant domain.Tenant
//         var createdAt, updatedAt time.Time
        
//         err := rows.Scan(&tenant.ID, &tenant.Name, &createdAt, &updatedAt)
//         if err != nil {
//             return nil, fmt.Errorf("failed to scan tenant: %w", err)
//         }
        
//         tenant.CreatedAt = createdAt
//         tenant.UpdatedAt = updatedAt
//         tenants = append(tenants, tenant)
//     }

//     return tenants, nil
// }

// func (r *sqlTenantsRepository) GetTenantByID(ctx context.Context, tenantID string) (*domain.Tenant, error) {
//     query := `SELECT id, name, created_at, updated_at FROM tenants WHERE id = ?`
    
//     var tenant domain.Tenant
//     var createdAt, updatedAt time.Time
    
//     err := r.mainDB.QueryRow(ctx, query, tenantID).Scan(
//         &tenant.ID, &tenant.Name, &createdAt, &updatedAt)
//     if err != nil {
//         return nil, fmt.Errorf("failed to get tenant by ID %s: %w", tenantID, err)
//     }
    
//     tenant.CreatedAt = createdAt
//     tenant.UpdatedAt = updatedAt
    
//     return &tenant, nil
// }

// func (r *sqlTenantsRepository) CreateTenant(ctx context.Context, tenant *domain.Tenant) error {
//     query := `INSERT INTO tenants (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`
    
//     now := time.Now()
//     tenant.CreatedAt = now
//     tenant.UpdatedAt = now
    
//     _, err := r.mainDB.Exec(ctx, query, tenant.ID, tenant.Name, tenant.CreatedAt, tenant.UpdatedAt)
//     if err != nil {
//         return fmt.Errorf("failed to create tenant: %w", err)
//     }
    
//     return nil
// }

// func (r *sqlTenantsRepository) UpdateTenant(ctx context.Context, tenant *domain.Tenant) error {
//     query := `UPDATE tenants SET name = ?, updated_at = ? WHERE id = ?`
    
//     tenant.UpdatedAt = time.Now()
    
//     _, err := r.mainDB.Exec(ctx, query, tenant.Name, tenant.UpdatedAt, tenant.ID)
//     if err != nil {
//         return fmt.Errorf("failed to update tenant: %w", err)
//     }
    
//     return nil
// }

// func (r *sqlTenantsRepository) DeleteTenant(ctx context.Context, tenantID string) error {
//     query := `DELETE FROM tenants WHERE id = ?`
    
//     _, err := r.mainDB.Exec(ctx, query, tenantID)
//     if err != nil {
//         return fmt.Errorf("failed to delete tenant: %w", err)
//     }
    
//     return nil
// }

// // SetTenantDatabaseCredentials stores database credentials for a tenant
// func (r *sqlTenantsRepository) SetTenantDatabaseCredentials(ctx context.Context, tenantID string, credentials map[string]interface{}) error {
//     return r.credentialManager.StoreCredentials(tenantID, credentials.DBCredential, "default", credentials)
// }

// // GetTenantDatabaseCredentials retrieves database credentials for a tenant
// func (r *sqlTenantsRepository) GetTenantDatabaseCredentials(ctx context.Context, tenantID string) (map[string]interface{}, error) {
//     return r.credentialManager.GetCredentials(tenantID, credentials.DBCredential, "default")
// }

package repos

import (
    "context"
    "encoding/json"
    "fmt"
    "strconv"

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

    // Insert into mainDB - let database handle timestamps
    tenantQuery := `INSERT INTO tenants (id, name) VALUES (?, ?)`
    _, err = txMain.Exec(ctx, tenantQuery, tenant.ID, tenant.Name)
    if err != nil {
        return fmt.Errorf("failed to insert tenant data: %w", err)
    }

    // Insert into tenantDB - let database handle timestamps
    metadataQuery := `INSERT INTO tenant_metadata (id, name) VALUES (?, ?)`
    _, err = txTenant.Exec(ctx, metadataQuery, tenant.ID, tenant.Name)
    if err != nil {
        return fmt.Errorf("failed to insert tenant metadata: %w", err)
    }

    // Store DB credentials in Vault if provided
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

    // Query without timestamps - let database handle them
    metadataQuery := `SELECT id, name FROM tenants WHERE id = ?`
    err := r.mainDB.QueryRow(ctx, metadataQuery, tenantID).Scan(&tenant.ID, &tenant.Name)
    if err != nil {
        return domain.Tenant{}, fmt.Errorf("failed to get tenant metadata: %w", err)
    }

    // Fetch DB credentials from Vault if available
    dbConfig, err := r.getDBCredentialsFromVault(tenantID)
    if err != nil {
        // Don't fail if no credentials found - tenant might not have custom DB
        tenant.DBConfig = nil
    } else {
        tenant.DBConfig = dbConfig
    }

    return tenant, nil
}

func (r *sqlTenantRepository) Update(ctx context.Context, tenant domain.Tenant) error {
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

    // Update tenant data in mainDB - database handles updated_at
    tenantQuery := `UPDATE tenants SET name = ? WHERE id = ?`
    _, err = txMain.Exec(ctx, tenantQuery, tenant.Name, tenant.ID)
    if err != nil {
        return fmt.Errorf("failed to update tenant data: %w", err)
    }

    // Update tenant metadata in tenantDB - database handles updated_at
    metadataQuery := `UPDATE tenant_metadata SET name = ? WHERE id = ?`
    _, err = txTenant.Exec(ctx, metadataQuery, tenant.Name, tenant.ID)
    if err != nil {
        return fmt.Errorf("failed to update tenant metadata: %w", err)
    }

    // Update DB credentials in Vault if provided
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

    err := vault.CreateCredential(tenantID, vault.DBCredential, "default", data)
    if err != nil {
        return fmt.Errorf("failed to store DB credentials in Vault: %w", err)
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

    err := vault.UpdateCredential(tenantID, vault.DBCredential, "default", data)
    if err != nil {
        return fmt.Errorf("failed to update DB credentials in Vault: %w", err)
    }

    return nil
}

func (r *sqlTenantRepository) getDBCredentialsFromVault(tenantID string) (*domain.DBCredentials, error) {
    // Retrieve credentials from Vault
    dbConfigData, err := vault.GetClientCredentials(tenantID, vault.DBCredential, "default")
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve DB credentials from Vault: %w", err)
    }

    // Check if credentials map is empty
    if len(dbConfigData) == 0 {
        return nil, fmt.Errorf("no DB credentials found in Vault for tenant %s", tenantID)
    }

    // Extract port safely
    var port int
    if portVal, exists := dbConfigData["port"]; exists {
        switch v := portVal.(type) {
        case int:
            port = v
        case string:
            if p, err := strconv.Atoi(v); err == nil {
                port = p
            }
        case float64:
            port = int(v)
        case json.Number:
            if p, err := v.Int64(); err == nil {
                port = int(p)
            }
        }
    }

    // Extract and assign credential data with proper type assertions
    dbConfig := &domain.DBCredentials{}
    
    if v, ok := dbConfigData["type"].(string); ok {
        dbConfig.Type = v
    }
    if v, ok := dbConfigData["host"].(string); ok {
        dbConfig.Host = v
    }
    dbConfig.Port = port
    if v, ok := dbConfigData["username"].(string); ok {
        dbConfig.Username = v
    }
    if v, ok := dbConfigData["password"].(string); ok {
        dbConfig.Password = v
    }
    if v, ok := dbConfigData["dbname"].(string); ok {
        dbConfig.DBName = v
    }
    if v, ok := dbConfigData["dsn"].(string); ok {
        dbConfig.DSN = v
    }

    return dbConfig, nil
}

type sqlTenantsRepository struct {
    mainDB db.Database
}

func NewTenantsRepository(mainDB db.Database) repository.TenantsRepository {
    return &sqlTenantsRepository{mainDB: mainDB}
}

func (r *sqlTenantsRepository) GetAllTenants(ctx context.Context) ([]domain.Tenant, error) {
    // Query without timestamps - let database handle them
    query := `SELECT id, name FROM tenants`
    rows, err := r.mainDB.Query(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to query tenants: %w", err)
    }
    defer rows.Close()

    var tenants []domain.Tenant
    for rows.Next() {
        var tenant domain.Tenant
        err := rows.Scan(&tenant.ID, &tenant.Name)
        if err != nil {
            return nil, fmt.Errorf("failed to scan tenant: %w", err)
        }
        tenants = append(tenants, tenant)
    }

    return tenants, nil
}