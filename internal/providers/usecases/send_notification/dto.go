package sendnotification

type SendNotificationRequest struct {
    Sender    string
    Receiver  string
    Channel string
    Content   string
    ProviderID string
}

type SendNotificationResponse struct {
    Success bool
    Message string
}
