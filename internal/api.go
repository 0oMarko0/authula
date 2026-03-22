package internal

import (
	"context"

	"github.com/0oMarko0/authula/internal/handlers"
	"github.com/0oMarko0/authula/internal/types"
	"github.com/0oMarko0/authula/internal/usecases"
	"github.com/0oMarko0/authula/models"
	"github.com/0oMarko0/authula/services"
)

type CoreAPI interface {
	GetMe(ctx context.Context, userID string) (*types.GetMeResult, error)
	SignOut(ctx context.Context, userID string, sessionID *string, signOutAll bool) (*types.SignOutResult, error)
}

// coreAPI implements the CoreAPI interface.
type coreAPI struct {
	useCases *usecases.UseCases
}

// NewCoreAPI creates a new CoreAPI instance.
func NewCoreAPI(logger models.Logger, userService services.UserService, sessionService services.SessionService) CoreAPI {
	useCases := BuildUseCases(logger, userService, sessionService)
	return &coreAPI{
		useCases: useCases,
	}
}

func (api *coreAPI) GetMe(ctx context.Context, userID string) (*types.GetMeResult, error) {
	return api.useCases.GetMeUseCase.GetMe(ctx, userID)
}

func (api *coreAPI) SignOut(ctx context.Context, userID string, sessionID *string, signOutAll bool) (*types.SignOutResult, error) {
	return api.useCases.SignOutUseCase.SignOut(ctx, userID, sessionID, signOutAll)
}

func CoreRoutes(logger models.Logger, userService services.UserService, sessionService services.SessionService) []models.Route {
	useCases := BuildUseCases(logger, userService, sessionService)

	getMeHandler := &handlers.GetMeHandler{
		UseCase: useCases.GetMeUseCase,
	}

	signOutHandler := &handlers.SignOutHandler{
		UseCase: useCases.SignOutUseCase,
	}

	return []models.Route{
		{
			Method:  "GET",
			Path:    "/me",
			Handler: getMeHandler.Handler(),
		},
		{
			Method:  "POST",
			Path:    "/sign-out",
			Handler: signOutHandler.Handler(),
		},
	}
}

func BuildUseCases(logger models.Logger, userService services.UserService, sessionService services.SessionService) *usecases.UseCases {
	return &usecases.UseCases{
		GetMeUseCase: &usecases.GetMeUseCase{
			Logger:         logger,
			UserService:    userService,
			SessionService: sessionService,
		},
		SignOutUseCase: &usecases.SignOutUseCase{
			Logger:         logger,
			SessionService: sessionService,
		},
	}
}
