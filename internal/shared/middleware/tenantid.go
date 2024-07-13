
package middleware

import (
    "context"
    "net/http"
)

type contextKey string

const TenantIDKey contextKey = "tenantID"

func TenantID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenantID := r.Header.Get("X-Tenant-ID")
        if tenantID == "" {
            http.Error(w, "Tenant ID is required", http.StatusBadRequest)
            return
        }
        ctx := context.WithValue(r.Context(), TenantIDKey, tenantID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
