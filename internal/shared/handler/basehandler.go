package handler

import (
	"encoding/json"
	"getnoti.com/internal/shared/middleware"
	"getnoti.com/pkg/db"
	"net/http"
)

type BaseHandler struct {
	DBManager *db.Manager
}

func NewBaseHandler(dbManager *db.Manager) *BaseHandler {
	return &BaseHandler{
		DBManager: dbManager,
	}
}

func (h *BaseHandler) GetTenantDB(r *http.Request) (db.Database, error) {
	tenantID := r.Context().Value(middleware.TenantIDKey).(string)
	return h.DBManager.GetDatabaseConnection(tenantID)
}

func (h *BaseHandler) DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return false
	}
	return true
}

func (h *BaseHandler) RespondWithJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *BaseHandler) HandleError(w http.ResponseWriter, message string, err error, statusCode int) {
	http.Error(w, message+" "+"Error is"+err.Error(), statusCode)
}
