package errors

import (
    "context"
    "fmt"
    "time"
)

// ErrorBuilder provides a fluent interface for building errors
type ErrorBuilder struct {
    err *AppError
}

// New creates a new error builder
func New(code ErrorCode) *ErrorBuilder {
    return &ErrorBuilder{
        err: &AppError{
            Code:       code,
            Category:   getErrorCategory(code),
            Timestamp:  time.Now().UTC(),
            Service:    "noti", // Can be configured globally
            Severity:   GetDefaultSeverity(code),
            Retryable:  GetDefaultRetryable(code),
            StackTrace: captureStackTrace(),
        },
    }
}

// WithMessage sets the error message
func (b *ErrorBuilder) WithMessage(message string) *ErrorBuilder {
    b.err.Message = message
    return b
}

// WithMessagef sets the error message with formatting
func (b *ErrorBuilder) WithMessagef(format string, args ...any) *ErrorBuilder {
    b.err.Message = fmt.Sprintf(format, args...)
    return b
}

// WithCause wraps an underlying error
func (b *ErrorBuilder) WithCause(cause error) *ErrorBuilder {
    b.err.Cause = cause
    return b
}

// WithOperation sets the operation context
func (b *ErrorBuilder) WithOperation(operation string) *ErrorBuilder {
    b.err.Operation = operation
    return b
}

// WithDetail adds a detail field
func (b *ErrorBuilder) WithDetail(key string, value any) *ErrorBuilder {
    if b.err.Details == nil {
        b.err.Details = make(map[string]any)
    }
    b.err.Details[key] = value
    return b
}

// WithDetails adds multiple detail fields
func (b *ErrorBuilder) WithDetails(details map[string]any) *ErrorBuilder {
    if b.err.Details == nil {
        b.err.Details = make(map[string]any)
    }
    for k, v := range details {
        b.err.Details[k] = v
    }
    return b
}

// WithSeverity sets the error severity
func (b *ErrorBuilder) WithSeverity(severity Severity) *ErrorBuilder {
    b.err.Severity = severity
    return b
}

// WithRetryable marks the error as retryable
func (b *ErrorBuilder) WithRetryable(retryable bool) *ErrorBuilder {
    b.err.Retryable = retryable
    return b
}

// WithContext extracts information from context
func (b *ErrorBuilder) WithContext(ctx context.Context) *ErrorBuilder {
    traceID, spanID, requestID, userID, sessionID := extractContextInfo(ctx)
    
    b.err.TraceID = traceID
    b.err.SpanID = spanID
    b.err.RequestID = requestID
    b.err.UserID = userID
    b.err.SessionID = sessionID
    
    return b
}

// WithPublicMessage sets a user-safe message
func (b *ErrorBuilder) WithPublicMessage(message string) *ErrorBuilder {
    b.err.PublicMessage = message
    return b
}

// WithTag adds a tag
func (b *ErrorBuilder) WithTag(key, value string) *ErrorBuilder {
    if b.err.Tags == nil {
        b.err.Tags = make(map[string]string)
    }
    b.err.Tags[key] = value
    return b
}

// WithMetric adds a metric
func (b *ErrorBuilder) WithMetric(key string, value float64) *ErrorBuilder {
    if b.err.Metrics == nil {
        b.err.Metrics = make(map[string]float64)
    }
    b.err.Metrics[key] = value
    return b
}

// WithService sets the service name
func (b *ErrorBuilder) WithService(service string) *ErrorBuilder {
    b.err.Service = service
    return b
}

// Build creates the final error
func (b *ErrorBuilder) Build() *AppError {
    return b.err
}

// BuildAndLog creates the error and logs it (requires logger to be passed)
func (b *ErrorBuilder) BuildAndLog(logger interface{}) *AppError {
    err := b.Build()
    
    // Type assertion to avoid circular dependency
    if log, ok := logger.(interface {
        LogError(error, ...interface{})
    }); ok {
        log.LogError(err)
    }
    
    return err
}