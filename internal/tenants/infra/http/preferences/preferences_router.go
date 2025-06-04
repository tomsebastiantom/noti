package preferencesroutes

import (
	"net/http"

	"getnoti.com/internal/shared/handler"
	"getnoti.com/internal/shared/middleware"
	"getnoti.com/internal/shared/utils"
	repository "getnoti.com/internal/tenants/repos"
	repos "getnoti.com/internal/tenants/repos/implementations"
	getuserpreferences "getnoti.com/internal/tenants/usecases/get_user_preferences"
	updateuserpreferences "getnoti.com/internal/tenants/usecases/update_user_preferences"
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

// Helper function to retrieve repositories
func (h *Handlers) getRepos(r *http.Request) (repository.UserPreferenceRepository, repository.UserRepository, error) {
	tenantID := r.Context().Value(middleware.TenantIDKey).(string)

	// Retrieve the database connection
	database, err := h.BaseHandler.DBManager.GetDatabaseConnection(tenantID)
	if err != nil {
		return nil, nil, err
	}

	// Initialize repositories
	userPrefRepo := repos.NewUserPreferenceRepository(database)
	userRepo := repos.NewUserRepository(database)
	
	return userPrefRepo, userRepo, nil
}

// GetUserPreferences retrieves user preferences
func (h *Handlers) GetUserPreferences(w http.ResponseWriter, r *http.Request) {
	userPrefRepo, userRepo, err := h.getRepos(r)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
		return
	}

	getUserPrefsUseCase := getuserpreferences.NewGetUserPreferencesUseCase(userPrefRepo, userRepo)
	getUserPrefsController := getuserpreferences.NewGetUserPreferencesController(getUserPrefsUseCase)

	var req getuserpreferences.GetUserPreferencesRequest
	
	// Get user ID from URL parameter
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		h.BaseHandler.HandleError(w, "User ID is required", nil, http.StatusBadRequest)
		return
	}
	req.UserID = userID

	// Add tenant ID to the request
	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		h.BaseHandler.HandleError(w, "Failed to process tenant ID", err, http.StatusInternalServerError)
		return
	}

	res, err := getUserPrefsController.GetUserPreferences(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to get user preferences", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

// UpdateUserPreferences updates user preferences
func (h *Handlers) UpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	userPrefRepo, userRepo, err := h.getRepos(r)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
		return
	}

	updateUserPrefsUseCase := updateuserpreferences.NewUpdateUserPreferencesUseCase(userPrefRepo, userRepo)
	updateUserPrefsController := updateuserpreferences.NewUpdateUserPreferencesController(updateUserPrefsUseCase)

	var req updateuserpreferences.UpdateUserPreferencesRequest
	if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
		return
	}

	// Get user ID from URL parameter
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		h.BaseHandler.HandleError(w, "User ID is required", nil, http.StatusBadRequest)
		return
	}
	req.UserID = userID

	// Add tenant ID to the request
	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		h.BaseHandler.HandleError(w, "Failed to process tenant ID", err, http.StatusInternalServerError)
		return
	}

	res, err := updateUserPrefsController.UpdateUserPreferences(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to update user preferences", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

// NewRouter sets up the router with all routes
func NewRouter(baseHandler *handler.BaseHandler) http.Handler {
	h := NewHandlers(baseHandler)
	r := chi.NewRouter()

	// Apply the tenant middleware to all routes
	r.Use(middleware.WithTenantID)

	r.Get("/users/{userID}/preferences", h.GetUserPreferences)
	r.Put("/users/{userID}/preferences", h.UpdateUserPreferences)

	return r
}
