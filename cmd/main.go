package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	authula "github.com/0oMarko0/authula"
	"github.com/0oMarko0/authula/cmd/shared/configloader"
	authulaenv "github.com/0oMarko0/authula/env"
	"github.com/0oMarko0/authula/internal/bootstrap"
	authulamodels "github.com/0oMarko0/authula/models"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Run Authula with plugins built from config file
// This demonstrates the unified architecture where both library and standalone modes
// use identical runtime behavior - they only differ in how plugins are instantiated
func main() {
	_ = godotenv.Load(".env")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	config := loadConfig()

	pluginsList := bootstrap.BuildPluginsFromConfig(config)

	auth := authula.New(&authula.AuthConfig{
		Config:  config,
		Plugins: pluginsList,
	})

	// Channel to signal restart
	restartChan := make(chan struct{})
	// Channel to signal shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	// Server loop with restart capability
	for {
		runServer(logger, auth, restartChan, shutdownChan)
	}
}

// runServer starts the HTTP server and handles restarts
func runServer(logger authulamodels.Logger, auth *authula.Auth, restartChan chan struct{}, shutdownChan chan os.Signal) {
	port := getEnv(authulaenv.EnvPort, "8080")

	// Create HTTP server with graceful shutdown support
	server := &http.Server{
		Addr:    ":" + port,
		Handler: auth.Handler(),
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("Starting Authula standalone server", "port", port)
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for shutdown, restart, or server error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("server error: %w", err))
		}
		return

	case <-restartChan:
		logger.Info("Restarting server due to configuration change")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error", "error", err)
		}
		if err := auth.ClosePlugins(); err != nil {
			logger.Error("Failed to close plugins", "error", err)
		}
		if err := auth.CloseSystems(); err != nil {
			logger.Error("Failed to close core systems", "error", err)
		}
		return

	case sig := <-shutdownChan:
		logger.Info("Shutdown signal received", "signal", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error", "error", err)
		}
		if err := auth.ClosePlugins(); err != nil {
			logger.Error("Failed to close plugins", "error", err)
		}
		if err := auth.CloseSystems(); err != nil {
			logger.Error("Failed to close core systems", "error", err)
		}
		os.Exit(0)
	}
}

// loadConfig loads configuration with proper precedence:
func loadConfig() *authulamodels.Config {
	configPath := getEnv(authulaenv.EnvConfigPath, "config.toml")

	config, exists, err := configloader.Load(configPath)
	if err != nil {
		panic(err)
	}
	if !exists {
		slog.Debug("No config file found, continuing", "path", configPath)
	}

	return config
}
