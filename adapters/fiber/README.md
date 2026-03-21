# Fiber Adapter for Authula

A [Fiber](https://gofiber.io/) middleware that bridges Authula's `http.Handler` to Fiber's fasthttp-based context.

## Why?

Fiber uses [fasthttp](https://github.com/valyala/fasthttp) under the hood, not `net/http`. Authula exposes a standard `http.Handler`. While Go provides `fasthttpadaptor`, it can lose the request body — a critical issue for authentication payloads like sign-in and sign-up.

This adapter manually builds `net/http` requests from Fiber's context, ensuring full request body preservation and correct header propagation (including multi-value `Set-Cookie` headers).

## Installation

```bash
go get github.com/Authula/authula
```

## Usage

```go
package main

import (
	"log"

	"github.com/gofiber/fiber/v3"

	authula "github.com/Authula/authula"
	authulaconfig "github.com/Authula/authula/config"
	fiberadapter "github.com/Authula/authula/adapters/fiber"
)

func main() {
	// 1. Create Authula instance
	auth := authula.New(&authula.AuthConfig{
		Config: authulaconfig.NewConfig(
			authulaconfig.WithBaseURL("http://localhost:3000"),
			authulaconfig.WithBasePath("/api/auth"),
			// ... your config
		),
		// ... your plugins
	})

	// 2. Create Fiber app
	app := fiber.New()

	// 3. Mount on /api/auth
	app.Use("/api/auth", fiberadapter.New(fiberadapter.Config{
		Handler: auth.Handler(),
	}))

	// 4. Your app routes
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello!")
	})

	log.Fatal(app.Listen(":3000"))
}
```

## Config

| Field          | Type                   | Required | Default  | Description                               |
| -------------- | ---------------------- | -------- | -------- | ----------------------------------------- |
| `Handler`      | `http.Handler`         | Yes      | —        | The Authula handler from `auth.Handler()` |
| `Next`         | `func(fiber.Ctx) bool` | No       | `nil`    | Skip middleware when returning `true`     |
| `ErrorHandler` | `fiber.ErrorHandler`   | No       | 500 JSON | Called on internal adapter errors         |

## How It Works

1. Reads the raw request body from Fiber's context
2. Builds a standard `*http.Request` with all headers, query params, and body
3. Creates a `ResponseWriter` that captures the response back into Fiber
4. Calls `handler.ServeHTTP()` with the bridged request/response

The key insight is bypassing `fasthttpadaptor` — it uses `fasthttp.Request.BodyStream()` which can return an empty reader for certain request types, causing authentication to fail silently.
