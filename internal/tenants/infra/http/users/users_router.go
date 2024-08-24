package userroutes

import (
	"net/http"

	"getnoti.com/internal/shared/handler"
	"getnoti.com/internal/shared/middleware"
	"getnoti.com/internal/shared/utils"
	repository "getnoti.com/internal/tenants/repos"
	repos "getnoti.com/internal/tenants/repos/implementations"
	createuser "getnoti.com/internal/tenants/usecases/create_user"
	getusers "getnoti.com/internal/tenants/usecases/get_users"
	updateuser "getnoti.com/internal/tenants/usecases/update_user"
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
func (h *Handlers) getUserRepo(r *http.Request) (repository.UserRepository, error) {
	tenantID := r.Context().Value(middleware.TenantIDKey).(string)

	// Retrieve the database connection
	database, err := h.BaseHandler.DBManager.GetDatabaseConnection(tenantID)
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
		h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
		return
	}

	createUserUseCase := createuser.NewCreateUserUseCase(userRepo)
	createUserController := createuser.NewCreateUserController(createUserUseCase)

	var req createuser.CreateUserRequest
	if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
		return
	}
	

	res, err := createUserController.CreateUser(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to create user", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userRepo, err := h.getUserRepo(r)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
		return
	}

	updateUserUseCase := updateuser.NewUpdateUserUseCase(userRepo)
	updateUserController := updateuser.NewUpdateUserController(updateUserUseCase)

	var req updateuser.UpdateUserRequest
	if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
		return
	}

	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		h.BaseHandler.HandleError(w, "Failed to process tenant ID", err, http.StatusInternalServerError)
		return
	}

	res, err := updateUserController.UpdateUser(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to update user", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	userRepo, err := h.getUserRepo(r)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
		return
	}

	getUsersUseCase := getusers.NewGetUsersUseCase(userRepo)
	getUserController := getusers.NewGetUsersController(getUsersUseCase)

	var req getusers.GetUsersRequest
	if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
		return
	}

	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		h.BaseHandler.HandleError(w, "Failed to process tenant ID", err, http.StatusInternalServerError)
		return
	}

	res, err := getUserController.GetUsers(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to get user", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

func (h *Handlers) GetUsers(w http.ResponseWriter, r *http.Request) {
	userRepo, err := h.getUserRepo(r)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
		return
	}

	getUsersUseCase := getusers.NewGetUsersUseCase(userRepo)
	getUserController := getusers.NewGetUsersController(getUsersUseCase)

	req := getusers.GetUsersRequest{}

	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		h.BaseHandler.HandleError(w, "Failed to process tenant ID", err, http.StatusInternalServerError)
		return
	}

	res, err := getUserController.GetUsers(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to get users", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

func NewRouter(dbManager *db.Manager) *chi.Mux {
	b := handler.NewBaseHandler(dbManager)
	h := NewHandlers(b)

	r := chi.NewRouter()

	// Set up routes
	r.Post("/", h.CreateUser)
	r.Put("/{id}", h.UpdateUser)
	r.Get("/{id}", h.GetUser)
	r.Get("/", h.GetUsers)

	return r
}
