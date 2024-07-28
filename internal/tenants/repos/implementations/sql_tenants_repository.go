package repos

import (
	"context"

	"encoding/json"
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
	preferences, err := json.Marshal(tenant.Preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

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

	tenantQuery := `INSERT INTO tenants (id, name, preferences, created_at, updated_at) 
                    VALUES (?, ?, ?, ?, ?)`
	_, err = r.tenantDB.Exec(ctx, tenantQuery, tenant.ID, tenant.Name, preferences, now, now)
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
	var preferences []byte

	metadataQuery := `SELECT id, name FROM tenant_metadata WHERE id = ?`
	err := r.mainDB.QueryRow(ctx, metadataQuery, tenantID).Scan(&tenant.ID, &tenant.Name)
	if err != nil {
		return domain.Tenant{}, fmt.Errorf("failed to get tenant metadata: %w", err)
	}

	tenantQuery := `SELECT preferences FROM tenants WHERE id = ?`
	err = r.tenantDB.QueryRow(ctx, tenantQuery, tenantID).Scan(&preferences)
	if err != nil {
		return domain.Tenant{}, fmt.Errorf("failed to get tenant data: %w", err)
	}

	err = json.Unmarshal(preferences, &tenant.Preferences)
	if err != nil {
		return domain.Tenant{}, fmt.Errorf("failed to unmarshal preferences: %w", err)
	}

	return tenant, nil
}

func (r *sqlTenantRepository) Update(ctx context.Context, tenant domain.Tenant) error {
    now := time.Now()
    preferences, err := json.Marshal(tenant.Preferences)
    if err != nil {
        return fmt.Errorf("failed to marshal preferences: %w", err)
    }

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

    tenantQuery := `UPDATE tenants SET name = ?, preferences = ?, updated_at = ? WHERE id = ?`
    _, err = r.tenantDB.Exec(ctx, tenantQuery, tenant.Name, preferences, now, tenant.ID)
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


func (r *sqlTenantRepository) GetAllTenants(ctx context.Context) ([]domain.Tenant, error) {
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

func (r *sqlTenantRepository) GetPreferenceByChannel(ctx context.Context, tenantID string, channel string) (map[string]string, error) {
	query := `SELECT preferences FROM tenants WHERE id = ?`
	var preferences []byte
	err := r.tenantDB.QueryRow(ctx, query, tenantID).Scan(&preferences)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant preferences: %w", err)
	}

	var prefs map[string]map[string]string
	err = json.Unmarshal(preferences, &prefs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal preferences: %w", err)
	}

	channelPref, ok := prefs[channel]
	if !ok {
		return nil, fmt.Errorf("channel preference not found")
	}

	return channelPref, nil
}

func (r *sqlTenantRepository) storeDBCredentialsInVault(tenantID string, dbConfigs map[string]*domain.DBConfig) error {
	err := vault.RefreshToken(tenantID)
	if err != nil {
		return fmt.Errorf("failed to refresh Vault token: %w", err)
	}

	for dbName, dbConfig := range dbConfigs {
		data := map[string]interface{}{
			"create_new_db": dbConfig.CreateNewDB,
		}

		if dbConfig.Credentials != nil {
			data["type"] = dbConfig.Credentials.Type
			data["host"] = dbConfig.Credentials.Host
			data["port"] = dbConfig.Credentials.Port
			data["username"] = dbConfig.Credentials.Username
			data["password"] = dbConfig.Credentials.Password
			data["dbname"] = dbConfig.Credentials.DBName
			data["dsn"] = dbConfig.Credentials.DSN
		}

		err = vault.CreateCredential(tenantID, vault.DBCredential, dbName, data)
		if err != nil {
			return fmt.Errorf("failed to store DB credentials in Vault for database %s: %w", dbName, err)
		}
	}

	return nil
}
func (r *sqlTenantRepository) updateDBCredentialsInVault(tenantID string, dbConfigs map[string]*domain.DBConfig) error {
    err := vault.RefreshToken(tenantID)
    if err != nil {
        return fmt.Errorf("failed to refresh Vault token: %w", err)
    }

    // Get existing credentials
    existingCredentials, err := vault.GetClientCredentials(tenantID, vault.DBCredential, "")
    if err != nil {
        return fmt.Errorf("failed to get existing DB credentials from Vault: %w", err)
    }

    for dbName, dbConfig := range dbConfigs {
        data := map[string]interface{}{
            "create_new_db": dbConfig.CreateNewDB,
        }

        if dbConfig.Credentials != nil {
            data["type"] = dbConfig.Credentials.Type
            data["host"] = dbConfig.Credentials.Host
            data["port"] = dbConfig.Credentials.Port
            data["username"] = dbConfig.Credentials.Username
            data["password"] = dbConfig.Credentials.Password
            data["dbname"] = dbConfig.Credentials.DBName
            data["dsn"] = dbConfig.Credentials.DSN
        }

        // Check if the database config already exists
        if _, exists := existingCredentials[dbName]; exists {
            err = vault.UpdateCredential(tenantID, vault.DBCredential, dbName, data)
        } else {
            err = vault.CreateCredential(tenantID, vault.DBCredential, dbName, data)
        }

        if err != nil {
            return fmt.Errorf("failed to update/create DB credentials in Vault for database %s: %w", dbName, err)
        }
    }

    // Remove any database configs that are no longer present
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


// func (r *sqlTenantRepository) getDBCredentialsFromVault(tenantID string) (*domain.DBCredentials, error) {
//     credentials, err := vault.GetClientCredentials(r.vault, tenantID)
//     if err != nil {
//         return nil, fmt.Errorf("failed to retrieve DB credentials from Vault: %w", err)
//     }

//     dbConfigData, ok := credentials["db_config"].(map[string]interface{})
//     if !ok {
//         return nil, fmt.Errorf("invalid DB config data in Vault")
//     }

//     dbConfig := &domain.DBCredentials{}
//     // Map the values from dbConfigData to dbConfig
//     // You might need to adjust this based on your actual DBConfig structure
//     dbConfig.Host = dbConfigData["host"].(string)
//     dbConfig.Port = int(dbConfigData["port"].(float64))
//     dbConfig.Username = dbConfigData["username"].(string)
//     dbConfig.Password = dbConfigData["password"].(string)
//     dbConfig.DBName = dbConfigData["db_name"].(string)

//     return dbConfig, nil
// }

type sqlTenantPreferenceRepository struct {
	tenantDB db.Database
}

func NewTenantPreferenceRepository(tenantDB db.Database) repository.TenantPreferenceRepository {
	return &sqlTenantPreferenceRepository{tenantDB: tenantDB}
}

func (r *sqlTenantPreferenceRepository) GetPreferenceByChannel(ctx context.Context, tenantID string, channel string) (map[string]string, error) {
	query := `SELECT preferences FROM tenants WHERE id = ?`
	var preferences []byte
	err := r.tenantDB.QueryRow(ctx, query, tenantID).Scan(&preferences)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant preferences: %w", err)
	}

	var prefs map[string]map[string]string
	err = json.Unmarshal(preferences, &prefs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal preferences: %w", err)
	}

	channelPref, ok := prefs[channel]
	if !ok {
		return nil, fmt.Errorf("channel preference not found")
	}

	return channelPref, nil
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
