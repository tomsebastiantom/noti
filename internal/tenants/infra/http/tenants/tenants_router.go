package tenantroutes

import (
	"context"
	"encoding/json"
	"net/http"

	// repository "getnoti.com/internal/tenants/repos"
	// custom "getnoti.com/internal/shared/middleware"
	custom "getnoti.com/internal/shared/middleware"
	"getnoti.com/internal/tenants/repos"
	"getnoti.com/internal/tenants/repos/implementations"
	"getnoti.com/internal/tenants/usecases/create_tenant"
	"getnoti.com/internal/tenants/usecases/get_tenants"
	"getnoti.com/internal/tenants/usecases/update_tenant"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/vault"
	"github.com/go-chi/chi/v5"
)

// Handlers struct to hold all the handlers
type Handlers struct {
	DBManager *db.Manager
	MainDB    db.Database
	Vault     *vault.VaultConfig
}

// NewHandlers initializes the Handlers struct with the DBManager, MainDB, and Vault
func NewHandlers(mainDB db.Database, dbManager *db.Manager, vault *vault.VaultConfig) *Handlers {
	return &Handlers{
		DBManager: dbManager,
		MainDB:    mainDB,
		Vault:     vault,
	}
}

// Helper function to retrieve tenant ID and database connection
func (h *Handlers) getTenantRepo(r *http.Request) (repository.TenantRepository, error) {
	tenantID := r.Context().Value(custom.TenantIDKey).(string)

	// Retrieve the database connection
	database, err := h.DBManager.GetDatabaseConnection(tenantID)
	if err != nil {
		return nil, err
	}

	// Initialize repository
	tenantRepo := repos.NewTenantRepository(h.MainDB, database, h.Vault)
	return tenantRepo, nil
}

func (h *Handlers) CreateTenant(w http.ResponseWriter, r *http.Request) {
	tenantRepo, err := h.getTenantRepo(r)
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
	getTenantsUseCase := gettenants.NewGetTenantsUseCase(tenantRepo)

	// Initialize controller
	getTenantController := gettenants.NewGetTenantsController(getTenantsUseCase)

	// Handle the request
	commonHandler(getTenantController.GetTenants)(w, r)
}

func (h *Handlers) GetTenants(w http.ResponseWriter, r *http.Request) {
	tenantRepo, err := h.getTenantRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	// Initialize use case
	getTenantsUseCase := gettenants.NewGetTenantsUseCase(tenantRepo)

	// Initialize controller
	getTenantController := gettenants.NewGetTenantsController(getTenantsUseCase)

	// Handle the request
	commonHandler(getTenantController.GetTenants)(w, r)
}

// NewRouter sets up the router with all routes
func NewRouter(mainDB db.Database, dbManager *db.Manager, vault *vault.VaultConfig) *chi.Mux {
	h := NewHandlers(mainDB, dbManager, vault)
	r := chi.NewRouter()

	// Set up routes
	r.Post("/", h.CreateTenant)
	r.Put("/{id}", h.UpdateTenant)
	r.Get("/{id}", h.GetTenant)
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
		var res interface{}
		switch h := handlerFunc.(type) {
		case func(context.Context, createtenant.CreateTenantRequest) createtenant.CreateTenantResponse:
			res = h(ctx, req.(createtenant.CreateTenantRequest))
		case func(context.Context, updatetenant.UpdateTenantRequest) updatetenant.UpdateTenantResponse:
			res = h(ctx, req.(updatetenant.UpdateTenantRequest))
		case func(context.Context, gettenants.GetTenantsRequest) gettenants.GetTenantsResponse:
			res = h(ctx, req.(gettenants.GetTenantsRequest))
		default:
			http.Error(w, "Unsupported handler function", http.StatusInternalServerError)
			return
		}

		// Encode the response and write it to the response writer
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
