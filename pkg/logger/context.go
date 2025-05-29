package logger

import (
    "context"
    
    "github.com/rs/zerolog"
    "go.opentelemetry.io/otel/trace"
)

func (l *ZerologLogger) logWithContext(ctx context.Context, event *zerolog.Event, msg string, fields ...Field) {
    // Add tracing information
    event = l.addTracingInfo(ctx, event)
    
    // Add request context
    event = l.addRequestContext(ctx, event)
    
    // Add provided fields
    for _, field := range fields {
        event = event.Interface(field.Key, field.Value)
    }
    
    event.Msg(msg)
}

func (l *ZerologLogger) addTracingInfo(ctx context.Context, event *zerolog.Event) *zerolog.Event {
    span := trace.SpanFromContext(ctx)
    if span.SpanContext().IsValid() {
        event = event.
            Str("trace_id", span.SpanContext().TraceID().String()).
            Str("span_id", span.SpanContext().SpanID().String())
    }
    return event
}

func (l *ZerologLogger) addRequestContext(ctx context.Context, event *zerolog.Event) *zerolog.Event {
    if requestID := getStringFromContext(ctx, "request_id"); requestID != "" {
        event = event.Str("request_id", requestID)
    }
    
    if userID := getStringFromContext(ctx, "user_id"); userID != "" {
        event = event.Str("user_id", userID)
    }
    
    if sessionID := getStringFromContext(ctx, "session_id"); sessionID != "" {
        event = event.Str("session_id", sessionID)
    }
    
    return event
}

// Child logger creation
func (l *ZerologLogger) With(fields ...Field) Logger {
    logger := l.logger.With()
    for _, field := range fields {
        logger = logger.Interface(field.Key, field.Value)
    }
    newLogger := logger.Logger()
    
    return &ZerologLogger{
        logger:  &newLogger,
        service: l.service,
        version: l.version,
        env:     l.env,
        tracer:  l.tracer,
    }
}

func (l *ZerologLogger) WithContext(ctx context.Context) Logger {
    var fields []Field
    
    // Extract tracing info
    span := trace.SpanFromContext(ctx)
    if span.SpanContext().IsValid() {
        fields = append(fields,
            TraceID(span.SpanContext().TraceID().String()),
            SpanID(span.SpanContext().SpanID().String()),
        )
    }
    
    // Extract request context
    if requestID := getStringFromContext(ctx, "request_id"); requestID != "" {
        fields = append(fields, RequestID(requestID))
    }
    
    if userID := getStringFromContext(ctx, "user_id"); userID != "" {
        fields = append(fields, UserID(userID))
    }
    
    return l.With(fields...)
}

// Error logging with AppError support
func (l *ZerologLogger) LogError(err error, fields ...Field) {
    // Check if it's an AppError from our errors package
    if appErr, ok := err.(interface {
        ToMap() map[string]interface{}
        GetSeverity() interface{}
    }); ok {
        l.logAppError(appErr, fields...)
    } else {
        l.Error("Error occurred", append(fields, Error(err))...)
    }
}

func (l *ZerologLogger) LogErrorContext(ctx context.Context, err error, fields ...Field) {
    if appErr, ok := err.(interface {
        ToMap() map[string]interface{}
        GetSeverity() interface{}
    }); ok {
        l.logAppErrorWithContext(ctx, appErr, fields...)
    } else {
        l.ErrorContext(ctx, "Error occurred", append(fields, Error(err))...)
    }
}

func (l *ZerologLogger) logAppError(appErr interface {
    ToMap() map[string]interface{}
    GetSeverity() interface{}
}, fields ...Field) {
    
    errorMap := appErr.ToMap()
    event := l.logger.Error()
    
    // Add error fields
    for key, value := range errorMap {
        event = event.Interface(key, value)
    }
    
    // Add additional fields
    for _, field := range fields {
        event = event.Interface(field.Key, field.Value)
    }
    
    if msg, ok := errorMap["error_message"].(string); ok {
        event.Msg(msg)
    } else {
        event.Msg("Application error occurred")
    }
}

func (l *ZerologLogger) logAppErrorWithContext(ctx context.Context, appErr interface {
    ToMap() map[string]interface{}
    GetSeverity() interface{}
}, fields ...Field) {
    
    event := l.addTracingInfo(ctx, l.logger.Error())
    event = l.addRequestContext(ctx, event)
    
    errorMap := appErr.ToMap()
    for key, value := range errorMap {
        event = event.Interface(key, value)
    }
    
    // Add provided fields
    for _, field := range fields {
        event = event.Interface(field.Key, field.Value)
    }
    
    if msg, ok := errorMap["error_message"].(string); ok {
        event.Msg(msg)
    } else {
        event.Msg("Application error occurred")
    }
}

// Utility functions
func getStringFromContext(ctx context.Context, key string) string {
    if value := ctx.Value(key); value != nil {
        if str, ok := value.(string); ok {
            return str
        }
    }
    return ""
}