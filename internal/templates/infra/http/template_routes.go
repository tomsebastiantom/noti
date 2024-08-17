package templateroutes

import (
	"encoding/json"
	"net/http"

	custom "getnoti.com/internal/shared/middleware"
	repository "getnoti.com/internal/templates/repos"
	"getnoti.com/internal/templates/repos/implementations"
	"getnoti.com/internal/templates/usecases/create_template"
	"getnoti.com/internal/templates/usecases/get_template"
	"getnoti.com/internal/templates/usecases/get_templates_by_tenant"
	"getnoti.com/internal/templates/usecases/update_template"
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
func (h *Handlers) getTemplateRepo(r *http.Request) (repository.TemplateRepository, error) {
	tenantID := r.Context().Value(custom.TenantIDKey).(string)

	// Retrieve the database connection
	database, err := h.DBManager.GetDatabaseConnection(tenantID)
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
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	createTemplateUseCase := createtemplate.NewCreateTemplateUseCase(templateRepo)
	createTemplateController := createtemplate.NewCreateTemplateController(createTemplateUseCase)

	var req createtemplate.CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := createTemplateController.CreateTemplate(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	templateRepo, err := h.getTemplateRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	updateTemplateUseCase := updatetemplate.NewUpdateTemplateUseCase(templateRepo)
	updateTemplateController := updatetemplate.NewUpdateTemplateController(updateTemplateUseCase)

	var req updatetemplate.UpdateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := updateTemplateController.UpdateTemplate(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) GetTemplate(w http.ResponseWriter, r *http.Request) {
	templateRepo, err := h.getTemplateRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	getTemplateUseCase := gettemplate.NewGetTemplateUseCase(templateRepo)
	getTemplateController := gettemplate.NewGetTemplateController(getTemplateUseCase)

	var req gettemplate.GetTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := getTemplateController.GetTemplate(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) GetTemplatesByTenant(w http.ResponseWriter, r *http.Request) {
	templateRepo, err := h.getTemplateRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	getTemplatesByTenantUseCase := gettemplates.NewGetTemplatesByTenantUseCase(templateRepo)
	getTemplateByTenantController := gettemplates.NewGetTemplatesByTenantController(getTemplatesByTenantUseCase)

	var req gettemplates.GetTemplatesByTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := getTemplateByTenantController.GetTemplatesByTenant(r.Context(), req)
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
	r.Post("/", h.CreateTemplate)
	r.Put("/{id}", h.UpdateTemplate)
	r.Get("/tenants/{id}", h.GetTemplatesByTenant)
	r.Get("/{id}", h.GetTemplate)

	return r
}
