package providerroutes

import (
	"context"
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
	templateRepo := repository.NewProviderRepository(database)
	return templateRepo, nil
}

func (h *Handlers) CreateProvider(w http.ResponseWriter, r *http.Request) {
	providerRepo, err := h.getProviderRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	// Initialize use case
	createProviderUseCase := createprovider.NewCreateProviderUseCase(providerRepo)

	// Initialize controller
	createProviderController := createprovider.NewCreateProviderController(createProviderUseCase)

	// Handle the request
	commonHandler(createProviderController.CreateProvider)(w, r)
}

func (h *Handlers) UpdateProvider(w http.ResponseWriter, r *http.Request) {
	providerRepo, err := h.getProviderRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	// Initialize use case
	updateProviderUseCase := updateprovider.NewUpdateProviderUseCase(providerRepo)

	// Initialize controller
	updateProviderController := updateprovider.NewUpdateProviderController(updateProviderUseCase)

	// Handle the request
	commonHandler(updateProviderController.UpdateProvider)(w, r)
}

func (h *Handlers) GetProvider(w http.ResponseWriter, r *http.Request) {
	providerRepo, err := h.getProviderRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	// Initialize use case
	getProviderUseCase := getprovider.NewGetProviderUseCase(providerRepo)

	// Initialize controller
	getProviderController := getprovider.NewGetProviderController(getProviderUseCase)

	// Handle the request
	commonHandler(getProviderController.GetProvider)(w, r)
}

func (h *Handlers) GetProviderByTenant(w http.ResponseWriter, r *http.Request) {
	templateRepo, err := h.getProviderRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	// Initialize use case
	getProvidersUseCase := getproviders.NewGetProvidersUseCase(templateRepo)

	// Initialize controller
	getProvidersController := getproviders.NewGetProvidersController(getProvidersUseCase)

	// Handle the request
	commonHandler(getProvidersController.GetProviders)(w, r)
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

// commonHandler is a generic HTTP handler function that handles requests and responses for different controllers.
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
		case func(context.Context, getprovider.GetProviderRequest) getprovider.GetProviderResponse:
			res = h(ctx, req.(getprovider.GetProviderRequest))
		case func(context.Context, createprovider.CreateProviderRequest) createprovider.CreateProviderResponse:
			res = h(ctx, req.(createprovider.CreateProviderRequest))
		case func(context.Context, updateprovider.UpdateProviderRequest) updateprovider.UpdateProviderResponse:
			res = h(ctx, req.(updateprovider.UpdateProviderRequest))
		case func(context.Context, getproviders.GetProvidersRequest) getproviders.GetProvidersResponse:
			res = h(ctx, req.(getproviders.GetProvidersRequest))
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
