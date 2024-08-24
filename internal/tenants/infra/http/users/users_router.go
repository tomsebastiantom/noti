package userroutes

import (
	"encoding/json"
	"net/http"

	"getnoti.com/internal/shared/middleware"
	"getnoti.com/internal/shared/utils"
	repository "getnoti.com/internal/tenants/repos"
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
	tenantID := r.Context().Value(middleware.TenantIDKey).(string)

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

	createUserUseCase := createuser.NewCreateUserUseCase(userRepo)
	createUserController := createuser.NewCreateUserController(createUserUseCase)

	var req createuser.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Add tenant ID to the request if not present
	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		http.Error(w, "Failed to process tenant ID", http.StatusInternalServerError)
		return
	}

	res, err := createUserController.CreateUser(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userRepo, err := h.getUserRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	updateUserUseCase := updateuser.NewUpdateUserUseCase(userRepo)
	updateUserController := updateuser.NewUpdateUserController(updateUserUseCase)

	var req updateuser.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Add tenant ID to the request if not present
	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		http.Error(w, "Failed to process tenant ID", http.StatusInternalServerError)
		return
	}

	res, err := updateUserController.UpdateUser(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	userRepo, err := h.getUserRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	getUsersUseCase := getusers.NewGetUsersUseCase(userRepo)
	getUserController := getusers.NewGetUsersController(getUsersUseCase)

	var req getusers.GetUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Add tenant ID to the request if not present
	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		http.Error(w, "Failed to process tenant ID", http.StatusInternalServerError)
		return
	}

	res, err := getUserController.GetUsers(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) GetUsers(w http.ResponseWriter, r *http.Request) {
	userRepo, err := h.getUserRepo(r)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	getUsersUseCase := getusers.NewGetUsersUseCase(userRepo)
	getUserController := getusers.NewGetUsersController(getUsersUseCase)

	req := getusers.GetUsersRequest{}

	// Add tenant ID to the request if not present
	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		http.Error(w, "Failed to process tenant ID", http.StatusInternalServerError)
		return
	}

	res, err := getUserController.GetUsers(r.Context(), req)
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
	r.Post("/", h.CreateUser)
	r.Put("/{id}", h.UpdateUser)
	r.Get("/{id}", h.GetUser)
	r.Get("/", h.GetUsers)

	return r
}
