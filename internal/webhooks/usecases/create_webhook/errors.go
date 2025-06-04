package create_webhook

import "fmt"

// CreateWebhookError represents errors that can occur during webhook creation
type CreateWebhookError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *CreateWebhookError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Common error constructors
func NewValidationError(message string) *CreateWebhookError {
	return &CreateWebhookError{
		Code:    "VALIDATION_ERROR",
		Message: message,
	}
}

func NewDuplicateNameError(name string) *CreateWebhookError {
	return &CreateWebhookError{
		Code:    "DUPLICATE_NAME",
		Message: fmt.Sprintf("webhook with name '%s' already exists", name),
	}
}

func NewInternalError(message string) *CreateWebhookError {
	return &CreateWebhookError{
		Code:    "INTERNAL_ERROR",
		Message: message,
	}
}

func NewUnauthorizedError() *CreateWebhookError {
	return &CreateWebhookError{
		Code:    "UNAUTHORIZED",
		Message: "unauthorized to create webhook",
	}
}
