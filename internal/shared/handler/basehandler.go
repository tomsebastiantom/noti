package handler

import (
	"encoding/json"
	"getnoti.com/internal/shared/middleware"
	"getnoti.com/pkg/db"
	"net/http"
)

type BaseHandler struct {
	DBManager *db.Manager
	MainDB    db.Database
}

func NewBaseHandler(mainDB db.Database, dbManager *db.Manager) *BaseHandler {
	return &BaseHandler{
		DBManager: dbManager,
		MainDB:    mainDB,
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
	// Log the error here if needed
}

//Create a Base handler and refactor

// func (h *BaseHandler) WithTenantID(next http.HandlerFunc) http.HandlerFunc {
//     return func(w http.ResponseWriter, r *http.Request) {
//         tenantID := r.Header.Get("X-Tenant-ID")
//         if tenantID == "" {
//             tenantID = r.URL.Query().Get("tenant_id")
//         }

//         if tenantID == "" {
//             http.Error(w, "Tenant ID is required", http.StatusBadRequest)
//             return
//         }

//         ctx := context.WithValue(r.Context(), middleware.TenantIDKey, tenantID)
//         next.ServeHTTP(w, r.WithContext(ctx))
//     }
// }

// NotificationHandler extends BaseHandler with specific types for the notification domain
// type NotificationHandler struct {
//     *BaseHandler
//     GenericCache      *cache.GenericCache
//     QueueManager      *queue.QueueManager
//     WorkerPoolManager *workerpool.WorkerPoolManager
// }

// Using an extended request model along with a base handler is a solid approach for building scalable
//and maintainable code in your Go application. Here’s how this design can enhance your application,
//along with some best practices to ensure it remains flexible and easy to maintain over time.
