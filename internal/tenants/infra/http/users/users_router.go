package userroutes

import (
	"context"
	"encoding/json"
	"net/http"
	"getnoti.com/internal/tenants/repos/implementations"
	"getnoti.com/internal/tenants/usecases/create_user"
	"getnoti.com/internal/tenants/usecases/update_user"
	"getnoti.com/internal/tenants/usecases/get_users"
	"github.com/go-chi/chi/v5"
	"getnoti.com/pkg/db"
)

func NewRouter(database db.Database) *chi.Mux {
	r := chi.NewRouter()

	// Initialize repository
	userRepo := postgres.NewPostgresUserRepository(database)

	// Initialize use cases
	createUserUseCase := createuser.NewCreateUserUseCase(userRepo)
	updateUserUseCase := updateuser.NewUpdateUserUseCase(userRepo)
	getUsersUseCase := getusers.NewGetUsersUseCase(userRepo)
	
	// Initialize controllers
	createUserController := createuser.NewCreateUserController(createUserUseCase)
	updateUserController := updateuser.NewUpdateUserController(updateUserUseCase)
	getUserController := getusers.NewGetUsersController(getUsersUseCase)

	// Set up routes
	r.Post("/", CommonHandler(createUserController.CreateUser))
	r.Put("/{id}", CommonHandler(updateUserController.UpdateUser))
	r.Get("/{id}", CommonHandler(getUserController.GetUsers))
	r.Get("/", CommonHandler(getUserController.GetUsers))

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