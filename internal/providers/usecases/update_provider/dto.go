package updateprovider



type UpdateProviderRequest struct {
    ID       string   `json:"id"`
    Name     string   `json:"name"`
    Channels []string `json:"channels"`
}

type UpdateProviderResponse struct {
    ID       string   `json:"id"`
    Name     string   `json:"name"`
    Channels []string `json:"channels"`
    TenantID string   `json:"tenant_id"`
}