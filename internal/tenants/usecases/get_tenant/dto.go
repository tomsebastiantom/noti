package gettenant



type GetTenantRequest struct {
    TenantID string
}

type GetTenantResponse struct {
    ID             string
    Name           string
}
