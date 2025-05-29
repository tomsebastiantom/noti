package errors

import (
    "context"
    "fmt"
    "runtime"
    "time"
    
    "go.opentelemetry.io/otel/trace"
)

// ErrorCode represents application-specific error codes
type ErrorCode string

// Severity levels for errors
type Severity string

const (
    SeverityLow      Severity = "low"
    SeverityMedium   Severity = "medium"
    SeverityHigh     Severity = "high"
    SeverityCritical Severity = "critical"
)

// ErrorCategory for grouping errors
type ErrorCategory string

const (
    CategoryInfrastructure ErrorCategory = "infrastructure"
    CategoryBusiness       ErrorCategory = "business"
    CategorySecurity       ErrorCategory = "security"
    CategoryPerformance    ErrorCategory = "performance"
    CategorySystem         ErrorCategory = "system"
)

// Frame represents a stack frame
type Frame struct {
    Function string `json:"function"`
    File     string `json:"file"`
    Line     int    `json:"line"`
}

// AppError represents a structured application error with full tracing support
type AppError struct {
    // Core error information
    Code        ErrorCode         `json:"code"`
    Message     string            `json:"message"`
    Details     map[string]any    `json:"details,omitempty"`
    Category    ErrorCategory     `json:"category"`
    
    // Error chain
    Cause       error             `json:"-"`
    
    // Observability
    TraceID     string            `json:"trace_id,omitempty"`
    SpanID      string            `json:"span_id,omitempty"`
    
    // Context
    Service     string            `json:"service"`
    Operation   string            `json:"operation"`
    UserID      string            `json:"user_id,omitempty"`
    RequestID   string            `json:"request_id,omitempty"`
    SessionID   string            `json:"session_id,omitempty"`
    
    // Metadata
    Timestamp   time.Time         `json:"timestamp"`
    Severity    Severity          `json:"severity"`
    Retryable   bool              `json:"retryable"`
    
    // Stack trace
    StackTrace  []Frame           `json:"-"`
    
    // User-facing information
    PublicMessage string          `json:"public_message,omitempty"`
    
    // Additional metadata
    Tags        map[string]string `json:"tags,omitempty"`
    Metrics     map[string]float64 `json:"metrics,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for error chain support
func (e *AppError) Unwrap() error {
    return e.Cause
}

// Is implements error matching for errors.Is()
func (e *AppError) Is(target error) bool {
    if t, ok := target.(*AppError); ok {
        return e.Code == t.Code
    }
    return false
}

// GetHTTPStatus returns appropriate HTTP status code
func (e *AppError) GetHTTPStatus() int {
    switch e.Code {
    case ErrCodeNotFound:
        return 404
    case ErrCodeUnauthorized:
        return 401
    case ErrCodeForbidden:
        return 403
    case ErrCodeValidation:
        return 400
    case ErrCodeConflict:
        return 409
    case ErrCodeRateLimit:
        return 429
    case ErrCodeUnavailable:
        return 503
    case ErrCodeTimeout:
        return 408
    default:
        return 500
    }
}

// IsRetryable returns whether the error is retryable
func (e *AppError) IsRetryable() bool {
    return e.Retryable
}

// GetSeverity returns the error severity
func (e *AppError) GetSeverity() Severity {
    return e.Severity
}

// WithTag adds a tag to the error
func (e *AppError) WithTag(key, value string) *AppError {
    if e.Tags == nil {
        e.Tags = make(map[string]string)
    }
    e.Tags[key] = value
    return e
}

// WithMetric adds a metric to the error
func (e *AppError) WithMetric(key string, value float64) *AppError {
    if e.Metrics == nil {
        e.Metrics = make(map[string]float64)
    }
    e.Metrics[key] = value
    return e
}

// ToMap returns error as a map for structured logging
func (e *AppError) ToMap() map[string]interface{} {
    result := map[string]interface{}{
        "error_code":      string(e.Code),
        "error_message":   e.Message,
        "error_category":  string(e.Category),
        "service":         e.Service,
        "operation":       e.Operation,
        "severity":        string(e.Severity),
        "retryable":       e.Retryable,
        "timestamp":       e.Timestamp,
    }
    
    if e.TraceID != "" {
        result["trace_id"] = e.TraceID
    }
    if e.SpanID != "" {
        result["span_id"] = e.SpanID
    }
    if e.RequestID != "" {
        result["request_id"] = e.RequestID
    }
    if e.UserID != "" {
        result["user_id"] = e.UserID
    }
    if len(e.Details) > 0 {
        result["details"] = e.Details
    }
    if len(e.Tags) > 0 {
        result["tags"] = e.Tags
    }
    if len(e.Metrics) > 0 {
        result["metrics"] = e.Metrics
    }
    
    return result
}

func captureStackTrace() []Frame {
    var frames []Frame
    for i := 2; i < 15; i++ { // Capture up to 15 frames
        pc, file, line, ok := runtime.Caller(i)
        if !ok {
            break
        }
        
        fn := runtime.FuncForPC(pc)
        if fn == nil {
            continue
        }
        
        frames = append(frames, Frame{
            Function: fn.Name(),
            File:     file,
            Line:     line,
        })
    }
    return frames
}

func extractContextInfo(ctx context.Context) (string, string, string, string, string) {
    var traceID, spanID, requestID, userID, sessionID string
    
    // Extract tracing information
    span := trace.SpanFromContext(ctx)
    if span.SpanContext().IsValid() {
        traceID = span.SpanContext().TraceID().String()
        spanID = span.SpanContext().SpanID().String()
    }
    
    // Extract context values
    if rid := ctx.Value("request_id"); rid != nil {
        if r, ok := rid.(string); ok {
            requestID = r
        }
    }
    
    if uid := ctx.Value("user_id"); uid != nil {
        if u, ok := uid.(string); ok {
            userID = u
        }
    }
    
    if sid := ctx.Value("session_id"); sid != nil {
        if s, ok := sid.(string); ok {
            sessionID = s
        }
    }
    
    return traceID, spanID, requestID, userID, sessionID
}

func getErrorCategory(code ErrorCode) ErrorCategory {
    switch code {
    case ErrCodeDatabase, ErrCodeQueue, ErrCodeVault, ErrCodeHTTP, ErrCodeGRPC, ErrCodeRedis:
        return CategoryInfrastructure
    case ErrCodeValidation, ErrCodeNotFound, ErrCodeConflict:
        return CategoryBusiness
    case ErrCodeUnauthorized, ErrCodeForbidden:
        return CategorySecurity
    case ErrCodeTimeout, ErrCodeRateLimit, ErrCodeCircuitBreaker:
        return CategoryPerformance
    default:
        return CategorySystem
    }
}