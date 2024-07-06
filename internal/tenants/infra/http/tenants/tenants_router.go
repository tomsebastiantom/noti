package tenantroutes

import (
	"context"
	"encoding/json"
	"net/http"
	"getnoti.com/internal/tenants/repos/implementations"
	"getnoti.com/internal/tenants/usecases/create_tenant"
	"getnoti.com/internal/tenants/usecases/update_tenant"
	"getnoti.com/internal/tenants/usecases/get_tenants"
	"github.com/go-chi/chi/v5"
	"getnoti.com/pkg/db"
)

func NewRouter(database db.Database) *chi.Mux {
	r := chi.NewRouter()

	// Initialize repository
	tenantRepo := postgres.NewPostgresTenantRepository(database)

	// Initialize use cases
	createTenantUseCase := createtenant.NewCreateTenantUseCase(tenantRepo)
	updateTenantUseCase := updatetenant.NewUpdateTenantUseCase(tenantRepo)
	getTenantsUseCase := gettenants.NewGetTenantsUseCase(tenantRepo)

	// Initialize controllers
	createTenantController := createtenant.NewCreateTenantController(createTenantUseCase)
	updateTenantController := updatetenant.NewUpdateTenantController(updateTenantUseCase)
	getTenantsController := gettenants.NewGetTenantsController(getTenantsUseCase)


	// Set up routes
	r.Post("/", CommonHandler(createTenantController.CreateTenant))
	r.Put("/{id}", CommonHandler(updateTenantController.UpdateTenant))
	r.Get("/{id}", CommonHandler(getTenantsController.GetTenants))
	r.Get("/", CommonHandler(getTenantsController.GetTenants))

	return r
}

func CommonHandler(handlerFunc interface{}) http.HandlerFunc {
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

	
		var res interface{}
		switch h := handlerFunc.(type) {
		case func(context.Context, createtenant.CreateTenantRequest) createtenant.CreateTenantResponse:
			res = h(ctx, req.(createtenant.CreateTenantRequest))
		case func(context.Context, updatetenant.UpdateTenantRequest) updatetenant.UpdateTenantResponse:
			res = h(ctx, req.(updatetenant.UpdateTenantRequest))
		case func(context.Context, gettenants.GetTenantsRequest) gettenants.GetTenantsResponse:
			res = h(ctx, req.(gettenants.GetTenantsRequest))
		
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