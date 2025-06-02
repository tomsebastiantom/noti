// package tenantroutes

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"

// 	"getnoti.com/internal/tenants/domain"
// 	"getnoti.com/internal/tenants/repos"
// 	"getnoti.com/internal/tenants/repos/implementations"
// 	"getnoti.com/pkg/db"
// 	"getnoti.com/pkg/migration"

// 	"github.com/google/uuid"
// 	"io"
// 	"net/http"
// )


// func (h *Handlers) createTenantRepo(r *http.Request) (repository.TenantsRepository, error) {
//     // Read the body content
//     bodyContent, err := io.ReadAll(r.Body)
//     if err != nil {
//         return nil, fmt.Errorf("failed to read request body: %w", err)
//     }
//     // Close the original body
//     r.Body.Close()

//     // Extract tenant info
//     tenantInfo, err := h.extractTenantInfo(bodyContent)
//     if err != nil {
//         return nil, fmt.Errorf("failed to extract tenant info: %w", err)
//     }

//     database, err := h.getDatabaseConnection(tenantInfo)
//     if err != nil {
//         return nil, fmt.Errorf("failed to get database connection: %w", err)
//     }
    
//     // Perform migration
//     if err := migration.Migrate(tenantInfo.DBConfig.DSN, tenantInfo.DBConfig.Type, false); err != nil {
//         return nil, fmt.Errorf("failed to run tenant database migrations: %w", err)
//     }
    
//     // Update request body
//     err = h.updateRequestBody(r, bodyContent, tenantInfo)
//     if err != nil {
//         return nil, fmt.Errorf("failed to update request body: %w", err)
//     }

//     // Get credential manager from container
//     credManager := h.BaseHandler.CredentialManager
    
//     // Create tenant repository
//     tenantRepo := repos.NewTenantRepository(h.mainDB, credManager)
//     return tenantRepo, nil
// }


// func (h *Handlers) extractTenantInfo(bodyContent []byte) (*TenantInfo, error) {
//     var requestBody map[string]interface{}
//     if err := json.Unmarshal(bodyContent, &requestBody); err != nil {
//         return nil, fmt.Errorf("invalid JSON in request body: %w", err)
//     }

//     tenantInfo := &TenantInfo{}
//     tenantInfo.ID = h.getTenantID(requestBody)
//     tenantInfo.DBConfig = h.extractDBConfig(requestBody)

//     return tenantInfo, nil
// }

// func (h *Handlers) updateRequestBody(r *http.Request, bodyContent []byte, tenantInfo *TenantInfo) error {
//     var requestBody map[string]interface{}
//     if err := json.Unmarshal(bodyContent, &requestBody); err != nil {
//         return fmt.Errorf("failed to decode request body: %w", err)
//     }

//     requestBody["id"] = tenantInfo.ID
//     requestBody["dbConfig"] = tenantInfo.DBConfig

//     updatedBody, err := json.Marshal(requestBody)
//     if err != nil {
//         return fmt.Errorf("failed to marshal updated request body: %w", err)
//     }

//     r.Body = io.NopCloser(bytes.NewBuffer(updatedBody))
//     r.ContentLength = int64(len(updatedBody))

//     return nil
// }


// func (h *Handlers) getTenantID(requestBody map[string]interface{}) string {
// 	if id, ok := requestBody["id"].(string); ok && id != "" {
// 		return id
// 	}
// 	return uuid.New().String()
// }

// func (h *Handlers) extractDBConfig(requestBody map[string]interface{}) *domain.DBCredentials {
// 	if config, ok := requestBody["dbConfig"].(map[string]interface{}); ok {
// 		return &domain.DBCredentials{
// 			Type:     config["Type"].(string),
// 			Host:     config["Host"].(string),
// 			Port:     int(config["Port"].(float64)),
// 			Username: config["Username"].(string),
// 			Password: config["Password"].(string),
// 			DBName:   config["DBName"].(string),
// 			DSN:      config["DSN"].(string),
// 		}
// 	}
// 	return nil
// }

// func (h *Handlers) getDatabaseConnection(tenantInfo *TenantInfo) (db.Database, error) {
// 	if tenantInfo.DBConfig != nil {
// 		dbConfig := h.convertDBCredentialsToMap(tenantInfo.DBConfig)
// 		return h.BaseHandler.DBManager.GetDatabaseConnectionWithConfig(tenantInfo.ID, dbConfig)
// 	}

// 	database, dbConfigMap, err := h.BaseHandler.DBManager.CreateNewTenantDatabase(tenantInfo.ID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	dbCredentials := &domain.DBCredentials{
// 		Type: dbConfigMap["type"].(string),
// 		DSN:  dbConfigMap["dsn"].(string),
// 	}

// 	tenantInfo.DBConfig = dbCredentials

// 	return database, nil
// }

// // Helper function to convert DBCredentials to map[string]interface{}
// func (h *Handlers) convertDBCredentialsToMap(dbCredentials *domain.DBCredentials) map[string]interface{} {
// 	if dbCredentials.DSN != "" {
// 		return map[string]interface{}{
// 			"type": dbCredentials.Type,
// 			"dsn":  dbCredentials.DSN,
// 		}
// 	}

// 	return map[string]interface{}{
// 		"type":     dbCredentials.Type,
// 		"host":     dbCredentials.Host,
// 		"port":     dbCredentials.Port,
// 		"username": dbCredentials.Username,
// 		"password": dbCredentials.Password,
// 		"database": dbCredentials.DBName,
// 	}
// }



// type TenantInfo struct {
// 	ID       string
// 	DBConfig *domain.DBCredentials
// }

package tenantroutes

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"

    "getnoti.com/internal/tenants/domain"
    "github.com/google/uuid"
)


func (h *Handlers) extractTenantInfo(bodyContent []byte) (*TenantInfo, error) {
    var requestBody map[string]interface{}
    if err := json.Unmarshal(bodyContent, &requestBody); err != nil {
        return nil, fmt.Errorf("invalid JSON in request body: %w", err)
    }

    tenantInfo := &TenantInfo{}
    
    // Extract tenant ID or generate one if not present
    if id, ok := requestBody["id"].(string); ok && id != "" {
        tenantInfo.ID = id
    } else {
        tenantInfo.ID = uuid.New().String()
    }
    
    // Extract DB config if present
    if dbConfigData, ok := requestBody["dbConfig"].(map[string]interface{}); ok {
        tenantInfo.DBConfig = &domain.DBCredentials{}
        
        // Extract fields with type checking
        if t, ok := dbConfigData["type"].(string); ok {
            tenantInfo.DBConfig.Type = t
        }
        if dsn, ok := dbConfigData["dsn"].(string); ok {
            tenantInfo.DBConfig.DSN = dsn
        }
        if host, ok := dbConfigData["host"].(string); ok {
            tenantInfo.DBConfig.Host = host
        }
        if port, ok := dbConfigData["port"].(float64); ok {
            tenantInfo.DBConfig.Port = int(port)
        }
        if username, ok := dbConfigData["username"].(string); ok {
            tenantInfo.DBConfig.Username = username
        }
        if password, ok := dbConfigData["password"].(string); ok {
            tenantInfo.DBConfig.Password = password
        }
        if dbName, ok := dbConfigData["dbName"].(string); ok {
            tenantInfo.DBConfig.DBName = dbName
        }
    }

    return tenantInfo, nil
}

// updateRequestBody updates the request body with potentially modified tenant info
func (h *Handlers) updateRequestBody(r *http.Request, originalBodyContent []byte, tenantInfo *TenantInfo) error {
    var requestBody map[string]interface{}
    if err := json.Unmarshal(originalBodyContent, &requestBody); err != nil {
        return fmt.Errorf("failed to decode request body: %w", err)
    }

    // Update tenant ID
    requestBody["id"] = tenantInfo.ID
    
    // Update DB config
    if tenantInfo.DBConfig != nil {
        dbConfigMap := make(map[string]interface{})
        dbConfigMap["type"] = tenantInfo.DBConfig.Type
        if tenantInfo.DBConfig.DSN != "" {
            dbConfigMap["dsn"] = tenantInfo.DBConfig.DSN
        }
        if tenantInfo.DBConfig.Host != "" {
            dbConfigMap["host"] = tenantInfo.DBConfig.Host
        }
        if tenantInfo.DBConfig.Port != 0 {
            dbConfigMap["port"] = tenantInfo.DBConfig.Port
        }
        if tenantInfo.DBConfig.Username != "" {
            dbConfigMap["username"] = tenantInfo.DBConfig.Username
        }
        if tenantInfo.DBConfig.Password != "" {
            dbConfigMap["password"] = tenantInfo.DBConfig.Password
        }
        if tenantInfo.DBConfig.DBName != "" {
            dbConfigMap["dbName"] = tenantInfo.DBConfig.DBName
        }
        requestBody["dbConfig"] = dbConfigMap
    }

    // Convert back to JSON
    updatedBody, err := json.Marshal(requestBody)
    if err != nil {
        return fmt.Errorf("failed to marshal updated request body: %w", err)
    }

    // Replace request body
    r.Body = io.NopCloser(bytes.NewBuffer(updatedBody))
    r.ContentLength = int64(len(updatedBody))

    return nil
}


func (h *Handlers) processTenantRequest(r *http.Request) (*TenantInfo, error) {
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
        return nil, err
    }

    // Configure database
    _, err = h.getDatabaseConnection(tenantInfo)
    if err != nil {
        return nil, err
    }

    // Update request body
    err = h.updateRequestBody(r, bodyContent, tenantInfo)
    if err != nil {
        return nil, err
    }

    return tenantInfo, nil
}