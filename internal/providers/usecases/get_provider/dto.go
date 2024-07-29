package getprovider

type GetProviderRequest struct {
    ID string `json:"id"`
}

type GetProviderResponse struct {
    ID       string              `json:"id"`
    Name     string              `json:"name"`
    Channels []ProviderChannelDTO `json:"channels"`
    Enabled  bool                `json:"enabled"`
}

type ProviderChannelDTO struct {
    Channel  string `json:"channel"`
    Priority int    `json:"priority"`
}
