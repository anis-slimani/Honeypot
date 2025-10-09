package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"honey/internal/config"
	"honey/internal/database"
	"honey/internal/honeypot"
	"honey/internal/logger"
	"honey/internal/web"
)

func main() {
	// Parse command line arguments
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := logger.New(cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	logger.Info("Starting Honey SSH Honeypot...")

	// Initialize database
	db, err := database.Initialize(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	logger.Info("Database initialized successfully")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize honeypot
	honeypotServer := honeypot.New(cfg, logger, db)

	// Start honeypot server
	go func() {
		if err := honeypotServer.Start(ctx); err != nil {
			logger.Errorf("Honeypot server error: %v", err)
		}
	}()

	// Start web interface if enabled
	if cfg.Web.Enabled {
		webServer := web.New(cfg.Web, logger, db)
		go func() {
			if err := webServer.Start(ctx); err != nil {
				logger.Errorf("Web server error: %v", err)
			}
		}()
		logger.Infof("Web interface started on http://%s:%d", cfg.Web.Host, cfg.Web.Port)
	}

	logger.Infof("Honeypot SSH server started on %s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("Press Ctrl+C to stop the server")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down honeypot...")
	cancel()

	// Give services time to shutdown gracefully
	time.Sleep(2 * time.Second)
	logger.Info("Honeypot stopped")
}
