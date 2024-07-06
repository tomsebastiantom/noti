package domain

type Tenant struct {
    ID             string
    Name           string
    DefaultChannel string
    Preferences    map[string]ChannelPreference
}

type ChannelPreference struct {
    ChannelName NotificationChannel
    Enabled     bool
    ProviderID  string
}

type NotificationChannel string

const (
    Email   NotificationChannel = "email"
    SMS     NotificationChannel = "sms"
    Push    NotificationChannel = "push"
	Call    NotificationChannel = "call"
    WebPush  NotificationChannel = "webpush"
)