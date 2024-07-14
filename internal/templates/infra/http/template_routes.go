

package templateroutes

import (
	"context"
	"encoding/json"
	"net/http"

	repository "getnoti.com/internal/templates/repos"
	custom "getnoti.com/internal/shared/middleware"
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

	// Initialize use case
	createTemplateUseCase := createtemplate.NewCreateTemplateUseCase(templateRepo)

	// Initialize controller
	createTemplateController := createtemplate.NewCreateTemplateController(createTemplateUseCase)

	// Handle the request
	commonHandler(createTemplateController.CreateTemplate)(w, r)
}

func (h *Handlers) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	templateRepo, err := h.getTemplateRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	// Initialize use case
	updateTemplateUseCase := updatetemplate.NewUpdateTemplateUseCase(templateRepo)

	// Initialize controller
	updateTemplateController := updatetemplate.NewUpdateTemplateController(updateTemplateUseCase)

	// Handle the request
	commonHandler(updateTemplateController.UpdateTemplate)(w, r)
}

func (h *Handlers) GetTemplate(w http.ResponseWriter, r *http.Request) {
	templateRepo, err := h.getTemplateRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	// Initialize use case
	getTemplateUseCase := gettemplate.NewGetTemplateUseCase(templateRepo)

	// Initialize controller
	getTemplateController := gettemplate.NewGetTemplateController(getTemplateUseCase)

	// Handle the request
	commonHandler(getTemplateController.GetTemplate)(w, r)
}

func (h *Handlers) GetTemplatesByTenant(w http.ResponseWriter, r *http.Request) {
	templateRepo, err := h.getTemplateRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	// Initialize use case
	getTemplatesByTenantUseCase := gettemplates.NewGetTemplatesByTenantUseCase(templateRepo)

	// Initialize controller
	getTemplateByTenantController := gettemplates.NewGetTemplatesByTenantController(getTemplatesByTenantUseCase)

	// Handle the request
	commonHandler(getTemplateByTenantController.GetTemplatesByTenant)(w, r)
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
		case func(context.Context, gettemplate.GetTemplateRequest) gettemplate.GetTemplateResponse:
			res = h(ctx, req.(gettemplate.GetTemplateRequest))
		case func(context.Context, createtemplate.CreateTemplateRequest) createtemplate.CreateTemplateResponse:
			res = h(ctx, req.(createtemplate.CreateTemplateRequest))
		case func(context.Context, updatetemplate.UpdateTemplateRequest) updatetemplate.UpdateTemplateResponse:
			res = h(ctx, req.(updatetemplate.UpdateTemplateRequest))
		case func(context.Context, gettemplates.GetTemplatesByTenantRequest) gettemplates.GetTemplatesByTenantResponse:
			res = h(ctx, req.(gettemplates.GetTemplatesByTenantRequest))
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
