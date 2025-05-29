package logger

import (
    "context"
    "fmt"
    "time"
    
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)

// Span management
func (l *ZerologLogger) StartSpan(ctx context.Context, operationName string, fields ...Field) (context.Context, trace.Span) {
    if l.tracer == nil {
        return ctx, trace.SpanFromContext(ctx) // Return existing span if no tracer
    }
    
    // Create span attributes from fields
    var attrs []attribute.KeyValue
    for _, field := range fields {
        attrs = append(attrs, attribute.String(field.Key, fmt.Sprintf("%v", field.Value)))
    }
    
    // Start new span
    ctx, span := l.tracer.Start(ctx, operationName, trace.WithAttributes(attrs...))
    
    // Log span start
    l.InfoContext(ctx, "Span started",
        Operation(operationName),
        String("span_kind", "internal"),
    )
    
    return ctx, span
}

func (l *ZerologLogger) EndSpan(span trace.Span, err error, fields ...Field) {
    if !span.IsRecording() {
        return
    }
    
    // Set span status
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        span.RecordError(err)
    } else {
        span.SetStatus(codes.Ok, "")
    }
    
    // Add fields as span attributes
    for _, field := range fields {
        span.SetAttributes(attribute.String(field.Key, fmt.Sprintf("%v", field.Value)))
    }
    
    // Log span end
    ctx := trace.ContextWithSpan(context.Background(), span)
    if err != nil {
        l.ErrorContext(ctx, "Span ended with error", append(fields, Error(err))...)
    } else {
        l.InfoContext(ctx, "Span ended successfully", fields...)
    }
    
    span.End()
}

// Performance logging
func (l *ZerologLogger) WithLatency(start time.Time) Logger {
    duration := time.Since(start)
    return l.With(Latency(duration))
}

func (l *ZerologLogger) LogDuration(operationName string, start time.Time, fields ...Field) {
    duration := time.Since(start)
    allFields := append(fields, 
        Operation(operationName),
        Latency(duration),
    )
    l.Info("Operation completed", allFields...)
}