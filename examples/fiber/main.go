// Package main demonstrates mounting Authula in a Fiber application
// using the Fiber adapter.
package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"

	authula "github.com/Authula/authula"
	"github.com/Authula/authula/config"
	"github.com/Authula/authula/models"

	fiberadapter "github.com/Authula/authula/adapters/fiber"

	emailpasswordplugin "github.com/Authula/authula/plugins/email-password"
	emailpasswordplugintypes "github.com/Authula/authula/plugins/email-password/types"

	emailplugin "github.com/Authula/authula/plugins/email"
	emailplugintypes "github.com/Authula/authula/plugins/email/types"

	sessionplugin "github.com/Authula/authula/plugins/session"
)

func main() {
	// Configure authula.
	authCfg := config.NewConfig(
		config.WithAppName("Fiber Example"),
		config.WithBaseURL("http://localhost:3000"),
		config.WithBasePath("/api/auth"),
		config.WithSecret("change-me-to-a-real-secret-at-least-32-chars"),
		config.WithDatabase(models.DatabaseConfig{
			Provider: "postgres",
			URL:      "postgresql://user:pass@localhost:5432/mydb",
		}),
		config.WithSession(models.SessionConfig{
			CookieName: "authula.session_token",
			ExpiresIn:  7 * 24 * time.Hour,
			HttpOnly:   true,
			SameSite:   "lax",
		}),
	)

	// Assemble plugins.
	plugins := []models.Plugin{
		emailplugin.New(emailplugintypes.EmailPluginConfig{
			Enabled:  true,
			Provider: emailplugintypes.ProviderSMTP,
		}),
		emailpasswordplugin.New(emailpasswordplugintypes.EmailPasswordPluginConfig{
			Enabled:           true,
			MinPasswordLength: 8,
			MaxPasswordLength: 128,
			AutoSignIn:        true,
		}),
		sessionplugin.New(sessionplugin.SessionPluginConfig{
			Enabled: true,
		}),
	}

	// Create Authula instance.
	auth := authula.New(&authula.AuthConfig{
		Config:  authCfg,
		Plugins: plugins,
	})

	// Create Fiber app.
	app := fiber.New()
	app.Use(logger.New())

	// Mount using the Fiber adapter.
	app.Use("/api/auth", fiberadapter.New(fiberadapter.Config{
		Handler: auth.Handler(),
	}))

	// Example: a public route.
	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Welcome! Auth is at /api/auth"})
	})

	log.Println("Starting server on :3000")
	log.Fatal(app.Listen(":3000"))
}
