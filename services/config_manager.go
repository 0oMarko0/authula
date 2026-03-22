package services

import (
	"context"

	"github.com/0oMarko0/authula/models"
)

type ConfigManagerService interface {
	GetConfig(ctx context.Context) (*models.Config, error)
	GetAuthSettings(ctx context.Context) (map[string]any, error)
	GetPluginConfig(ctx context.Context, pluginName string) (any, error)
	RegisterConfigWatcher(pluginID string, plugin models.PluginWithConfigWatcher) error
	NotifyWatchers(config *models.Config) error
}
