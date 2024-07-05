package dtos

type SendNotificationRequest struct {
    Sender    string
    Receiver  string
    Content   string
    ProviderID string
}

type SendNotificationResponse struct {
    Success bool
    Message string
}
