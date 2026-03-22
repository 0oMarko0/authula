package services

import "github.com/0oMarko0/authula/models"

// SecondaryStorageService provides access to the configured secondary storage backend
type SecondaryStorageService interface {
	// GetStorage returns the configured SecondaryStorage backend
	GetStorage() models.SecondaryStorage
	// GetProviderName returns the name of the currently active provider (e.g., "redis", "database", "memory")
	GetProviderName() string
}
