package providerroutes

import (
	"net/http"

	"getnoti.com/internal/providers/repos"
	"getnoti.com/internal/providers/repos/implementations"
	"getnoti.com/internal/providers/usecases/create_provider"
	"getnoti.com/internal/providers/usecases/get_provider"
	"getnoti.com/internal/providers/usecases/get_providers"
	"getnoti.com/internal/providers/usecases/update_provider"
	"getnoti.com/internal/shared/handler"
	"getnoti.com/internal/shared/middleware"
	"getnoti.com/internal/shared/utils"
	"getnoti.com/pkg/db"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	BaseHandler *handler.BaseHandler
}

func NewHandlers(baseHandler *handler.BaseHandler) *Handlers {
	return &Handlers{
		BaseHandler: baseHandler,
	}
}

// Helper function to retrieve tenant ID and database connection
func (h *Handlers) getProviderRepo(r *http.Request) (repos.ProviderRepository, error) {
	tenantID := r.Context().Value(middleware.TenantIDKey).(string)

	// Retrieve the database connection
	database, err := h.BaseHandler.DBManager.GetDatabaseConnection(tenantID)
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
		h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
		return
	}

	createProviderUseCase := createprovider.NewCreateProviderUseCase(providerRepo)
	createProviderController := createprovider.NewCreateProviderController(createProviderUseCase)

	var req createprovider.CreateProviderRequest
	if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
		return
	}

	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		h.BaseHandler.HandleError(w, "Failed to process tenant ID", err, http.StatusInternalServerError)
		return
	}

	res, err := createProviderController.CreateProvider(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to create provider", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

func (h *Handlers) UpdateProvider(w http.ResponseWriter, r *http.Request) {
	providerRepo, err := h.getProviderRepo(r)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
		return
	}

	updateProviderUseCase := updateprovider.NewUpdateProviderUseCase(providerRepo)
	updateProviderController := updateprovider.NewUpdateProviderController(updateProviderUseCase)

	var req updateprovider.UpdateProviderRequest
	if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
		return
	}

	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		h.BaseHandler.HandleError(w, "Failed to process tenant ID", err, http.StatusInternalServerError)
		return
	}

	res, err := updateProviderController.UpdateProvider(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to update provider", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

func (h *Handlers) GetProvider(w http.ResponseWriter, r *http.Request) {
	providerRepo, err := h.getProviderRepo(r)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
		return
	}

	getProviderUseCase := getprovider.NewGetProviderUseCase(providerRepo)
	getProviderController := getprovider.NewGetProviderController(getProviderUseCase)

	var req getprovider.GetProviderRequest
	if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
		return
	}

	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		h.BaseHandler.HandleError(w, "Failed to process tenant ID", err, http.StatusInternalServerError)
		return
	}

	res, err := getProviderController.GetProvider(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to get provider", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

func (h *Handlers) GetProviderByTenant(w http.ResponseWriter, r *http.Request) {
	providerRepo, err := h.getProviderRepo(r)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
		return
	}

	getProvidersUseCase := getproviders.NewGetProvidersUseCase(providerRepo)
	getProvidersController := getproviders.NewGetProvidersController(getProvidersUseCase)

	var req getproviders.GetProvidersRequest
	if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
		return
	}

	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		h.BaseHandler.HandleError(w, "Failed to process tenant ID", err, http.StatusInternalServerError)
		return
	}

	res, err := getProvidersController.GetProviders(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to get providers", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

// NewRouter sets up the router with all routes
func NewRouter(dbManager *db.Manager) *chi.Mux {
	b := handler.NewBaseHandler(dbManager)
	h := NewHandlers(b)

	r := chi.NewRouter()

	// Set up routes
	r.Post("/", h.CreateProvider)
	r.Put("/{id}", h.UpdateProvider)
	r.Get("/tenants/{id}", h.GetProviderByTenant)
	r.Get("/{id}", h.GetProvider)

	return r
}
