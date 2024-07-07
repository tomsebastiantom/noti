package sendnotification

type SendNotificationRequest struct {
    TenantID    string
    UserID      string
    Type        string
    Channel     string
    TemplateID  string
    Content     string
    ProviderID  string
    Variables   []TemplateVariable
}

type SendNotificationResponse struct {
    ID      string
    Status  string
    Error   string 
}

type TemplateVariable struct {
    Key   string
    Value string
}
