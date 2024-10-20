package getproviders

import (
    "getnoti.com/internal/providers/domain"
)
type GetProvidersRequest struct {
   
}

type GetProvidersResponse struct {
    Providers []ProviderResponse `json:"providers"`
}

type ProviderResponse struct {
    ID          string              `json:"id"`
    Name        string              `json:"name"`
    Channels    []ProviderChannelDTO `json:"channels"`
    Credentials interface{}         `json:"credentials"`
}

type ProviderChannelDTO struct {
    Type     domain.ChannelType `json:"type"`
    Priority int                `json:"priority"`
    Enabled  bool               `json:"enabled"`
}
