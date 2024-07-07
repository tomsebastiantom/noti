package createtenant

import "getnoti.com/internal/tenants/domain"

type CreateTenantRequest struct {
    ID             string
    Name           string
    Preferences    map[string]CreateChannelPreference
}

type CreateChannelPreference struct {
    ChannelName domain.NotificationChannel
    Enabled     bool
    ProviderID  string
}

type CreateTenantResponse struct {
    Tenant domain.Tenant
}
