package domain

type Tenant struct {
    ID             string
    Name           string
    Preferences    map[string]ChannelPreference
}

type ChannelPreference struct {
    ChannelName string
    Enabled     bool
    ProviderID  string
}