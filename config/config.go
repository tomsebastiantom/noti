
package config

import (
	"fmt"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config structure
type Config struct {
	App  App  `yaml:"app"`
	HTTP HTTP `yaml:"http"`
	Log  Log  `yaml:"logger"`
	PG   PG   `yaml:"postgres"`
	RMQ  RMQ  `yaml:"rabbitmq"`
}

// App structure
type App struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// HTTP structure
type HTTP struct {
	Port string `yaml:"port"`
}

// Log structure
type Log struct {
	Level string `yaml:"log_level"`
}

// PG structure
type PG struct {
	PoolMax int    `yaml:"pool_max"`
	URL     string `yaml:"url"`
}

// RMQ structure
type RMQ struct {
	ServerExchange string `yaml:"rpc_server_exchange"`
	ClientExchange string `yaml:"rpc_client_exchange"`
	URL            string `yaml:"url"`
}

// Global koanf instance. Use "." as the key path delimiter.
var k = koanf.New(".")

func LoadConfig(configPath string) (*Config, error) {
	// Load YAML config.
	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %v", err)
	}

	return &cfg, nil
}
