package getprovider

type GetProviderRequest struct {
    ID string `json:"id"`
}

type GetProviderResponse struct {
    ID       string   `json:"id"`
    Name     string   `json:"name"`
    Channels []string `json:"channels"`
    TenantID string   `json:"tenant_id"`
}