package logger

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HTTPLogger middleware for logging HTTP requests
func HTTPLogger(logger Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Create a response writer wrapper to capture status code
            ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
            
            // Log request start
            logger.InfoContext(r.Context(), "HTTP request started",
                Method(r.Method),
                URL(r.URL.String()),
                RemoteAddr(r.RemoteAddr),
                UserAgent(r.UserAgent()),
                Protocol(r.Proto),
            )
            
            // Process request
            next.ServeHTTP(ww, r)
            
            // Log request completion
            duration := time.Since(start)
            fields := []Field{
                Method(r.Method),
                URL(r.URL.String()),
                StatusCode(ww.statusCode),
                Latency(duration),
                Int("response_size", ww.size),
            }
            
            // Log level based on status code
            if ww.statusCode >= 500 {
                logger.ErrorContext(r.Context(), "HTTP request completed with server error", fields...)
            } else if ww.statusCode >= 400 {
                logger.WarnContext(r.Context(), "HTTP request completed with client error", fields...)
            } else {
                logger.InfoContext(r.Context(), "HTTP request completed successfully", fields...)
            }
        })
    }
}

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
    http.ResponseWriter
    statusCode int
    size       int
}

func (w *responseWriter) WriteHeader(statusCode int) {
    w.statusCode = statusCode
    w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(data []byte) (int, error) {
    size, err := w.ResponseWriter.Write(data)
    w.size += size
    return size, err
}

// RequestIDMiddleware adds request ID to context and logs
func RequestIDMiddleware(logger Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            requestID := r.Header.Get("X-Request-ID")
            if requestID == "" {
                requestID = generateRequestID()
            }
            
            // Add request ID to response headers
            w.Header().Set("X-Request-ID", requestID)
            
            // Add request ID to context
            ctx := r.Context()
            ctx = context.WithValue(ctx, "request_id", requestID)
            
            // Continue with updated context
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func generateRequestID() string {
    // Simple request ID generation - in production, use a proper UUID library
    return fmt.Sprintf("%d", time.Now().UnixNano())
}