package getproviderbytenant

type GetProviderByTenantRequest struct {
    TenantID string `json:"tenant_id"`
}

type GetProviderByTenantResponse struct {
    Providers []ProviderResponse `json:"providers"`
}

type ProviderResponse struct {
    ID       string   `json:"id"`
    Name     string   `json:"name"`
    Channels []string `json:"channels"`
    TenantID string   `json:"tenant_id"`
}