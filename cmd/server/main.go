package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lystage-proj/internals/config"
	"lystage-proj/internals/db"
	"lystage-proj/internals/observability"
	"lystage-proj/internals/queue"
	api "lystage-proj/internals/routes"
	"lystage-proj/internals/worker"

	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	if err := observability.InitializeZap(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer observability.Logger.Sync()

	// Load config
	cfg := config.Load()

	// Init DB
	db.InitPostgres(cfg.DatabaseURL)

	// Init Kafka producer (using the new global producer)
	if err := queue.InitGlobalProducer(cfg.KafkaBroker, cfg.ClicksTopic); err != nil {
		observability.Logger.Fatal("Failed to initialize Kafka producer", zap.Error(err))
	}

	// Set up graceful shutdown for Kafka producer
	defer func() {
		if err := queue.CloseGlobalProducer(); err != nil {
			observability.Logger.Error("Error closing Kafka producer", zap.Error(err))
		}
	}()

	// Start Kafka consumer worker
	worker.StartClickConsumer(cfg.KafkaBroker, cfg.ClicksTopic, "click-consumers")

	// Setup router with all middleware and handlers
	router := api.SetupRouter(cfg)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Set up graceful shutdown handling
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	go func() {
		observability.Logger.Info("Starting HTTP server", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			observability.Logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-c
	observability.Logger.Info("Shutting down gracefully...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		observability.Logger.Error("Server forced to shutdown", zap.Error(err))
	} else {
		observability.Logger.Info("Server exited gracefully")
	}
}
