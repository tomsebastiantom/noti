package tenantroutes

import (
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

func (h *Handlers) CreateTenant(w http.ResponseWriter, r *http.Request) {
	tenantRepo, err := h.getNewTenantRepo(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	createTenantUseCase := createtenant.NewCreateTenantUseCase(tenantRepo)
	createTenantController := createtenant.NewCreateTenantController(createTenantUseCase)

	var req createtenant.CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := createTenantController.CreateTenant(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	tenantRepo, err := h.getTenantRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	updateTenantUseCase := updatetenant.NewUpdateTenantUseCase(tenantRepo)
	updateTenantController := updatetenant.NewUpdateTenantController(updateTenantUseCase)

	var req updatetenant.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := updateTenantController.UpdateTenant(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) GetTenant(w http.ResponseWriter, r *http.Request) {
	tenantRepo, err := h.getTenantRepo(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	getTenantUseCase := gettenant.NewGetTenantUseCase(tenantRepo)
	getTenantController := gettenant.NewGetTenantController(getTenantUseCase)

	var req gettenant.GetTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := getTenantController.GetTenant(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) GetTenants(w http.ResponseWriter, r *http.Request) {
	tenantRepo := repos.NewTenantsRepository(h.MainDB)

	getTenantsUseCase := gettenants.NewGetTenantsUseCase(tenantRepo)
	getTenantsController := gettenants.NewGetTenantsController(getTenantsUseCase)

	res, err := getTenantsController.GetTenants(r.Context(), gettenants.GetTenantsRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
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
