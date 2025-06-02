package container

import (
	"fmt"

	"getnoti.com/pkg/cache"
	"getnoti.com/pkg/credentials"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/queue"
	"getnoti.com/pkg/workerpool"
)


func (c *ServiceContainer) initializeInfrastructure() error {
    c.logger.Info("Initializing infrastructure services")
    
    // STEP 1: Initialize main database first - required for everything
    mainDB, err := db.NewDatabaseFactory(
        map[string]interface{}{
            "type": c.config.Database.Type,
            "dsn":  c.config.Database.DSN,
        },
        c.logger,
    )
    if err != nil {
        return fmt.Errorf("failed to initialize main database: %w", err)
    }
    c.mainDB = mainDB
    c.logger.Info("Main database initialized successfully")

    // STEP 2: Initialize cache
    c.cache = cache.NewGenericCache(1e7, 1<<30, 64)
    c.logger.Info("Cache initialized successfully")
    
    // STEP 3: Initialize credential manager BEFORE the DB manager
    // This breaks the circular dependency
    credentialManager, err := credentials.NewManager(
        c.config,
        c.logger,
        c.mainDB, // Pass the main DB directly
    )
    if err != nil {
        return fmt.Errorf("failed to initialize credential manager: %w", err)
    }
    c.credentialManager = credentialManager
    c.logger.Info("Credential manager initialized successfully")
    
    // STEP 4: Initialize config resolver with credential manager
    c.configResolver = db.NewConfigResolver(
        c.config,
        c.logger,
        credentialManager, // Already initialized
    )
    c.logger.Info("Database config resolver initialized successfully")
    
    // STEP 5: Initialize DB manager with config resolver
    c.dbManager = db.NewManager(
        c.cache,
        c.config,
        c.logger,
        mainDB,
    )
    c.logger.Info("Database manager initialized successfully")

	// Initialize queue manager
	c.queueManager = queue.NewQueueManager(
		queue.Config(c.config.Queue),
		c.logger,
	)
	c.logger.Info("Queue manager initialized successfully")

	// Initialize worker pool manager
	c.workerPoolManager = workerpool.NewWorkerPoolManager(c.logger)
	c.logger.Info("Worker pool manager initialized successfully")


	c.logger.Info("Infrastructure initialization completed successfully")
	return nil
}
