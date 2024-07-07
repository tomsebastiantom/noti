package templateroutes

import (
    "context"
    "encoding/json"
    "net/http"


    "getnoti.com/internal/templates/usecases/create_template"
    "getnoti.com/internal/templates/usecases/get_template"
    "getnoti.com/internal/templates/usecases/get_templates_by_tenant"
    "getnoti.com/internal/templates/usecases/update_template"
    "getnoti.com/internal/templates/repos/implementations"
    "github.com/go-chi/chi/v5"
    "getnoti.com/pkg/db"
)

func NewRouter(database db.Database) *chi.Mux {
    r := chi.NewRouter()

    // Initialize repository
    templateRepo := postgres.NewPostgresTemplateRepository(database)

    // Initialize use case
    createTemplateUseCase := createtemplate.NewCreateTemplateUseCase(templateRepo)
    getTemplateUseCase := gettemplate.NewGetTemplateUseCase(templateRepo)
    getTemplateByTenantUseCase := gettemplates.NewGetTemplatesByTenantUseCase(templateRepo)
    updateTemplateUseCase := updatetemplate.NewUpdateTemplateUseCase(templateRepo)
    // Initialize controller
    createTemplateController := createtemplate.NewCreateTemplateController(createTemplateUseCase)
    getTemplateController := gettemplate.NewGetTemplateController(getTemplateUseCase)
    getTemplateByTenantController := gettemplates.NewGetTemplatesByTenantController(getTemplateByTenantUseCase)
    updateTemplateController := updatetemplate.NewUpdateTemplateController(updateTemplateUseCase)

    // Set up routes
	r.Post("/", commonHandler(createTemplateController.CreateTemplate))
	r.Put("/{id}", commonHandler(updateTemplateController.UpdateTemplate))
	r.Get("/tenants/{id}", commonHandler(getTemplateByTenantController.GetTemplatesByTenant))
	r.Get("/{id}", commonHandler(getTemplateController.GetTemplate ))

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
