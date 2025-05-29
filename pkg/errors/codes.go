package errors

const (
    // Infrastructure errors
    ErrCodeDatabase     ErrorCode = "DATABASE_ERROR"
    ErrCodeQueue        ErrorCode = "QUEUE_ERROR"
    ErrCodeVault        ErrorCode = "VAULT_ERROR"
    ErrCodeHTTP         ErrorCode = "HTTP_ERROR"
    ErrCodeGRPC         ErrorCode = "GRPC_ERROR"
    ErrCodeRedis        ErrorCode = "REDIS_ERROR"
    ErrCodeElasticsearch ErrorCode = "ELASTICSEARCH_ERROR"
    ErrCodeS3           ErrorCode = "S3_ERROR"
    
    // Business logic errors
    ErrCodeValidation   ErrorCode = "VALIDATION_ERROR"
    ErrCodeNotFound     ErrorCode = "NOT_FOUND"
    ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
    ErrCodeForbidden    ErrorCode = "FORBIDDEN"
    ErrCodeConflict     ErrorCode = "CONFLICT"
    ErrCodeRateLimit    ErrorCode = "RATE_LIMIT"
    ErrCodeBusinessRule ErrorCode = "BUSINESS_RULE_VIOLATION"
    
    // System errors
    ErrCodeInternal       ErrorCode = "INTERNAL_ERROR"
    ErrCodeTimeout        ErrorCode = "TIMEOUT_ERROR"
    ErrCodeUnavailable    ErrorCode = "SERVICE_UNAVAILABLE"
    ErrCodeCircuitBreaker ErrorCode = "CIRCUIT_BREAKER"
    ErrCodeDeadlock       ErrorCode = "DEADLOCK_ERROR"
    ErrCodeOutOfMemory    ErrorCode = "OUT_OF_MEMORY"
    
    // Notification specific errors
    ErrCodeNotificationInvalid    ErrorCode = "NOTIFICATION_INVALID"
    ErrCodeNotificationFailed     ErrorCode = "NOTIFICATION_FAILED"
    ErrCodeTemplateNotFound       ErrorCode = "TEMPLATE_NOT_FOUND"
    ErrCodeRecipientInvalid       ErrorCode = "RECIPIENT_INVALID"
    ErrCodeChannelUnavailable     ErrorCode = "CHANNEL_UNAVAILABLE"
    ErrCodeMessageTooLarge        ErrorCode = "MESSAGE_TOO_LARGE"
    ErrCodeQuotaExceeded          ErrorCode = "QUOTA_EXCEEDED"
)

// GetDefaultSeverity returns the default severity for an error code
func GetDefaultSeverity(code ErrorCode) Severity {
    switch code {
    case ErrCodeDatabase, ErrCodeVault, ErrCodeOutOfMemory:
        return SeverityCritical
    case ErrCodeQueue, ErrCodeHTTP, ErrCodeTimeout, ErrCodeUnavailable, ErrCodeCircuitBreaker:
        return SeverityHigh
    case ErrCodeValidation, ErrCodeRateLimit, ErrCodeConflict, ErrCodeBusinessRule:
        return SeverityMedium
    case ErrCodeNotFound, ErrCodeUnauthorized, ErrCodeForbidden:
        return SeverityLow
    default:
        return SeverityMedium
    }
}

// GetDefaultRetryable returns whether an error code is retryable by default
func GetDefaultRetryable(code ErrorCode) bool {
    switch code {
    case ErrCodeDatabase, ErrCodeQueue, ErrCodeHTTP, ErrCodeTimeout, ErrCodeUnavailable, ErrCodeCircuitBreaker:
        return true
    case ErrCodeValidation, ErrCodeNotFound, ErrCodeUnauthorized, ErrCodeForbidden, ErrCodeConflict:
        return false
    default:
        return false
    }
}