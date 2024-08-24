package templateroutes

import (
	"net/http"

	"getnoti.com/internal/shared/handler"
	"getnoti.com/internal/shared/middleware"
	"getnoti.com/internal/shared/utils"
	repository "getnoti.com/internal/templates/repos"
	"getnoti.com/internal/templates/repos/implementations"
	"getnoti.com/internal/templates/usecases/create_template"
	"getnoti.com/internal/templates/usecases/get_template"
	"getnoti.com/internal/templates/usecases/get_templates_by_tenant"
	"getnoti.com/internal/templates/usecases/update_template"
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
func (h *Handlers) getTemplateRepo(r *http.Request) (repository.TemplateRepository, error) {
	tenantID := r.Context().Value(middleware.TenantIDKey).(string)

	// Retrieve the database connection
	database, err := h.BaseHandler.DBManager.GetDatabaseConnection(tenantID)
	if err != nil {
		return nil, err
	}

	// Initialize repository
	templateRepo := repos.NewTemplateRepository(database)
	return templateRepo, nil
}

func (h *Handlers) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	templateRepo, err := h.getTemplateRepo(r)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
		return
	}

	createTemplateUseCase := createtemplate.NewCreateTemplateUseCase(templateRepo)
	createTemplateController := createtemplate.NewCreateTemplateController(createTemplateUseCase)

	var req createtemplate.CreateTemplateRequest
	if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
		return
	}

	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		h.BaseHandler.HandleError(w, "Failed to process tenant ID", err, http.StatusInternalServerError)
		return
	}

	res, err := createTemplateController.CreateTemplate(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to create template", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

func (h *Handlers) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	templateRepo, err := h.getTemplateRepo(r)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
		return
	}

	updateTemplateUseCase := updatetemplate.NewUpdateTemplateUseCase(templateRepo)
	updateTemplateController := updatetemplate.NewUpdateTemplateController(updateTemplateUseCase)

	var req updatetemplate.UpdateTemplateRequest
	if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
		return
	}

	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		h.BaseHandler.HandleError(w, "Failed to process tenant ID", err, http.StatusInternalServerError)
		return
	}

	res, err := updateTemplateController.UpdateTemplate(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to update template", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

func (h *Handlers) GetTemplate(w http.ResponseWriter, r *http.Request) {
	templateRepo, err := h.getTemplateRepo(r)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
		return
	}

	getTemplateUseCase := gettemplate.NewGetTemplateUseCase(templateRepo)
	getTemplateController := gettemplate.NewGetTemplateController(getTemplateUseCase)

	var req gettemplate.GetTemplateRequest
	if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
		return
	}

	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		h.BaseHandler.HandleError(w, "Failed to process tenant ID", err, http.StatusInternalServerError)
		return
	}

	res, err := getTemplateController.GetTemplate(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to get template", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

func (h *Handlers) GetTemplatesByTenant(w http.ResponseWriter, r *http.Request) {
	templateRepo, err := h.getTemplateRepo(r)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
		return
	}

	getTemplatesByTenantUseCase := gettemplates.NewGetTemplatesByTenantUseCase(templateRepo)
	getTemplateByTenantController := gettemplates.NewGetTemplatesByTenantController(getTemplatesByTenantUseCase)

	var req gettemplates.GetTemplatesByTenantRequest
	if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
		return
	}

	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		h.BaseHandler.HandleError(w, "Failed to process tenant ID", err, http.StatusInternalServerError)
		return
	}

	res, err := getTemplateByTenantController.GetTemplatesByTenant(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to get templates by tenant", err, http.StatusInternalServerError)
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
	r.Post("/", h.CreateTemplate)
	r.Put("/{id}", h.UpdateTemplate)
	r.Get("/tenants/{id}", h.GetTemplatesByTenant)
	r.Get("/{id}", h.GetTemplate)

	return r
}
