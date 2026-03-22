package configmanager

import (
	"net/http"
	"os"

	"github.com/0oMarko0/authula/internal/util"
	"github.com/0oMarko0/authula/plugins/config-manager/constants"
)

func ConfigManagerAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")
		expectedKey := os.Getenv(constants.EnvAdminApiKey)
		if apiKey != expectedKey || apiKey == "" {
			util.JSONResponse(w, http.StatusUnauthorized, map[string]any{
				"message": "missing or invalid API key.",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
