package tenantroutes

import (
	"bytes"
	"encoding/json"
	"fmt"

	"getnoti.com/internal/tenants/domain"
	"getnoti.com/internal/tenants/repos"
	"getnoti.com/internal/tenants/repos/implementations"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/migration"

	"github.com/google/uuid"
	"io"
	"net/http"
)

func (h *Handlers) createTenantRepo(r *http.Request) (repository.TenantRepository, error) {
    // Read the body content
    bodyContent, err := io.ReadAll(r.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read request body: %w", err)
    }
    // Close the original body
    r.Body.Close()

    // Extract tenant info
    tenantInfo, err := h.extractTenantInfo(bodyContent)
    if err != nil {
        return nil, fmt.Errorf("failed to extract tenant info: %w", err)
    }

    database, err := h.getDatabaseConnection(tenantInfo)
    if err != nil {
        return nil, fmt.Errorf("failed to get database connection: %w", err)
    }
	//perform migration
	if err := migrate.Migrate(tenantInfo.DBConfig.DSN,tenantInfo.DBConfig.Type,false); err != nil {
		return nil, fmt.Errorf("failed to run main Database migrations: %w", err)
	}
    // Update request body
    err = h.updateRequestBody(r, bodyContent, tenantInfo)
    if err != nil {
        return nil, fmt.Errorf("failed to update request body: %w", err)
    }

    tenantRepo := repos.NewTenantRepository(h.BaseHandler.MainDB, database)
    return tenantRepo, nil
}


func (h *Handlers) extractTenantInfo(bodyContent []byte) (*TenantInfo, error) {
    var requestBody map[string]interface{}
    if err := json.Unmarshal(bodyContent, &requestBody); err != nil {
        return nil, fmt.Errorf("invalid JSON in request body: %w", err)
    }

    tenantInfo := &TenantInfo{}
    tenantInfo.ID = h.getTenantID(requestBody)
    tenantInfo.DBConfig = h.extractDBConfig(requestBody)

    return tenantInfo, nil
}

func (h *Handlers) updateRequestBody(r *http.Request, bodyContent []byte, tenantInfo *TenantInfo) error {
    var requestBody map[string]interface{}
    if err := json.Unmarshal(bodyContent, &requestBody); err != nil {
        return fmt.Errorf("failed to decode request body: %w", err)
    }

    requestBody["id"] = tenantInfo.ID
    requestBody["dbConfig"] = tenantInfo.DBConfig

    updatedBody, err := json.Marshal(requestBody)
    if err != nil {
        return fmt.Errorf("failed to marshal updated request body: %w", err)
    }

    r.Body = io.NopCloser(bytes.NewBuffer(updatedBody))
    r.ContentLength = int64(len(updatedBody))

    return nil
}


func (h *Handlers) getTenantID(requestBody map[string]interface{}) string {
	if id, ok := requestBody["id"].(string); ok && id != "" {
		return id
	}
	return uuid.New().String()
}

func (h *Handlers) extractDBConfig(requestBody map[string]interface{}) *domain.DBCredentials {
	if config, ok := requestBody["dbConfig"].(map[string]interface{}); ok {
		return &domain.DBCredentials{
			Type:     config["Type"].(string),
			Host:     config["Host"].(string),
			Port:     int(config["Port"].(float64)),
			Username: config["Username"].(string),
			Password: config["Password"].(string),
			DBName:   config["DBName"].(string),
			DSN:      config["DSN"].(string),
		}
	}
	return nil
}

func (h *Handlers) getDatabaseConnection(tenantInfo *TenantInfo) (db.Database, error) {
	if tenantInfo.DBConfig != nil {
		dbConfig := h.convertDBCredentialsToMap(tenantInfo.DBConfig)
		return h.BaseHandler.DBManager.GetDatabaseConnectionWithConfig(tenantInfo.ID, dbConfig)
	}

	database, dbConfigMap, err := h.BaseHandler.DBManager.CreateNewTenantDatabase(tenantInfo.ID)
	if err != nil {
		return nil, err
	}

	dbCredentials := &domain.DBCredentials{
		Type: dbConfigMap["type"].(string),
		DSN:  dbConfigMap["dsn"].(string),
	}

	tenantInfo.DBConfig = dbCredentials

	return database, nil
}

// Helper function to convert DBCredentials to map[string]interface{}
func (h *Handlers) convertDBCredentialsToMap(dbCredentials *domain.DBCredentials) map[string]interface{} {
	if dbCredentials.DSN != "" {
		return map[string]interface{}{
			"type": dbCredentials.Type,
			"dsn":  dbCredentials.DSN,
		}
	}

	return map[string]interface{}{
		"type":     dbCredentials.Type,
		"host":     dbCredentials.Host,
		"port":     dbCredentials.Port,
		"username": dbCredentials.Username,
		"password": dbCredentials.Password,
		"database": dbCredentials.DBName,
	}
}



type TenantInfo struct {
	ID       string
	DBConfig *domain.DBCredentials
}
