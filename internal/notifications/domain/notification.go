
package domain

type TemplateVariable struct {
    Key   string
    Value string
}

type Notification struct {
    ID          string
    TenantID    string
    UserID      string
    Type        string
    Channel     string
    TemplateID  string
    Status      string
    Content     string
    Variables   []TemplateVariable  
}
