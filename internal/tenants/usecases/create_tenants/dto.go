package createtenants

type CreateTenantsInput struct {
    Tenants []TenantInput
}

type TenantInput struct {
    ID             string
    Name           string
    DefaultChannel string
    Preferences    map[string]ChannelPreferenceInput
}

type ChannelPreferenceInput struct {
    ChannelName string
    Enabled     bool
    ProviderID  string
}

type CreateTenantsOutput struct {
    SuccessTenants []string
    FailedTenants  []FailedTenant
}

type FailedTenant struct {
    ID    string
    Error string
}
