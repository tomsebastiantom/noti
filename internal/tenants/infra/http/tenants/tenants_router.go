// package tenantroutes

// import (
// 	"net/http"

// 	"getnoti.com/internal/shared/handler"
// 	tenantMiddleware "getnoti.com/internal/shared/middleware"
// 	"getnoti.com/internal/shared/utils"
// 	"getnoti.com/internal/tenants/domain"
// 	"getnoti.com/internal/tenants/repos"
// 	repos "getnoti.com/internal/tenants/repos/implementations"
// 	"getnoti.com/internal/tenants/usecases/create_tenant"
// 	"getnoti.com/internal/tenants/usecases/get_tenant"
// 	"getnoti.com/internal/tenants/usecases/get_tenants"
// 	"getnoti.com/internal/tenants/usecases/update_tenant"
// 	"getnoti.com/pkg/credentials"
// 	"getnoti.com/pkg/db"
// 	"github.com/go-chi/chi/v5"
// 	"github.com/google/uuid"
// )

// type Handlers struct {
//     BaseHandler      *handler.BaseHandler
//     mainDB           db.Database
//     credentialManager *credentials.Manager
// }

// func NewHandlers(baseHandler *handler.BaseHandler, mainDB db.Database, credentialManager *credentials.Manager) *Handlers {
//     return &Handlers{
//         BaseHandler:      baseHandler,
//         mainDB:           mainDB,
//         credentialManager: credentialManager,
//     }
// }

// // Helper function to retrieve tenant ID and database connection
// func (h *Handlers) getTenantsRepo(r *http.Request) (repository.TenantsRepository, error) {

//     tenantsRepo := repos.NewTenantRepository(h.mainDB, h.credentialManager)
//     return tenantsRepo, nil
// }


// func (h *Handlers) CreateTenant(w http.ResponseWriter, r *http.Request) {
//     // 1. Parse the request
//     var req createtenant.CreateTenantRequest
//     if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
//         return
//     }
    
//     // 2. Generate UUID if tenant ID not provided
//     if req.ID == "" {
//         req.ID = uuid.New().String()
//     }
    
//     // 3. Create database if config not provided
//     if req.DBConfig == nil {
//         database, dbConfigMap, err := h.BaseHandler.DBManager.CreateNewTenantDatabase(req.ID)
//         if err != nil {
//             h.BaseHandler.HandleError(w, "Failed to create tenant database", err, http.StatusInternalServerError)
//             return
//         }
        
//         // Convert the config map to DBCredentials
//         req.DBConfig = &domain.DBCredentials{
//             Type: dbConfigMap["type"].(string),
//             DSN:  dbConfigMap["dsn"].(string),
//         }
        
//         // Run migrations on the new database
//         if err := migration.Migrate(req.DBConfig.DSN, req.DBConfig.Type, false); err != nil {
//             h.BaseHandler.HandleError(w, "Failed to run migrations", err, http.StatusInternalServerError)
//             return
//         }
//     } else {
//         // If config is provided, validate it and run migrations
//         // Create a test connection to verify credentials
//         dbConfig := map[string]interface{}{
//             "type": req.DBConfig.Type,
//             "dsn":  req.DBConfig.DSN,
//         }
        
//         // Validate connection
//         conn, err := h.BaseHandler.DBManager.GetDatabaseConnectionWithConfig(req.ID, dbConfig)
//         if err != nil {
//             h.BaseHandler.HandleError(w, "Invalid database configuration", err, http.StatusBadRequest)
//             return
//         }
        
//         // Run migrations
//         if err := migration.Migrate(req.DBConfig.DSN, req.DBConfig.Type, false); err != nil {
//             h.BaseHandler.HandleError(w, "Failed to run migrations", err, http.StatusInternalServerError)
//             return
//         }
//     }
    
//     // 4. Create tenant in repository
//     tenantRepo := repos.NewTenantRepository(h.mainDB, h.credentialManager)
//     createTenantUseCase := createtenant.NewCreateTenantUseCase(tenantRepo)
//     createTenantController := createtenant.NewCreateTenantController(createTenantUseCase)
    
//     res, err := createTenantController.CreateTenant(r.Context(), req)
//     if err != nil {
//         h.BaseHandler.HandleError(w, "Failed to create tenant", err, http.StatusInternalServerError)
//         return
//     }
    
//     h.BaseHandler.RespondWithJSON(w, res)
// }

// // UpdateTenant handles the updating of an existing tenant
// func (h *Handlers) UpdateTenant(w http.ResponseWriter, r *http.Request) {
// 	tenantRepo, err := h.getTenantsRepo(r)
// 	if err != nil {
// 		h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
// 		return
// 	}

// 	updateTenantUseCase := updatetenant.NewUpdateTenantUseCase(tenantRepo)
// 	updateTenantController := updatetenant.NewUpdateTenantController(updateTenantUseCase)

// 	var req updatetenant.UpdateTenantRequest
// 	if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
// 		return
// 	}

// 	res, err := updateTenantController.UpdateTenant(r.Context(), req)
// 	if err != nil {
// 		h.BaseHandler.HandleError(w, "Failed to update tenant", err, http.StatusInternalServerError)
// 		return
// 	}

// 	h.BaseHandler.RespondWithJSON(w, res)
// }

// // GetTenant retrieves a tenant by ID
// func (h *Handlers) GetTenant(w http.ResponseWriter, r *http.Request) {
// 	tenantRepo, err := h.getTenantsRepo(r)
// 	if err != nil {
// 		h.BaseHandler.HandleError(w, "Failed to get tenant repository", err, http.StatusInternalServerError)
// 		return
// 	}

// 	getTenantUseCase := gettenant.NewGetTenantUseCase(tenantRepo)
// 	getTenantController := gettenant.NewGetTenantController(getTenantUseCase)
// 	id, err := utils.GetIDFromReq(r)
// 	if err != nil {
// 		h.BaseHandler.HandleError(w, "Failed to get tenant ID", err, http.StatusBadRequest)
// 		return
// 	}

// 	req := gettenant.GetTenantRequest{TenantID: id}

// 	res, err := getTenantController.GetTenant(r.Context(), req)
// 	if err != nil {
// 		h.BaseHandler.HandleError(w, "Failed to get tenant", err, http.StatusInternalServerError)
// 		return
// 	}

// 	h.BaseHandler.RespondWithJSON(w, res)
// }

// // GetTenants retrieves all tenants
// func (h *Handlers) GetTenants(w http.ResponseWriter, r *http.Request) {
// 	tenantsRepo,err := h.getTenantsRepo(r)
//     if err != nil {
// 		h.BaseHandler.HandleError(w, "Failed to get tenant repository", err, http.StatusInternalServerError)
// 		return
// 	}
// 	getTenantsUseCase := gettenants.NewGetTenantsUseCase(tenantsRepo)
// 	getTenantsController := gettenants.NewGetTenantsController(getTenantsUseCase)

// 	res, err := getTenantsController.GetTenants(r.Context(), gettenants.GetTenantsRequest{})
// 	if err != nil {
// 		h.BaseHandler.HandleError(w, "Failed to get tenants", err, http.StatusInternalServerError)
// 		return
// 	}

// 	h.BaseHandler.RespondWithJSON(w, res)
// }

// // NewRouter sets up the router with all routes
// func NewRouter(mainDB db.Database, dbManager *db.Manager, credentialManager *credentials.Manager) *chi.Mux {
// 	b := handler.NewBaseHandler(dbManager)
//     h := NewHandlers(b, mainDB, credentialManager)

// 	r := chi.NewRouter()

// 	// Set up routes
// 	r.Post("/", h.CreateTenant)
// 	r.With(tenantMiddleware.WithTenantID).Put("/", h.UpdateTenant)
// 	r.With(tenantMiddleware.WithTenantID).Get("/me", h.GetTenant)
// 	r.With(tenantMiddleware.WithTenantID).Get("/{id}", h.GetTenant)
// 	r.Get("/", h.GetTenants)

// 	return r
// }

package tenantroutes

import (
    "fmt"
    "net/http"

    "getnoti.com/internal/shared/handler"
    tenantMiddleware "getnoti.com/internal/shared/middleware"
    "getnoti.com/internal/shared/utils"
    "getnoti.com/internal/tenants/domain"
    repository "getnoti.com/internal/tenants/repos"
    repos "getnoti.com/internal/tenants/repos/implementations"
    "getnoti.com/internal/tenants/usecases/create_tenant"
    "getnoti.com/internal/tenants/usecases/get_tenant"
    "getnoti.com/internal/tenants/usecases/get_tenants"
    "getnoti.com/internal/tenants/usecases/update_tenant"
    "getnoti.com/pkg/credentials"
    "getnoti.com/pkg/db"
    "getnoti.com/pkg/migration"
    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
)

// TenantInfo stores tenant database configuration information
type TenantInfo struct {
    ID       string
    DBConfig *domain.DBCredentials
}

type Handlers struct {
    BaseHandler       *handler.BaseHandler
    mainDB            db.Database
    credentialManager *credentials.Manager
}

func NewHandlers(baseHandler *handler.BaseHandler, mainDB db.Database, credentialManager *credentials.Manager) *Handlers {
    return &Handlers{
        BaseHandler:       baseHandler,
        mainDB:            mainDB,
        credentialManager: credentialManager,
    }
}

// Helper function to retrieve tenant repository
func (h *Handlers) getTenantsRepo(r *http.Request) (repository.TenantsRepository, error) {
    tenantsRepo := repos.NewTenantRepository(h.mainDB, h.credentialManager)
    return tenantsRepo, nil
}

func (h *Handlers) CreateTenant(w http.ResponseWriter, r *http.Request) {
    // 1. Parse the request body
    var req createtenant.CreateTenantRequest
    if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
        return
    }
    
    // 2. Process tenant database configuration
    tenantInfo := &TenantInfo{
        ID:       req.ID,
        DBConfig: req.DBConfig,
    }
    
    // Generate tenant ID if not provided
    if tenantInfo.ID == "" {
        tenantInfo.ID = uuid.New().String()
        req.ID = tenantInfo.ID
    }
    
    // 3. Get or create database connection and configure it
    _, err := h.getDatabaseConnection(tenantInfo)
    if err != nil {
        h.BaseHandler.HandleError(w, "Failed to configure tenant database", err, http.StatusInternalServerError)
        return
    }
    
    // Update request with final configuration
    req.DBConfig = tenantInfo.DBConfig
    
    // 4. Create tenant in repository
    tenantRepo := repos.NewTenantRepository(h.mainDB, h.credentialManager)
    createTenantUseCase := createtenant.NewCreateTenantUseCase(tenantRepo)
    createTenantController := createtenant.NewCreateTenantController(createTenantUseCase)
    
    res, err := createTenantController.CreateTenant(r.Context(), req)
    if err != nil {
        h.BaseHandler.HandleError(w, "Failed to create tenant", err, http.StatusInternalServerError)
        return
    }
    
    h.BaseHandler.RespondWithJSON(w, res)
}

// UpdateTenant handles the updating of an existing tenant
func (h *Handlers) UpdateTenant(w http.ResponseWriter, r *http.Request) {
    tenantRepo, err := h.getTenantsRepo(r)
    if err != nil {
        h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
        return
    }

    updateTenantUseCase := updatetenant.NewUpdateTenantUseCase(tenantRepo)
    updateTenantController := updatetenant.NewUpdateTenantController(updateTenantUseCase)

    var req updatetenant.UpdateTenantRequest
    if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
        return
    }

    // If DB config is provided, validate the connection
    if req.DBConfig != nil {
        tenantInfo := &TenantInfo{
            ID:       req.ID,
            DBConfig: req.DBConfig,
        }
        _, err := h.validateDatabaseConfig(tenantInfo)
        if err != nil {
            h.BaseHandler.HandleError(w, "Invalid database configuration", err, http.StatusBadRequest)
            return
        }
    }

    res, err := updateTenantController.UpdateTenant(r.Context(), req)
    if err != nil {
        h.BaseHandler.HandleError(w, "Failed to update tenant", err, http.StatusInternalServerError)
        return
    }

    h.BaseHandler.RespondWithJSON(w, res)
}

// GetTenant retrieves a tenant by ID
func (h *Handlers) GetTenant(w http.ResponseWriter, r *http.Request) {
    tenantRepo, err := h.getTenantsRepo(r)
    if err != nil {
        h.BaseHandler.HandleError(w, "Failed to get tenant repository", err, http.StatusInternalServerError)
        return
    }

    getTenantUseCase := gettenant.NewGetTenantUseCase(tenantRepo)
    getTenantController := gettenant.NewGetTenantController(getTenantUseCase)
    id, err := utils.GetIDFromReq(r)
    if err != nil {
        h.BaseHandler.HandleError(w, "Failed to get tenant ID", err, http.StatusBadRequest)
        return
    }

    req := gettenant.GetTenantRequest{TenantID: id}

    res, err := getTenantController.GetTenant(r.Context(), req)
    if err != nil {
        h.BaseHandler.HandleError(w, "Failed to get tenant", err, http.StatusInternalServerError)
        return
    }

    h.BaseHandler.RespondWithJSON(w, res)
}

// GetTenants retrieves all tenants
func (h *Handlers) GetTenants(w http.ResponseWriter, r *http.Request) {
    tenantsRepo, err := h.getTenantsRepo(r)
    if err != nil {
        h.BaseHandler.HandleError(w, "Failed to get tenant repository", err, http.StatusInternalServerError)
        return
    }
    getTenantsUseCase := gettenants.NewGetTenantsUseCase(tenantsRepo)
    getTenantsController := gettenants.NewGetTenantsController(getTenantsUseCase)

    res, err := getTenantsController.GetTenants(r.Context(), gettenants.GetTenantsRequest{})
    if err != nil {
        h.BaseHandler.HandleError(w, "Failed to get tenants", err, http.StatusInternalServerError)
        return
    }

    h.BaseHandler.RespondWithJSON(w, res)
}

// Database connection related helper methods
func (h *Handlers) getDatabaseConnection(tenantInfo *TenantInfo) (db.Database, error) {
    // If DB config is provided, validate it
    if tenantInfo.DBConfig != nil {
        return h.validateDatabaseConfig(tenantInfo)
    }
    
    // Create new tenant database with default config
    database, dbConfigMap, err := h.BaseHandler.DBManager.CreateNewTenantDatabase(tenantInfo.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to create tenant database: %w", err)
    }
    
    // Set DB config in tenant info
    tenantInfo.DBConfig = &domain.DBCredentials{
        Type: dbConfigMap["type"].(string),
        DSN:  dbConfigMap["dsn"].(string),
    }
    
    // Run migrations on the new database
    if err := migration.Migrate(tenantInfo.DBConfig.DSN, tenantInfo.DBConfig.Type, false); err != nil {
        return nil, fmt.Errorf("failed to run migrations: %w", err)
    }
    
    return database, nil
}

func (h *Handlers) validateDatabaseConfig(tenantInfo *TenantInfo) (db.Database, error) {
    dbConfig := h.convertDBCredentialsToMap(tenantInfo.DBConfig)
    
    // Validate connection
    database, err := h.BaseHandler.DBManager.GetDatabaseConnectionWithConfig(tenantInfo.ID, dbConfig)
    if err != nil {
        return nil, fmt.Errorf("invalid database configuration: %w", err)
    }
    
    // Run migrations
    if err := migration.Migrate(tenantInfo.DBConfig.DSN, tenantInfo.DBConfig.Type, false); err != nil {
        return nil, fmt.Errorf("failed to run migrations: %w", err)
    }
    
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

// NewRouter sets up the router with all routes
func NewRouter(mainDB db.Database, dbManager *db.Manager, credentialManager *credentials.Manager) *chi.Mux {
    b := handler.NewBaseHandler(dbManager)
    h := NewHandlers(b, mainDB, credentialManager)

    r := chi.NewRouter()

    // Set up routes
    r.Post("/", h.CreateTenant)
    r.With(tenantMiddleware.WithTenantID).Put("/", h.UpdateTenant)
    r.With(tenantMiddleware.WithTenantID).Get("/me", h.GetTenant)
    r.With(tenantMiddleware.WithTenantID).Get("/{id}", h.GetTenant)
    r.Get("/", h.GetTenants)

    return r
}
