package domain

type ChannelType int

const (
	Email ChannelType = iota
	SMS
	Push
	// Add more channel types as needed
)

// Provider represents a service provider with prioritized channels.
// The Channels map uses ChannelType as keys and priority as values.
// Lower priority numbers indicate higher priority (e.g., 1 is highest priority).


type PrioritizedChannel struct {
    Type     ChannelType
    Priority int
    Enabled  bool
}

type Provider struct {
    ID          string
    Name        string
    Channels    []PrioritizedChannel
    Credentials interface{}
}