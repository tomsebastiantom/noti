package middleware

import (
    "context"
    "net/http"
    "github.com/go-chi/chi/v5"
)

func ExtractIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if userID := chi.URLParam(r, "userID"); userID != "" {
            ctx := context.WithValue(r.Context(), "userID", userID)
            r = r.WithContext(ctx)
        }
        if roleID := chi.URLParam(r, "roleID"); roleID != "" {
            ctx := context.WithValue(r.Context(), "roleID", roleID)
            r = r.WithContext(ctx)
        }
        if groupID := chi.URLParam(r, "groupID"); groupID != "" {
            ctx := context.WithValue(r.Context(), "groupID", groupID)
            r = r.WithContext(ctx)
        }
        next.ServeHTTP(w, r)
    })
}