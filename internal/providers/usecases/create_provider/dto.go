package createprovider

import (
    "getnoti.com/internal/providers/domain"
)

type CreateProviderRequest struct {
    Name        string                `json:"name"`
    Channels    []domain.ChannelType  `json:"channels"`
    Credentials interface{}           `json:"credentials"`
}

type CreateProviderResponse struct {
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