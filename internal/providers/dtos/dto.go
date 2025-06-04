package dtos

type SendNotificationRequest struct {
	Sender     string
	Receiver   string
	Channel    string
	Content    string
	ProviderID string
	TenantID   string
	UserID     string // User ID for preference checking
	Category   string // Notification category for preference filtering
}

type SendNotificationResponse struct {
	Success bool
	Message string
}
