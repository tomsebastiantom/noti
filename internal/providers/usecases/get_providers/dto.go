package getproviders

type GetProvidersRequest struct {
    TenantID string `json:"tenant_id"`
}

type GetProvidersResponse struct {
    Providers []ProviderResponse `json:"providers"`
}

type ProviderResponse struct {
    ID       string              `json:"id"`
    Name     string              `json:"name"`
    Channels []ProviderChannelDTO `json:"channels"`
    TenantID string              `json:"tenant_id"`
    Enabled  bool                `json:"enabled"`
}

type ProviderChannelDTO struct {
    Channel  string `json:"channel"`
    Priority int    `json:"priority"`
}
