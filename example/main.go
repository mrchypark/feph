package main

import (
	"github.com/gofiber/fiber"
	"github.com/gofiber/logger"
)

func main() {
	// Fiber instance
	app := fiber.New()
	lcfg := logger.Config{
		Format:     "${time} example! ${method} ${path} - ${status} - ${latency}\nRequest :\n${body}\n",
		TimeFormat: "2006-01-02T15:04:05-0700",
	}

	app.Use(logger.New(lcfg))
	// Routes
	app.Post("/helloPost", hello)
	app.Get("/helloGet", hello)
	// Start server
	app.Listen(3000)
}

// Handler
func hello(c *fiber.Ctx) {
	c.JSON(fiber.Map{
    "return": "Hello, World 👋!",
    })
}
