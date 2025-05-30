package logger

import (
    "context"
    
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "go.opentelemetry.io/otel/trace"
    "getnoti.com/config"
)

type Field struct {
    Key   string
    Value interface{}
}


type Logger interface {

    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)
    
    // Context-aware logging
    DebugContext(ctx context.Context, msg string, fields ...Field)
    InfoContext(ctx context.Context, msg string, fields ...Field)
    WarnContext(ctx context.Context, msg string, fields ...Field)
    ErrorContext(ctx context.Context, msg string, fields ...Field)
    

    LogError(err error, fields ...Field)
    LogErrorContext(ctx context.Context, err error, fields ...Field)
    
 
    With(fields ...Field) Logger
  
    // String(key, value string) Field
    // Int(key string, value int) Field
    // Int64(key string, value int64) Field
    // Float64(key string, value float64) Field
    // Bool(key string, value bool) Field
    // Duration(key string, value time.Duration) Field
    // Any(key string, value interface{}) Field
}

// zapLogger implements Logger using Zap
type zapLogger struct {
    logger *zap.Logger
}

// New creates a logger using Zap as implementation
func New(cfg *config.Config) Logger {
    var zapConfig zap.Config
    
    if cfg.Env == "production" {
        zapConfig = zap.NewProductionConfig()
        // Production optimizations
        zapConfig.EncoderConfig.TimeKey = "timestamp"
        zapConfig.EncoderConfig.CallerKey = "caller"
        zapConfig.EncoderConfig.StacktraceKey = "stacktrace"
    } else {
        zapConfig = zap.NewDevelopmentConfig()
        // Development settings for better readability
        zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    }
    
    // Set log level
    switch cfg.Logger.Level {
    case "debug":
        zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
    case "info":
        zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
    case "warn":
        zapConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
    case "error":
        zapConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
    default:
        zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
    }
    
    logger, err := zapConfig.Build()
    if err != nil {
        // Fallback to development logger if build fails
        logger = zap.NewExample()
    }
    
    baseLogger := logger.With(
        zap.String("service", cfg.App.Name),
        zap.String("version", cfg.App.Version),
        zap.String("environment", cfg.Env),
    )
    
    return &zapLogger{logger: baseLogger}
}

// Basic logging implementation
func (l *zapLogger) Debug(msg string, fields ...Field) {
    l.logger.Debug(msg, l.convertFields(fields...)...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
    l.logger.Info(msg, l.convertFields(fields...)...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
    l.logger.Warn(msg, l.convertFields(fields...)...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
    l.logger.Error(msg, l.convertFields(fields...)...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
    l.logger.Fatal(msg, l.convertFields(fields...)...)
}

// Context-aware logging implementation
func (l *zapLogger) DebugContext(ctx context.Context, msg string, fields ...Field) {
    zapFields := l.convertFields(fields...)
    zapFields = append(zapFields, l.contextFields(ctx)...)
    l.logger.Debug(msg, zapFields...)
}

func (l *zapLogger) InfoContext(ctx context.Context, msg string, fields ...Field) {
    zapFields := l.convertFields(fields...)
    zapFields = append(zapFields, l.contextFields(ctx)...)
    l.logger.Info(msg, zapFields...)
}

func (l *zapLogger) WarnContext(ctx context.Context, msg string, fields ...Field) {
    zapFields := l.convertFields(fields...)
    zapFields = append(zapFields, l.contextFields(ctx)...)
    l.logger.Warn(msg, zapFields...)
}

func (l *zapLogger) ErrorContext(ctx context.Context, msg string, fields ...Field) {
    zapFields := l.convertFields(fields...)
    zapFields = append(zapFields, l.contextFields(ctx)...)
    l.logger.Error(msg, zapFields...)
}


func (l *zapLogger) LogError(err error, fields ...Field) {
    zapFields := l.convertFields(fields...)
    zapFields = append(zapFields, zap.Error(err))
    l.logger.Error("Error occurred", zapFields...)
}

func (l *zapLogger) LogErrorContext(ctx context.Context, err error, fields ...Field) {
    zapFields := l.convertFields(fields...)
    zapFields = append(zapFields, l.contextFields(ctx)...)
    zapFields = append(zapFields, zap.Error(err))
    l.logger.Error("Error occurred", zapFields...)
}


func (l *zapLogger) With(fields ...Field) Logger {
    return &zapLogger{
        logger: l.logger.With(l.convertFields(fields...)...),
    }
}

// Field helpers - these create Field structs for your API
// func (l *zapLogger) String(key, value string) Field {
//     return Field{Key: key, Value: value}
// }

// func (l *zapLogger) Int(key string, value int) Field {
//     return Field{Key: key, Value: value}
// }

// func (l *zapLogger) Int64(key string, value int64) Field {
//     return Field{Key: key, Value: value}
// }

// func (l *zapLogger) Float64(key string, value float64) Field {
//     return Field{Key: key, Value: value}
// }

// func (l *zapLogger) Bool(key string, value bool) Field {
//     return Field{Key: key, Value: value}
// }

// func (l *zapLogger) Duration(key string, value time.Duration) Field {
//     return Field{Key: key, Value: value}
// }

// func (l *zapLogger) Any(key string, value interface{}) Field {
//     return Field{Key: key, Value: value}
// }
// func (l *zapLogger) Error(key string, value interface{}) Field {
//     return Field{Key: key, Value: value}
// }

// Helper methods
func (l *zapLogger) convertFields(fields ...Field) []zap.Field {
    zapFields := make([]zap.Field, len(fields))
    for i, field := range fields {
        zapFields[i] = zap.Any(field.Key, field.Value)
    }
    return zapFields
}

func (l *zapLogger) contextFields(ctx context.Context) []zap.Field {
    var fields []zap.Field
    
    // Extract request ID from context
    if requestID := ctx.Value("request_id"); requestID != nil {
        fields = append(fields, zap.String("request_id", requestID.(string)))
    }
    
    // Extract user ID from context
    if userID := ctx.Value("user_id"); userID != nil {
        fields = append(fields, zap.String("user_id", userID.(string)))
    }
    
    // Extract tracing information
    span := trace.SpanFromContext(ctx)
    if span.IsRecording() {
        spanCtx := span.SpanContext()
        fields = append(fields,
            zap.String("trace_id", spanCtx.TraceID().String()),
            zap.String("span_id", spanCtx.SpanID().String()),
        )
    }
    
    return fields
}