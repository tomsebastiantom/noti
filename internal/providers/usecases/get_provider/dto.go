package getprovider

import (
    "getnoti.com/internal/providers/domain"
)
type GetProviderRequest struct {
    ID string `json:"id"`
}

type GetProviderResponse struct {
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

