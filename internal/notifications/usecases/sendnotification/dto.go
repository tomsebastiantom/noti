package sendnotification

type SendNotificationRequest struct {
    TenantID   string
    UserID     string
    Type       string
    Channel    string
    TemplateID string
    Content    string
    Variables  []TemplateVariable
}


type SendNotificationResponse struct {
    ID         string
    TenantID   string
    UserID     string
    Type       string
    Channel    string
    TemplateID string
    Status     string
    Content    string
    Variables  []TemplateVariable
    Error      string
}

type TemplateVariable struct {
    Key   string
    Value string
}