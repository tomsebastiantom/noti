package logger

import (
    "context"
    "io"
    "os"
    "strings"


    "getnoti.com/config"
    "github.com/rs/zerolog"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/codes" 
    "go.opentelemetry.io/otel/trace"
)

// ZerologLogger implements the Logger interface
type ZerologLogger struct {
    logger  *zerolog.Logger
    service string
    version string
    env     string
    tracer  trace.Tracer
}

var _ Logger = (*ZerologLogger)(nil)

// New creates a new logger instance
func New(cfg *config.Config) Logger {
    loggerCfg := LoggerConfig{
        Level:       cfg.Log.Level,
        Service:     cfg.App.Name,
        Version:     cfg.App.Version,
        Environment: getEnv(cfg),
        EnableColor: cfg.Env != "production",
        EnableCaller: true,
        EnableTrace: true,
    }
    
    return NewWithConfig(loggerCfg)
}

func NewWithConfig(cfg LoggerConfig) Logger {
    // Parse log level
    level := parseLogLevel(cfg.Level)
    zerolog.SetGlobalLevel(level)
    
    // Configure output
    var output io.Writer = os.Stdout
    if cfg.EnableColor && cfg.Environment != "production" {
        output = zerolog.ConsoleWriter{
            Out:        os.Stdout,
            TimeFormat: "15:04:05.000",
            NoColor:    false,
        }
    }
    
    // Create base logger
    logger := zerolog.New(output).
        Level(level).
        With().
        Timestamp().
        Str("service", cfg.Service).
        Str("version", cfg.Version).
        Str("environment", cfg.Environment)
    
    if cfg.EnableCaller {
        logger = logger.Caller()
    }
    
    baseLogger := logger.Logger()
    
    // Create tracer for spans
    var tracer trace.Tracer
    if cfg.EnableTrace {
        tracer = otel.Tracer(cfg.Service)
    }
    
    return &ZerologLogger{
        logger:  &baseLogger,
        service: cfg.Service,
        version: cfg.Version,
        env:     cfg.Environment,
        tracer:  tracer,
    }
}

// Basic logging methods
func (l *ZerologLogger) Debug(msg string, fields ...Field) {
    l.logWithFields(l.logger.Debug(), msg, fields...)
}

func (l *ZerologLogger) Info(msg string, fields ...Field) {
    l.logWithFields(l.logger.Info(), msg, fields...)
}

func (l *ZerologLogger) Warn(msg string, fields ...Field) {
    l.logWithFields(l.logger.Warn(), msg, fields...)
}

func (l *ZerologLogger) Error(msg string, fields ...Field) {
    l.logWithFields(l.logger.Error(), msg, fields...)
}

func (l *ZerologLogger) Fatal(msg string, fields ...Field) {
    l.logWithFields(l.logger.Fatal(), msg, fields...)
}

// Context-aware logging methods
func (l *ZerologLogger) DebugContext(ctx context.Context, msg string, fields ...Field) {
    l.logWithContext(ctx, l.logger.Debug(), msg, fields...)
}

func (l *ZerologLogger) InfoContext(ctx context.Context, msg string, fields ...Field) {
    l.logWithContext(ctx, l.logger.Info(), msg, fields...)
}

func (l *ZerologLogger) WarnContext(ctx context.Context, msg string, fields ...Field) {
    l.logWithContext(ctx, l.logger.Warn(), msg, fields...)
}

func (l *ZerologLogger) ErrorContext(ctx context.Context, msg string, fields ...Field) {
    l.logWithContext(ctx, l.logger.Error(), msg, fields...)
    
    // Add error to current span if exists
    if span := trace.SpanFromContext(ctx); span.IsRecording() {
        span.SetStatus(codes.Error, msg)  // Fixed: Use codes.Error instead of trace.StatusError
    }
}

func (l *ZerologLogger) Raw() *zerolog.Logger {
    return l.logger
}

// Helper functions
func parseLogLevel(level string) zerolog.Level {
    switch strings.ToLower(level) {
    case "trace":
        return zerolog.TraceLevel
    case "debug":
        return zerolog.DebugLevel
    case "info":
        return zerolog.InfoLevel
    case "warn", "warning":
        return zerolog.WarnLevel
    case "error":
        return zerolog.ErrorLevel
    case "fatal":
        return zerolog.FatalLevel
    case "panic":
        return zerolog.PanicLevel
    default:
        return zerolog.InfoLevel
    }
}

func getEnv(cfg *config.Config) string {
    if cfg.Env != "" {
        return cfg.Env
    }
    return "development"
}

func (l *ZerologLogger) logWithFields(event *zerolog.Event, msg string, fields ...Field) {
    for _, field := range fields {
        event = event.Interface(field.Key, field.Value)
    }
    event.Msg(msg)
}