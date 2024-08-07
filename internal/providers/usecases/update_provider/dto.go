package updateprovider

type UpdateProviderRequest struct {
    ID       string              `json:"id"`
    Name     string              `json:"name"`
    Channels []ProviderChannelDTO `json:"channels"`
    Enabled  bool                `json:"enabled"`
}

type UpdateProviderResponse struct {
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