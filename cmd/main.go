package main

import (
	"fmt"
	"log"

	"getnoti.com/config"
	"getnoti.com/internal/app"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}
	defer application.Cleanup()

	if err := application.Run(); err != nil {
		return fmt.Errorf("application run error: %w", err)
	}

	return nil
}