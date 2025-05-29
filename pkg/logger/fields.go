package logger

import (
    "time"
)

// Field represents a key-value pair for structured logging
type Field struct {
    Key   string
    Value interface{}
}

// Basic field constructors
func String(key, value string) Field { 
    return Field{Key: key, Value: value} 
}

func Int(key string, value int) Field { 
    return Field{Key: key, Value: value} 
}

func Int64(key string, value int64) Field { 
    return Field{Key: key, Value: value} 
}

func Float64(key string, value float64) Field { 
    return Field{Key: key, Value: value} 
}

func Bool(key string, value bool) Field { 
    return Field{Key: key, Value: value} 
}

func Duration(key string, value time.Duration) Field { 
    return Field{Key: key, Value: value} 
}

func Time(key string, value time.Time) Field { 
    return Field{Key: key, Value: value} 
}

func Any(key string, value interface{}) Field { 
    return Field{Key: key, Value: value} 
}

// Common application field constructors
func Error(err error) Field { 
    return Field{Key: "error", Value: err} 
}

func TraceID(traceID string) Field { 
    return Field{Key: "trace_id", Value: traceID} 
}

func SpanID(spanID string) Field { 
    return Field{Key: "span_id", Value: spanID} 
}

func Operation(op string) Field { 
    return Field{Key: "operation", Value: op} 
}

func UserID(userID string) Field { 
    return Field{Key: "user_id", Value: userID} 
}

func RequestID(requestID string) Field { 
    return Field{Key: "request_id", Value: requestID} 
}

func SessionID(sessionID string) Field { 
    return Field{Key: "session_id", Value: sessionID} 
}

func Component(component string) Field { 
    return Field{Key: "component", Value: component} 
}

func Method(method string) Field { 
    return Field{Key: "method", Value: method} 
}

func URL(url string) Field { 
    return Field{Key: "url", Value: url} 
}

func StatusCode(code int) Field { 
    return Field{Key: "status_code", Value: code} 
}

func Latency(duration time.Duration) Field { 
    return Field{Key: "latency_ms", Value: duration.Milliseconds()} 
}

// Database-related fields
func Table(table string) Field { 
    return Field{Key: "table", Value: table} 
}

func Query(query string) Field { 
    return Field{Key: "query", Value: query} 
}

func RowsAffected(count int64) Field { 
    return Field{Key: "rows_affected", Value: count} 
}

// Notification-related fields
func NotificationID(id string) Field { 
    return Field{Key: "notification_id", Value: id} 
}

func Channel(channel string) Field { 
    return Field{Key: "channel", Value: channel} 
}

func Recipient(recipient string) Field { 
    return Field{Key: "recipient", Value: recipient} 
}

func Template(template string) Field { 
    return Field{Key: "template", Value: template} 
}

func MessageSize(size int) Field { 
    return Field{Key: "message_size", Value: size} 
}

// HTTP-related fields
func RemoteAddr(addr string) Field { 
    return Field{Key: "remote_addr", Value: addr} 
}

func UserAgent(ua string) Field { 
    return Field{Key: "user_agent", Value: ua} 
}

func Referer(ref string) Field { 
    return Field{Key: "referer", Value: ref} 
}

func Protocol(protocol string) Field { 
    return Field{Key: "protocol", Value: protocol} 
}

// Queue-related fields
func QueueName(name string) Field { 
    return Field{Key: "queue_name", Value: name} 
}

func MessageID(id string) Field { 
    return Field{Key: "message_id", Value: id} 
}

func RetryCount(count int) Field { 
    return Field{Key: "retry_count", Value: count} 
}

func Exchange(exchange string) Field { 
    return Field{Key: "exchange", Value: exchange} 
}

func RoutingKey(key string) Field { 
    return Field{Key: "routing_key", Value: key} 
}

// Performance fields
func MemoryUsage(bytes int64) Field { 
    return Field{Key: "memory_usage_bytes", Value: bytes} 
}

func CPUUsage(percent float64) Field { 
    return Field{Key: "cpu_usage_percent", Value: percent} 
}

func ConnectionCount(count int) Field { 
    return Field{Key: "connection_count", Value: count} 
}

func QueueLength(length int) Field { 
    return Field{Key: "queue_length", Value: length} 
}