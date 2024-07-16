package config

import (
	"fmt"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Config structure
type Config struct {
	App      App      `yaml:"app"`
	HTTP     HTTP     `yaml:"http"`
	Log      Log      `yaml:"logger"`
	Database Database `yaml:"database"`
	Queue    Queue    `yaml:"queue"`
	Vault    Vault    `yaml:"vault"`
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
	Type string `yaml:"type"`
	DSN  string `yaml:"dsn"`
}

// Queue structure
type Queue struct {
	URL               string        `yaml:"url"`
	ReconnectInterval time.Duration `yaml:"reconnect_interval"`
}

type Vault struct {
	Address string `yaml:"address"`
	Token   string `yaml:"token"`
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

	var cfg Config
	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "yaml"}); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %v", err)
	}

	return &cfg, nil
}
