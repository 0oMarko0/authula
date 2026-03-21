package configmanager

import (
	"github.com/uptrace/bun"

	"github.com/Authula/authula/models"
	"github.com/Authula/authula/services"
)

// NewConfigManager creates a config manager based on the runtime mode and settings.
func NewConfigManager(config *models.Config, db bun.IDB, tokenService services.TokenService) models.ConfigManager {
	return NewDatabaseConfigManager(config, db, tokenService)
}
