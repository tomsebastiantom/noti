package main

// import (
// 	"fmt"
// 	"os"
// 	"os/signal"
// 	"syscall"

// 	"getnoti.com/internal"
// 	"getnoti.com/pkg/httpserver"
// 	"getnoti.com/pkg/logger"
// )

// func main() {
// 	log, err := internal.InitializeLogger()
// 	if err != nil {
// 		fmt.Printf("Failed to initialize logger: %v\n", err)
// 		os.Exit(1)
// 	}

// 	httpServer := startServers(log)
// 	err = waitForSignals(log, httpServer)
// 	shutdown(err, httpServer, log)
// }

// func startServers(log *logger.Logger) *httpserver.Server {
// 	httpServer, err := internal.InitializeNewHttpServer()
// 	if err != nil {
// 		log.Error(fmt.Errorf("app - Run - httpServer initialization: %w", err))
// 		os.Exit(1)
// 	}
// 	return httpServer
// }

// func waitForSignals(log *logger.Logger, httpServer *httpserver.Server) error {
// 	// Waiting for signal
// 	interrupt := make(chan os.Signal, 1)
// 	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

// 	var err error
// 	select {
// 	case s := <-interrupt:
// 		log.Info("app - Run - signal: " + s.String())
// 	case err = <-httpServer.Notify():
// 		log.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
// 	}
// 	return err
// }

// func shutdown(err error, httpServer *httpserver.Server, log *logger.Logger) {
// 	if err != nil {
// 		log.Error(fmt.Errorf("app - Run - shutdown error: %w", err))
// 	}

// 	err = httpServer.Shutdown()
// 	if err != nil {
// 		log.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
// 	}
// }
