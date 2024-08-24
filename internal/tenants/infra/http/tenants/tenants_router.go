package tenantroutes

import (
	"net/http"

	"getnoti.com/internal/shared/handler"
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


type Handlers struct {
	BaseHandler *handler.BaseHandler
}

func NewHandlers(baseHandler *handler.BaseHandler) *Handlers {
	return &Handlers{
		BaseHandler: baseHandler,
	}
}

// Helper function to retrieve tenant ID and database connection
func (h *Handlers) getTenantRepo(r *http.Request) (repository.TenantRepository, error) {
	tenantID := r.Context().Value(tenantMiddleware.TenantIDKey).(string)

	// Retrieve the database connection
	database, err := h.BaseHandler.DBManager.GetDatabaseConnection(tenantID)
	if err != nil {
		return nil, err
	}
	// Initialize repository
	tenantRepo := repos.NewTenantRepository(h.BaseHandler.MainDB, database)
	return tenantRepo, nil
}

// CreateTenant handles the creation of a new tenant
// TODO Unique Name and ID Check Unique Error
func (h *Handlers) CreateTenant(w http.ResponseWriter, r *http.Request) {
	tenantRepo, err := h.createTenantRepo(r)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to get tenant repository", err, http.StatusInternalServerError)
		return
	}

	createTenantUseCase := createtenant.NewCreateTenantUseCase(tenantRepo)
	createTenantController := createtenant.NewCreateTenantController(createTenantUseCase)

	var req createtenant.CreateTenantRequest
	if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
		return
	}

	res, err := createTenantController.CreateTenant(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to create tenant", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

// UpdateTenant handles the updating of an existing tenant
func (h *Handlers) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	tenantRepo, err := h.getTenantRepo(r)
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

	res, err := updateTenantController.UpdateTenant(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to update tenant", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

// GetTenant retrieves a tenant by ID
func (h *Handlers) GetTenant(w http.ResponseWriter, r *http.Request) {
	tenantRepo, err := h.getTenantRepo(r)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to get tenant repository", err, http.StatusInternalServerError)
		return
	}

	getTenantUseCase := gettenant.NewGetTenantUseCase(tenantRepo)
	getTenantController := gettenant.NewGetTenantController(getTenantUseCase)
	id, err := utils.GetTenantIDFromReq(r)
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
	tenantRepo := repos.NewTenantsRepository(h.BaseHandler.MainDB)

	getTenantsUseCase := gettenants.NewGetTenantsUseCase(tenantRepo)
	getTenantsController := gettenants.NewGetTenantsController(getTenantsUseCase)

	res, err := getTenantsController.GetTenants(r.Context(), gettenants.GetTenantsRequest{})
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to get tenants", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

// NewRouter sets up the router with all routes
func NewRouter(mainDB db.Database, dbManager *db.Manager) *chi.Mux {
	b := handler.NewBaseHandler(mainDB, dbManager)
	h := NewHandlers(b)

	r := chi.NewRouter()

	// Set up routes
	r.Post("/", h.CreateTenant)
	r.With(tenantMiddleware.WithTenantID).Put("/", h.UpdateTenant)
	r.With(tenantMiddleware.WithTenantID).Get("/me", h.GetTenant)
	r.With(tenantMiddleware.WithTenantID).Get("/{id}", h.GetTenant)
	r.Get("/", h.GetTenants)

	return r
}
