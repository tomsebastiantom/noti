package domain


type ProviderChannel struct {
    Channel  string
    Priority int  // order 1 mean highest priority
}

type Provider struct {
    ID       string
    Name     string
    Enabled  bool
    Channels map[string]ProviderChannel
}

