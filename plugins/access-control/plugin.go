package accesscontrol

import (
	"github.com/0oMarko0/authula/internal/util"
	"github.com/0oMarko0/authula/migrations"
	"github.com/0oMarko0/authula/models"
	"github.com/0oMarko0/authula/plugins/access-control/repositories"
	"github.com/0oMarko0/authula/plugins/access-control/services"
	"github.com/0oMarko0/authula/plugins/access-control/types"
	"github.com/0oMarko0/authula/plugins/access-control/usecases"
)

type AccessControlPlugin struct {
	config types.AccessControlPluginConfig
	ctx    *models.PluginContext
	logger models.Logger
	Api    *API
}

func New(config types.AccessControlPluginConfig) *AccessControlPlugin {
	config.ApplyDefaults()
	return &AccessControlPlugin{config: config}
}

func (p *AccessControlPlugin) Metadata() models.PluginMetadata {
	return models.PluginMetadata{
		ID:          models.PluginAccessControl.String(),
		Version:     "1.0.0",
		Description: "Provides access control functionality.",
	}
}

func (p *AccessControlPlugin) Config() any {
	return p.config
}

func (p *AccessControlPlugin) Init(ctx *models.PluginContext) error {
	p.ctx = ctx
	p.logger = ctx.Logger

	if err := util.LoadPluginConfig(ctx.GetConfig(), p.Metadata().ID, &p.config); err != nil {
		return err
	}

	rolePermissionRepo := repositories.NewBunRolePermissionRepository(ctx.DB)
	userAccessRepo := repositories.NewBunUserAccessRepository(ctx.DB)
	rolePermissionService := services.NewRolePermissionService(rolePermissionRepo)
	userAccessService := services.NewUserAccessService(userAccessRepo)

	useCases := usecases.NewAccessControlUseCases(
		usecases.NewRolePermissionUseCase(rolePermissionService),
		usecases.NewUserRolesUseCase(userAccessService),
	)
	p.Api = NewAPI(
		useCases,
		rolePermissionRepo,
		userAccessRepo,
	)

	return nil
}

func (p *AccessControlPlugin) Migrations(provider string) []migrations.Migration {
	return accessControlMigrationsForProvider(provider)
}

func (p *AccessControlPlugin) DependsOn() []string {
	return []string{}
}

func (p *AccessControlPlugin) Routes() []models.Route {
	return Routes(p.Api)
}

func (p *AccessControlPlugin) Close() error {
	return nil
}
