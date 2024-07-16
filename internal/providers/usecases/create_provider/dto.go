package createprovider


type CreateProviderRequest struct {
    Name     string   `json:"name"`
    Channels []string `json:"channels"`
    TenantID string   `json:"tenant_id"`
}

type CreateProviderResponse struct {
    ID       string   `json:"id"`
    Name     string   `json:"name"`
    Channels []string `json:"channels"`
    TenantID string   `json:"tenant_id"`
}