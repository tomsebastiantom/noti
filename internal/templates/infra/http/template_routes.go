package templateroutes

import (
    "context"
    "encoding/json"
    "net/http"

    "getnoti.com/internal/templates/repos"
    "getnoti.com/internal/templates/usecases/template"
    "github.com/go-chi/chi/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(db *pgxpool.Pool) *chi.Mux {
    r := chi.NewRouter()

    // Initialize repository
    templateRepo := repos.NewPostgresTemplateRepository(db)

    // Initialize use case
    createTemplateUseCase := template.NewCreateTemplateUseCase(templateRepo)
    getTemplateUseCase := template.NewGetTemplateUseCase(templateRepo)

    // Initialize controller
    templateController := template.NewTemplateController(createTemplateUseCase, getTemplateUseCase)

    // Set up routes
    r.Post("/create", commonHandler(templateController.CreateTemplate))
    r.Get("/get", commonHandler(templateController.GetTemplate))

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
        case func(context.Context, templatedto.CreateTemplateRequest) templatedto.CreateTemplateResponse:
            res = h(ctx, req.(templatedto.CreateTemplateRequest))
        case func(context.Context, templatedto.GetTemplateRequest) templatedto.GetTemplateResponse:
            res = h(ctx, req.(templatedto.GetTemplateRequest))
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
