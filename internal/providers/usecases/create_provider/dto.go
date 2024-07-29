package createprovider


type CreateProviderRequest struct {
    Name     string   `json:"name"`
    Channels []string `json:"channels"`
}

type CreateProviderResponse struct {
    ID       string              `json:"id"`
    Name     string              `json:"name"`
    Channels []ProviderChannelDTO `json:"channels"`
    Enabled  bool                `json:"enabled"`
}

type ProviderChannelDTO struct {
    Channel  string `json:"channel"`
    Priority int    `json:"priority"`
}