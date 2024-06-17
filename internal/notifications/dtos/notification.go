
package dto

type TemplateVariableDTO struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

type NotificationDTO struct {
    ID          string                `json:"id,omitempty"`
    TenantID    string                `json:"tenant_id"`
    UserID      string                `json:"user_id"`
    Type        string                `json:"type"`
    Channel     string                `json:"channel"`
    TemplateID  string                `json:"template_id"`
    Status      string                `json:"status"`
    Content     string                `json:"content"`
    Variables   []TemplateVariableDTO `json:"variables"`
    CreatedAt   string                `json:"created_at,omitempty"`
    UpdatedAt   string                `json:"updated_at,omitempty"`
    Error       string                `json:"error,omitempty"`
}
