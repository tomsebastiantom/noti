package domain

type Tenant struct {
    ID             string
    Name           string
    Preferences    map[string]ChannelPreference
    DBConfigs      map[string]*DBConfig
}

type ChannelPreference struct {
    ChannelName string
    Enabled     bool
    ProviderID  string
}

type DBConfig struct {
    CreateNewDB bool
    Credentials *DBCredentials
}

type DBCredentials struct {
    Type     string
    Host     string
    Port     int
    Username string
    Password string
    DBName   string
    DSN      string
}
