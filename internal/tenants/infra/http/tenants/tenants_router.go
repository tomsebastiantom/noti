package tenantroutes

import (
	"context"
	"encoding/json"
	"net/http"

	tenantMiddleware "getnoti.com/internal/shared/middleware"
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

// Helper function to retrieve tenant ID and database connection
func (h *Handlers) getNewTenantRepo(r *http.Request) (repository.TenantRepository, error) {
	tenantID := r.Context().Value(tenantMiddleware.TenantIDKey).(string)

	// Retrieve the database connection
	database, err := h.DBManager.CreateNewTenantDatabase(tenantID)
	if err != nil {
		return nil, err
	}
	//create a new db in sql db
	// Initialize repository
	tenantRepo := repos.NewTenantRepository(h.MainDB, database)
	return tenantRepo, nil
}

func (h *Handlers) CreateTenant(w http.ResponseWriter, r *http.Request) {
	tenantRepo, err := h.getNewTenantRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	// Initialize use case
	createTenantUseCase := createtenant.NewCreateTenantUseCase(tenantRepo)

	// Initialize controller
	createTenantController := createtenant.NewCreateTenantController(createTenantUseCase)

	// Handle the request
	commonHandler(createTenantController.CreateTenant)(w, r)
}

func (h *Handlers) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	tenantRepo, err := h.getTenantRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	// Initialize use case
	updateTenantUseCase := updatetenant.NewUpdateTenantUseCase(tenantRepo)

	// Initialize controller
	updateTenantController := updatetenant.NewUpdateTenantController(updateTenantUseCase)

	// Handle the request
	commonHandler(updateTenantController.UpdateTenant)(w, r)
}

func (h *Handlers) GetTenant(w http.ResponseWriter, r *http.Request) {
	tenantRepo, err := h.getTenantRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	// Initialize use case
	getTenantUseCase := gettenant.NewGetTenantUseCase(tenantRepo)

	// Initialize controller
	getTenantController := gettenant.NewGetTenantController(getTenantUseCase)

	// Handle the request
	commonHandler(getTenantController.GetTenant)(w, r)
}

func (h *Handlers) GetTenants(w http.ResponseWriter, r *http.Request) {
	tenantRepo := repos.NewTenantsRepository(h.MainDB)

	// Initialize use case
	getTenantsUseCase := gettenants.NewGetTenantsUseCase(tenantRepo)

	// Initialize controller
	getTenantsController := gettenants.NewGetTenantsController(getTenantsUseCase)
	// log.Printf("GetTenants handler type: %T", getTenantsController.GetTenants)
	// Handle the request
	commonHandler(getTenantsController.GetTenants)(w, r)
}

// NewRouter sets up the router with all routes
func NewRouter(mainDB db.Database, dbManager *db.Manager) *chi.Mux {
	h := NewHandlers(mainDB, dbManager)
	r := chi.NewRouter()

	// Set up routes
	r.Post("/", h.CreateTenant)
	r.Put("/", h.UpdateTenant)
	r.With(tenantMiddleware.WithTenantID).Get("/me", h.GetTenant)
	r.Get("/", h.GetTenants)

	return r
}

// CommonHandler is a generic HTTP handler function that handles requests and responses for different controllers.
func commonHandler(handlerFunc interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Decode the request body into the appropriate request type
		var req interface{}
		if r.Method != http.MethodGet && r.Method != http.MethodDelete {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
		}

		// Call the handler function with the context and request
		// Call the handler function with the context and request
		var res interface{}
		var err error

		switch h := handlerFunc.(type) {
		case func(context.Context, createtenant.CreateTenantRequest) (createtenant.CreateTenantResponse, error):

			res, err = h(ctx, req.(createtenant.CreateTenantRequest))
		case func(context.Context, updatetenant.UpdateTenantRequest) (updatetenant.UpdateTenantResponse, error):

			res, err = h(ctx, req.(updatetenant.UpdateTenantRequest))
		case func(context.Context, gettenants.GetTenantsRequest) (gettenants.GetTenantsResponse, error):

			res, err = h(ctx, gettenants.GetTenantsRequest{})

		case func(context.Context, gettenant.GetTenantRequest) (gettenant.GetTenantResponse, error):

			res, err = h(ctx, req.(gettenant.GetTenantRequest))
		default:

			http.Error(w, "Unsupported handler function", http.StatusInternalServerError)
			return
		}

		if err != nil {

			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Encode the response and write it to the response writer
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
