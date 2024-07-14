package domain

type Tenant struct {
    ID             string
    Name           string
    Preferences    map[string]ChannelPreference
	DBConfig       *DBConfig
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
    Host     string
    Port     int
    Username string
    Password string
    DBName   string
}