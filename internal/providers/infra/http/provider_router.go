package providerroutes

import (
	"encoding/json"
	"net/http"

	"getnoti.com/internal/providers/repos"
	"getnoti.com/internal/providers/repos/implementations"
	"getnoti.com/internal/providers/usecases/create_provider"
	"getnoti.com/internal/providers/usecases/get_provider"
	"getnoti.com/internal/providers/usecases/get_providers"
	"getnoti.com/internal/providers/usecases/update_provider"
	custom "getnoti.com/internal/shared/middleware"
	"getnoti.com/pkg/db"
	"github.com/go-chi/chi/v5"
)

// Handlers struct to hold all the handlers
type Handlers struct {
	DBManager *db.Manager
}

// NewHandlers initializes the Handlers struct with the DBManager
func NewHandlers(dbManager *db.Manager) *Handlers {
	return &Handlers{
		DBManager: dbManager,
	}
}

// Helper function to retrieve tenant ID and database connection
func (h *Handlers) getProviderRepo(r *http.Request) (repos.ProviderRepository, error) {
	tenantID := r.Context().Value(custom.TenantIDKey).(string)

	// Retrieve the database connection
	database, err := h.DBManager.GetDatabaseConnection(tenantID)
	if err != nil {
		return nil, err
	}

	// Initialize repository
	providerRepo := repository.NewProviderRepository(database)
	return providerRepo, nil
}

func (h *Handlers) CreateProvider(w http.ResponseWriter, r *http.Request) {
	providerRepo, err := h.getProviderRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	createProviderUseCase := createprovider.NewCreateProviderUseCase(providerRepo)
	createProviderController := createprovider.NewCreateProviderController(createProviderUseCase)

	var req createprovider.CreateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := createProviderController.CreateProvider(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) UpdateProvider(w http.ResponseWriter, r *http.Request) {
	providerRepo, err := h.getProviderRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	updateProviderUseCase := updateprovider.NewUpdateProviderUseCase(providerRepo)
	updateProviderController := updateprovider.NewUpdateProviderController(updateProviderUseCase)

	var req updateprovider.UpdateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := updateProviderController.UpdateProvider(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) GetProvider(w http.ResponseWriter, r *http.Request) {
	providerRepo, err := h.getProviderRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	getProviderUseCase := getprovider.NewGetProviderUseCase(providerRepo)
	getProviderController := getprovider.NewGetProviderController(getProviderUseCase)

	var req getprovider.GetProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := getProviderController.GetProvider(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) GetProviderByTenant(w http.ResponseWriter, r *http.Request) {
	providerRepo, err := h.getProviderRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	getProvidersUseCase := getproviders.NewGetProvidersUseCase(providerRepo)
	getProvidersController := getproviders.NewGetProvidersController(getProvidersUseCase)

	var req getproviders.GetProvidersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := getProvidersController.GetProviders(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// NewRouter sets up the router with all routes
func NewRouter(dbManager *db.Manager) *chi.Mux {
	h := NewHandlers(dbManager)
	r := chi.NewRouter()

	// Set up routes
	r.Post("/", h.CreateProvider)
	r.Put("/{id}", h.UpdateProvider)
	r.Get("/tenants/{id}", h.GetProviderByTenant)
	r.Get("/{id}", h.GetProvider)

	return r
}