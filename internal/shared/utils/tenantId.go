package utils

import (
	"fmt"
	tenantMiddleware "getnoti.com/internal/shared/middleware"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// GetTenantIDFromReq is a helper to get tenantId
func GetIDFromReq(r *http.Request) (string, error) {
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
		tenantID, err := GetIDFromReq(r)
		if err != nil {
			return fmt.Errorf("failed to get tenant ID: %w", err)
		}
		setter.SetTenantID(tenantID)
	}
	return nil
}

type IDSetter interface {
	SetID(string)
}

// AddTenantIDToRequest adds the tenant ID to the request if it implements TenantIDSetter
func AddIDToRequest(r *http.Request, req interface{}) error {
	if setter, ok := req.(IDSetter); ok {
		ID, err := GetIDFromReq(r)
		if err != nil {
			return fmt.Errorf("failed to get tenant ID: %w", err)
		}
		setter.SetID(ID)
	}
	return nil
}
