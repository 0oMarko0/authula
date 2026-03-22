package handlers

import (
	"net/http"

	"github.com/0oMarko0/authula/models"
	"github.com/0oMarko0/authula/plugins/totp/types"
	"github.com/0oMarko0/authula/plugins/totp/usecases"
)

type GenerateBackupCodesHandler struct {
	UseCase *usecases.GenerateBackupCodesUseCase
}

func (h *GenerateBackupCodesHandler) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		reqCtx, _ := models.GetRequestContext(ctx)

		userID, ok := models.GetUserIDFromContext(ctx)
		if !ok {
			reqCtx.SetJSONResponse(http.StatusUnauthorized, map[string]any{
				"message": "Unauthorized",
			})
			reqCtx.Handled = true
			return
		}

		codes, err := h.UseCase.Generate(ctx, userID)
		if err != nil {
			reqCtx.SetJSONResponse(http.StatusBadRequest, map[string]any{
				"message": err.Error(),
			})
			reqCtx.Handled = true
			return
		}

		reqCtx.SetJSONResponse(http.StatusOK, &types.GenerateBackupCodesResponse{
			BackupCodes: codes,
		})
	}
}
