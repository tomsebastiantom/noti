package domain

import (
    "errors"
    "strings"
)

type Tenant struct {
    ID       string
    Name     string
    DBConfig *DBCredentials
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

func NewTenant(id, name string) *Tenant {
    return &Tenant{
        ID:       id,
        Name:     name,
        DBConfig: nil,
    }
}

func (t *Tenant) SetDBConfig(config *DBCredentials) {
    t.DBConfig = config
}

func (t *Tenant) Validate() error {
    if t.ID == "" {
        return errors.New("tenant ID cannot be empty")
    }
    if t.Name == "" {
        return errors.New("tenant name cannot be empty")
    }
    if t.DBConfig == nil {
        return errors.New("DBConfig cannot be nil")
    }
    return t.DBConfig.Validate()
}

func NewDBCredentials(dbType string, dsn string, host string, port int, username, password, dbName string) (*DBCredentials, error) {
    creds := &DBCredentials{
        Type: dbType,
    }

    if dbType == "" {
        return nil, errors.New("database type is required")
    }

    if dsn != "" {
        // If DSN is provided, use it and ignore other fields
        creds.DSN = dsn
    } else {
        // If no DSN, validate and set individual fields
        if host == "" || port == 0 || username == "" || password == "" {
            return nil, errors.New("host, port, username, and password are required when DSN is not provided")
        }
        creds.Host = host
        creds.Port = port
        creds.Username = username
        creds.Password = password
        creds.DBName = dbName
    }

    return creds, nil
}

func (db *DBCredentials) Validate() error {
    if db.Type == "" {
        return errors.New("database type is required")
    }

    if db.DSN != "" {
        // If DSN is provided, no need to validate other fields
        return nil
    }

    var missingFields []string
    if db.Host == "" {
        missingFields = append(missingFields, "host")
    }
    if db.Port == 0 {
        missingFields = append(missingFields, "port")
    }
    if db.Username == "" {
        missingFields = append(missingFields, "username")
    }
    if db.Password == "" {
        missingFields = append(missingFields, "password")
    }

    if len(missingFields) > 0 {
        return errors.New("missing required fields: " + strings.Join(missingFields, ", "))
    }

    return nil
}
