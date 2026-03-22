package secondarystorage

import "github.com/0oMarko0/authula/models"

// SecondaryStorageAPI is the API exposed by the secondary storage plugin
type SecondaryStorageAPI interface {
	// GetStorage returns the configured SecondaryStorage backend
	GetStorage() models.SecondaryStorage
	// GetProviderName returns the name of the currently active provider
	GetProviderName() string
}
