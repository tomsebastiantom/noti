package dto

type CreateTemplateRequest struct {
    TenantID  string
    Name      string
    Content   string
    IsPublic  bool
    Variables []string
}

type CreateTemplateResponse struct {
    Success bool
    Message string
}

type GetTemplateRequest struct {
    TemplateID string
    TenantID   string
}

type GetTemplateResponse struct {
    Template string
    Success  bool
    Message  string
}
