package tenantroutes

import (
	"encoding/json"
	"net/http"

	tenantMiddleware "getnoti.com/internal/shared/middleware"
	"getnoti.com/internal/shared/utils"
	"getnoti.com/internal/tenants/repos"
	"getnoti.com/internal/tenants/repos/implementations"
	"getnoti.com/internal/tenants/usecases/create_tenant"
	"getnoti.com/internal/tenants/usecases/get_tenant"
	"getnoti.com/internal/tenants/usecases/get_tenants"
	"getnoti.com/internal/tenants/usecases/update_tenant"
	"getnoti.com/pkg/db"
	"github.com/go-chi/chi/v5"
)

// Handlers struct to hold all the handlers
type Handlers struct {
	DBManager *db.Manager
	MainDB    db.Database
}

// NewHandlers initializes the Handlers struct with the DBManager, MainDB, and Vault
func NewHandlers(mainDB db.Database, dbManager *db.Manager) *Handlers {
	return &Handlers{
		DBManager: dbManager,
		MainDB:    mainDB,
	}
}

// Helper function to retrieve tenant ID and database connection
func (h *Handlers) getTenantRepo(r *http.Request) (repository.TenantRepository, error) {
	tenantID := r.Context().Value(tenantMiddleware.TenantIDKey).(string)

	// Retrieve the database connection
	database, err := h.DBManager.GetDatabaseConnection(tenantID)
	if err != nil {
		return nil, err
	}
	// Initialize repository
	tenantRepo := repos.NewTenantRepository(h.MainDB, database)
	return tenantRepo, nil
}

// decodeJSONBody is a helper function to decode JSON request bodies
func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return false
	}
	return true
}

// CreateTenant handles the creation of a new tenant
// TODO Unique Name and ID Check Unique Error 
func (h *Handlers) CreateTenant(w http.ResponseWriter, r *http.Request) {
	tenantRepo, err := h.createTenantRepo(r)
	if err != nil {
		handleError(w, "Failed to get tenant repository", err, http.StatusInternalServerError)
		return
	}

	createTenantUseCase := createtenant.NewCreateTenantUseCase(tenantRepo)
	createTenantController := createtenant.NewCreateTenantController(createTenantUseCase)

	var req createtenant.CreateTenantRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	res, err := createTenantController.CreateTenant(r.Context(), req)
	if err != nil {
		handleError(w, "Failed to create tenant", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, res)
}

// UpdateTenant handles the updating of an existing tenant
func (h *Handlers) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	tenantRepo, err := h.getTenantRepo(r)
	if err != nil {
		handleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
		return
	}

	updateTenantUseCase := updatetenant.NewUpdateTenantUseCase(tenantRepo)
	updateTenantController := updatetenant.NewUpdateTenantController(updateTenantUseCase)

	var req updatetenant.UpdateTenantRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	res, err := updateTenantController.UpdateTenant(r.Context(), req)
	if err != nil {
		handleError(w, "Failed to update tenant", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, res)
}

// GetTenant retrieves a tenant by ID
func (h *Handlers) GetTenant(w http.ResponseWriter, r *http.Request) {
	tenantRepo, err := h.getTenantRepo(r)
	if err != nil {
		handleError(w, "Failed to get tenant repository", err, http.StatusInternalServerError)
		return
	}

	getTenantUseCase := gettenant.NewGetTenantUseCase(tenantRepo)
	getTenantController := gettenant.NewGetTenantController(getTenantUseCase)
	id, err := utils.GetTenantIDFromReq(r)
	if err != nil {
		handleError(w, "Failed to get tenant ID", err, http.StatusBadRequest)
		return
	}

	req := gettenant.GetTenantRequest{TenantID: id}

	res, err := getTenantController.GetTenant(r.Context(), req)
	if err != nil {
		handleError(w, "Failed to get tenant", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, res)
}

// GetTenants retrieves all tenants
func (h *Handlers) GetTenants(w http.ResponseWriter, r *http.Request) {
	tenantRepo := repos.NewTenantsRepository(h.MainDB)

	getTenantsUseCase := gettenants.NewGetTenantsUseCase(tenantRepo)
	getTenantsController := gettenants.NewGetTenantsController(getTenantsUseCase)

	res, err := getTenantsController.GetTenants(r.Context(), gettenants.GetTenantsRequest{})
	if err != nil {
		handleError(w, "Failed to get tenants", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, res)
}

// handleError is a helper function to handle errors and send HTTP responses
func handleError(w http.ResponseWriter, message string, err error, statusCode int) {
	http.Error(w, message, statusCode)
	// Log the error here if needed
}

// respondWithJSON is a helper function to send JSON responses
func respondWithJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}



// NewRouter sets up the router with all routes
func NewRouter(mainDB db.Database, dbManager *db.Manager) *chi.Mux {
	h := NewHandlers(mainDB, dbManager)
	r := chi.NewRouter()

	// Set up routes
	r.Post("/", h.CreateTenant)
	r.With(tenantMiddleware.WithTenantID).Put("/", h.UpdateTenant)
	r.With(tenantMiddleware.WithTenantID).Get("/me", h.GetTenant)
	r.With(tenantMiddleware.WithTenantID).Get("/{id}", h.GetTenant)
	r.Get("/", h.GetTenants)

	return r
}



//Create a Base handler and refactor


// package handlers

// import (
//     "context"
//     "encoding/json"
//     "net/http"
//     "getnoti.com/pkg/db"
//     "getnoti.com/internal/shared/middleware"
// )

// type BaseHandler struct {
//     DBManager *db.Manager
//     MainDB    db.Database
// }

// func NewBaseHandler(mainDB db.Database, dbManager *db.Manager) *BaseHandler {
//     return &BaseHandler{
//         DBManager: dbManager,
//         MainDB:    mainDB,
//     }
// }

// func (h *BaseHandler) WithTenantID(next http.HandlerFunc) http.HandlerFunc {
//     return func(w http.ResponseWriter, r *http.Request) {
//         tenantID := r.Header.Get("X-Tenant-ID")
//         if tenantID == "" {
//             tenantID = r.URL.Query().Get("tenant_id")
//         }

//         if tenantID == "" {
//             http.Error(w, "Tenant ID is required", http.StatusBadRequest)
//             return
//         }

//         ctx := context.WithValue(r.Context(), middleware.TenantIDKey, tenantID)
//         next.ServeHTTP(w, r.WithContext(ctx))
//     }
// }

// func (h *BaseHandler) GetTenantDB(r *http.Request) (db.Database, error) {
//     tenantID := r.Context().Value(middleware.TenantIDKey).(string)
//     return h.DBManager.GetDatabaseConnection(tenantID)
// }

// func (h *BaseHandler) DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
//     if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
//         return err
//     }
//     return nil
// }

// func (h *BaseHandler) HandleRequest(w http.ResponseWriter, r *http.Request, handler func(interface{}) (interface{}, error)) {
//     var requestBody interface{}
//     if err := h.DecodeJSONBody(w, r, &requestBody); err != nil {
//         http.Error(w, "Invalid request body", http.StatusBadRequest)
//         return
//     }

//     responseData, err := handler(requestBody)
//     if err != nil {
//         h.HandleError(w, "Request processing failed", err, http.StatusInternalServerError)
//         return
//     }

//     h.RespondWithJSON(w, responseData)
// }

// func (h *BaseHandler) RespondWithJSON(w http.ResponseWriter, data interface{}) {
//     w.Header().Set("Content-Type", "application/json")
//     if err := json.NewEncoder(w).Encode(data); err != nil {
//         http.Error(w, "Failed to encode response", http.StatusInternalServerError)
//     }
// }

// func (h *BaseHandler) HandleError(w http.ResponseWriter, message string, err error, statusCode int) {
//     http.Error(w, message, statusCode)
//     // Log the error here if needed
// }

// NotificationHandler extends BaseHandler with specific types for the notification domain
// type NotificationHandler struct {
//     *BaseHandler
//     GenericCache      *cache.GenericCache
//     QueueManager      *queue.QueueManager
//     WorkerPoolManager *workerpool.WorkerPoolManager
// }

// Using an extended request model along with a base handler is a solid approach for building scalable 
//and maintainable code in your Go application. Here’s how this design can enhance your application, 
//along with some best practices to ensure it remains flexible and easy to maintain over time.