package utils

import (
    "fmt"
    "net/http"
    "github.com/go-chi/chi/v5"
    tenantMiddleware "getnoti.com/internal/shared/middleware"
)

// GetTenantIDFromReq is a helper to get tenantId
func GetTenantIDFromReq(r *http.Request) (string, error) {
    id := chi.URLParam(r, "id")
    if id != "" {
        return id, nil
    }

    tenantID, ok := r.Context().Value(tenantMiddleware.TenantIDKey).(string)
    if !ok {
        return "", fmt.Errorf("tenant ID not found in context")
    }
    return tenantID, nil
}

// TenantIDSetter is an interface for structs that can set a tenant ID
type TenantIDSetter interface {
    SetTenantID(string)
}

// AddTenantIDToRequest adds the tenant ID to the request if it implements TenantIDSetter
func AddTenantIDToRequest(r *http.Request, req interface{}) error {
    if setter, ok := req.(TenantIDSetter); ok {
        tenantID, err := GetTenantIDFromReq(r)
        if err != nil {
            return fmt.Errorf("failed to get tenant ID: %w", err)
        }
        setter.SetTenantID(tenantID)
    }
    return nil
}
