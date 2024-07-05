package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config structure
type Config struct {
	App      App      `yaml:"app"`
	HTTP     HTTP     `yaml:"http"`
	Log      Log      `yaml:"logger"`
	Database Database `yaml:"database"`
	RMQ      RMQ      `yaml:"rabbitmq"`
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

// Database structure
type Database struct {
	Type     string `yaml:"type"`
	Postgres PG     `yaml:"postgres"`
	MySQL    MySQL  `yaml:"mysql"`
}

// PG structure
type PG struct {
	PoolMax int    `yaml:"pool_max"`
	DSN     string `yaml:"dsn"`
}

// MySQL structure
type MySQL struct {
	DSN string `yaml:"dsn"`
}

// RMQ structure
type RMQ struct {
	ServerExchange string `yaml:"rpc_server_exchange"`
	ClientExchange string `yaml:"rpc_client_exchange"`
	URL            string `yaml:"url"`
}

// Global koanf instance. Use "." as the key path delimiter.
var k = koanf.New(".")

func LoadConfig() (*Config, error) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current working directory: %v", err)
	}
	fmt.Printf("Current working directory: %s\n", cwd)

	// Construct the path to config.yaml relative to the source file's directory
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("error getting source file directory")
	}
	sourceDir := filepath.Dir(sourceFile)
	configPath := filepath.Join(sourceDir, "config.yaml")

	// Check if the file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config.yaml not found in the source directory")
	}

	// Load YAML config
	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}

	// Debug print statement to verify the loaded configuration in koanf
	fmt.Printf("Koanf raw data: %+v\n", k.Raw())

	// Debug print statement to verify the flattened configuration map
	// confMapFlat := k.All()
	// fmt.Printf("Flattened configuration map: %+v\n", confMapFlat)

	var cfg Config
	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "yaml"}); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %v", err)
	}

	// Debug print statements to verify the loaded configuration
	// fmt.Printf("Loaded configuration: %+v\n", cfg)
	// fmt.Printf("Postgres PoolMax: %d\n", cfg.PG.PoolMax)
	// fmt.Printf("Postgres URL: %s\n", cfg.PG.URL)

	return &cfg, nil
}
