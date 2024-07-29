package domain

// ChannelType represents the type of communication channel
type ChannelType string

const (
    ChannelTypeCall  ChannelType = "call"
    ChannelTypeSMS   ChannelType = "sms"
    ChannelTypeEmail ChannelType = "email"
)

// Channel represents a communication channel
type Channel struct {
    ID          string
    Type        ChannelType
    Name        string
    Description string
}

// Provider represents a service provider
type Provider struct {
    ID       string
    Name     string
    Channels []string // Store Channel IDs
    Enabled  bool
}

// ChannelPreference represents preferences for a specific channel and provider
type ChannelPreference struct {
    ProviderID string
    ChannelID  string
    Settings   map[string]interface{} // Flexible preference storage
}
