package errors

import (
    "context"
    "fmt"
)

// Database errors
func DatabaseError(ctx context.Context, operation string, cause error) *AppError {
    return New(ErrCodeDatabase).
        WithContext(ctx).
        WithOperation(operation).
        WithCause(cause).
        WithMessage("Database operation failed").
        WithPublicMessage("Temporary service issue. Please try again.").
        Build()
}

func DatabaseConnectionError(ctx context.Context, cause error) *AppError {
    return New(ErrCodeDatabase).
        WithContext(ctx).
        WithOperation("database_connection").
        WithCause(cause).
        WithMessage("Failed to connect to database").
        WithSeverity(SeverityCritical).
        WithPublicMessage("Service temporarily unavailable").
        Build()
}

// Vault errors
func VaultError(ctx context.Context, operation string, cause error) *AppError {
    return New(ErrCodeVault).
        WithContext(ctx).
        WithOperation(operation).
        WithCause(cause).
        WithMessage("Vault operation failed").
        WithSeverity(SeverityCritical).
        WithRetryable(false).
        WithPublicMessage("Authentication service unavailable").
        Build()
}

func VaultUnauthorizedError(ctx context.Context) *AppError {
    return New(ErrCodeVault).
        WithContext(ctx).
        WithOperation("vault_auth").
        WithMessage("Vault authentication failed").
        WithSeverity(SeverityHigh).
        WithRetryable(false).
        WithPublicMessage("Authentication failed").
        Build()
}

// Validation errors
func ValidationError(ctx context.Context, field string, reason string) *AppError {
    return New(ErrCodeValidation).
        WithContext(ctx).
        WithOperation("validation").
        WithMessagef("Validation failed for field '%s': %s", field, reason).
        WithDetail("field", field).
        WithDetail("reason", reason).
        WithPublicMessage(fmt.Sprintf("Invalid %s: %s", field, reason)).
        Build()
}

func ValidationErrors(ctx context.Context, fieldErrors map[string]string) *AppError {
    details := make(map[string]any)
    for field, reason := range fieldErrors {
        details[field] = reason
    }
    
    return New(ErrCodeValidation).
        WithContext(ctx).
        WithOperation("validation").
        WithMessage("Multiple validation errors").
        WithDetails(details).
        WithPublicMessage("Invalid input data").
        Build()
}

// Not found errors
func NotFoundError(ctx context.Context, resource string, id string) *AppError {
    return New(ErrCodeNotFound).
        WithContext(ctx).
        WithMessagef("%s with ID '%s' not found", resource, id).
        WithDetail("resource", resource).
        WithDetail("id", id).
        WithPublicMessage(fmt.Sprintf("%s not found", resource)).
        Build()
}

// Queue errors
func QueueError(ctx context.Context, operation string, cause error) *AppError {
    return New(ErrCodeQueue).
        WithContext(ctx).
        WithOperation(operation).
        WithCause(cause).
        WithMessage("Queue operation failed").
        WithPublicMessage("Message delivery delayed. Will retry automatically.").
        Build()
}

func QueueConnectionError(ctx context.Context, cause error) *AppError {
    return New(ErrCodeQueue).
        WithContext(ctx).
        WithOperation("queue_connection").
        WithCause(cause).
        WithMessage("Failed to connect to queue").
        WithSeverity(SeverityCritical).
        WithPublicMessage("Service temporarily unavailable").
        Build()
}

// HTTP errors
func HTTPError(ctx context.Context, statusCode int, message string) *AppError {
    var code ErrorCode
    var severity Severity
    
    switch {
    case statusCode >= 500:
        code = ErrCodeInternal
        severity = SeverityHigh
    case statusCode == 404:
        code = ErrCodeNotFound
        severity = SeverityLow
    case statusCode == 401:
        code = ErrCodeUnauthorized
        severity = SeverityMedium
    case statusCode == 403:
        code = ErrCodeForbidden
        severity = SeverityMedium
    case statusCode == 429:
        code = ErrCodeRateLimit
        severity = SeverityMedium
    default:
        code = ErrCodeHTTP
        severity = SeverityMedium
    }
    
    return New(code).
        WithContext(ctx).
        WithMessage(message).
        WithDetail("status_code", statusCode).
        WithSeverity(severity).
        WithRetryable(statusCode >= 500 || statusCode == 429).
        WithPublicMessage(message).
        Build()
}

// Notification specific errors
func NotificationValidationError(ctx context.Context, field string, reason string) *AppError {
    return New(ErrCodeNotificationInvalid).
        WithContext(ctx).
        WithOperation("notification_validation").
        WithMessagef("Invalid notification: %s - %s", field, reason).
        WithDetail("field", field).
        WithDetail("reason", reason).
        WithPublicMessage(fmt.Sprintf("Invalid notification: %s", reason)).
        Build()
}

func NotificationSendError(ctx context.Context, channel string, cause error) *AppError {
    return New(ErrCodeNotificationFailed).
        WithContext(ctx).
        WithOperation("notification_send").
        WithCause(cause).
        WithMessagef("Failed to send notification via %s", channel).
        WithDetail("channel", channel).
        WithPublicMessage("Failed to send notification").
        Build()
}

func TemplateNotFoundError(ctx context.Context, templateID string) *AppError {
    return New(ErrCodeTemplateNotFound).
        WithContext(ctx).
        WithOperation("template_lookup").
        WithMessagef("Template '%s' not found", templateID).
        WithDetail("template_id", templateID).
        WithPublicMessage("Template not found").
        Build()
}

func RecipientInvalidError(ctx context.Context, recipient string, reason string) *AppError {
    return New(ErrCodeRecipientInvalid).
        WithContext(ctx).
        WithOperation("recipient_validation").
        WithMessagef("Invalid recipient '%s': %s", recipient, reason).
        WithDetail("recipient", recipient).
        WithDetail("reason", reason).
        WithPublicMessage("Invalid recipient").
        Build()
}

// System errors
func InternalError(ctx context.Context, operation string, cause error) *AppError {
    return New(ErrCodeInternal).
        WithContext(ctx).
        WithOperation(operation).
        WithCause(cause).
        WithMessage("Internal server error").
        WithSeverity(SeverityHigh).
        WithPublicMessage("An unexpected error occurred").
        Build()
}

func TimeoutError(ctx context.Context, operation string, timeout string) *AppError {
    return New(ErrCodeTimeout).
        WithContext(ctx).
        WithOperation(operation).
        WithMessagef("Operation timed out after %s", timeout).
        WithDetail("timeout", timeout).
        WithPublicMessage("Request timed out").
        Build()
}

func ServiceUnavailableError(ctx context.Context, service string) *AppError {
    return New(ErrCodeUnavailable).
        WithContext(ctx).
        WithOperation("service_call").
        WithMessagef("Service '%s' is unavailable", service).
        WithDetail("service", service).
        WithPublicMessage("Service temporarily unavailable").
        Build()
}