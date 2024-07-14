package userroutes

import (
	"context"
	"encoding/json"
	"net/http"

	repository "getnoti.com/internal/tenants/repos"
	custom "getnoti.com/internal/shared/middleware"
	"getnoti.com/internal/tenants/repos/implementations"
	"getnoti.com/internal/tenants/usecases/create_user"
	"getnoti.com/internal/tenants/usecases/get_users"
	"getnoti.com/internal/tenants/usecases/update_user"
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
func (h *Handlers) getUserRepo(r *http.Request) (repository.UserRepository, error) {
	tenantID := r.Context().Value(custom.TenantIDKey).(string)


	// Retrieve the database connection
	database, err := h.DBManager.GetDatabaseConnection(tenantID)
	if err != nil {
		return nil, err
	}

	// Initialize repository
	userRepo := repos.NewUserRepository(database)
	return userRepo, nil
}

func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	userRepo, err := h.getUserRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	// Initialize use case
	createUserUseCase := createuser.NewCreateUserUseCase(userRepo)

	// Initialize controller
	createUserController := createuser.NewCreateUserController(createUserUseCase)

	// Handle the request
	commonHandler(createUserController.CreateUser)(w, r)
}

func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userRepo, err := h.getUserRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	// Initialize use case
	updateUserUseCase := updateuser.NewUpdateUserUseCase(userRepo)

	// Initialize controller
	updateUserController := updateuser.NewUpdateUserController(updateUserUseCase)

	// Handle the request
	commonHandler(updateUserController.UpdateUser)(w, r)
}

func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	userRepo, err := h.getUserRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	// Initialize use case
	getUsersUseCase := getusers.NewGetUsersUseCase(userRepo)

	// Initialize controller
	getUserController := getusers.NewGetUsersController(getUsersUseCase)

	// Handle the request
	commonHandler(getUserController.GetUsers)(w, r)
}

func (h *Handlers) GetUsers(w http.ResponseWriter, r *http.Request) {
	userRepo, err := h.getUserRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	// Initialize use case
	getUsersUseCase := getusers.NewGetUsersUseCase(userRepo)

	// Initialize controller
	getUserController := getusers.NewGetUsersController(getUsersUseCase)

	// Handle the request
	commonHandler(getUserController.GetUsers)(w, r)
}

// NewRouter sets up the router with all routes
func NewRouter(dbManager *db.Manager) *chi.Mux {
	h := NewHandlers(dbManager)
	r := chi.NewRouter()

	// Set up routes
	r.Post("/", h.CreateUser)
	r.Put("/{id}", h.UpdateUser)
	r.Get("/{id}", h.GetUser)
	r.Get("/", h.GetUsers)

	return r
}

// CommonHandler is a generic HTTP handler function that handles requests and responses for different controllers.
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
		case func(context.Context, createuser.CreateUserRequest) createuser.CreateUserResponse:
			res = h(ctx, req.(createuser.CreateUserRequest))
		case func(context.Context, updateuser.UpdateUserRequest) updateuser.UpdateUserResponse:
			res = h(ctx, req.(updateuser.UpdateUserRequest))
		case func(context.Context, getusers.GetUsersRequest) getusers.GetUsersResponse:
			res = h(ctx, req.(getusers.GetUsersRequest))

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

