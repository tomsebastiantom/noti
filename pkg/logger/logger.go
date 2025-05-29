package logger

import (
    "context"
    "time"
    
    "github.com/rs/zerolog"
    "go.opentelemetry.io/otel/trace"
)

// Logger interface for structured logging with tracing
type Logger interface {
    // Basic logging methods
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)
    
    // Context-aware logging (with automatic tracing)
    DebugContext(ctx context.Context, msg string, fields ...Field)
    InfoContext(ctx context.Context, msg string, fields ...Field)
    WarnContext(ctx context.Context, msg string, fields ...Field)
    ErrorContext(ctx context.Context, msg string, fields ...Field)
    
    // Span-aware logging
    StartSpan(ctx context.Context, operationName string, fields ...Field) (context.Context, trace.Span)
    EndSpan(span trace.Span, err error, fields ...Field)
    
    // Error-specific logging
    LogError(err error, fields ...Field)
    LogErrorContext(ctx context.Context, err error, fields ...Field)
    
    // Child logger with persistent fields
    With(fields ...Field) Logger
    WithContext(ctx context.Context) Logger
    
    // Performance logging
    WithLatency(start time.Time) Logger
    LogDuration(operationName string, start time.Time, fields ...Field)
    
    // Raw access for advanced use
    Raw() *zerolog.Logger
}

// LoggerConfig holds configuration for logger
type LoggerConfig struct {
    Level       string
    Service     string
    Version     string
    Environment string
    EnableColor bool
    EnableCaller bool
    EnableTrace bool
}