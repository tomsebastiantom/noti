package errors

import (
    "encoding/json"
    "net/http"
    "runtime/debug"
)

// ErrorHandler middleware converts AppErrors to HTTP responses
func ErrorHandler() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    appErr := New(ErrCodeInternal).
                        WithContext(r.Context()).
                        WithOperation(r.URL.Path).
                        WithMessagef("Panic recovered: %v", err).
                        WithSeverity(SeverityCritical).
                        WithDetail("panic", err).
                        WithDetail("stack", string(debug.Stack())).
                        Build()
                    
                    writeErrorResponse(w, appErr)
                }
            }()
            
            next.ServeHTTP(w, r)
        })
    }
}

// ErrorResponse represents the JSON error response
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code      ErrorCode   `json:"code"`
    Message   string      `json:"message"`
    TraceID   string      `json:"trace_id,omitempty"`
    RequestID string      `json:"request_id,omitempty"`
    Details   interface{} `json:"details,omitempty"`
    Timestamp string      `json:"timestamp"`
}

func writeErrorResponse(w http.ResponseWriter, err *AppError) {
    statusCode := err.GetHTTPStatus()
    
    response := ErrorResponse{
        Error: ErrorDetail{
            Code:      err.Code,
            TraceID:   err.TraceID,
            RequestID: err.RequestID,
            Timestamp: err.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
        },
    }
    
    // Use public message if available, otherwise use internal message
    if err.PublicMessage != "" {
        response.Error.Message = err.PublicMessage
    } else {
        response.Error.Message = err.Message
    }
    
    // Include details for validation errors and when appropriate
    if err.Code == ErrCodeValidation && len(err.Details) > 0 {
        response.Error.Details = err.Details
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Request-ID", err.RequestID)
    if err.TraceID != "" {
        w.Header().Set("X-Trace-ID", err.TraceID)
    }
    
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}

// WriteError writes an AppError as HTTP response
func WriteError(w http.ResponseWriter, err *AppError) {
    writeErrorResponse(w, err)
}

// WriteErrorFromCode creates and writes an error from code
func WriteErrorFromCode(w http.ResponseWriter, r *http.Request, code ErrorCode, message string) {
    err := New(code).
        WithContext(r.Context()).
        WithOperation(r.URL.Path).
        WithMessage(message).
        Build()
    
    writeErrorResponse(w, err)
}