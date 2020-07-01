package main

import (
	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware"
)

func main() {
	// Fiber instance
	app := fiber.New()

	app.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format:     "${time} example log ${method} ${path} - ${status} - ${latency}\nRequest : ${body}\n",
		TimeFormat: "2006-01-02T15:04:05-0700",
	}))
	// Routes
	app.Post("/helloPostList", helloList)
	app.Post("/helloPost", hello)
	app.Get("/helloGet", hello)
	app.Get("/text", text)
	// Start server
	app.Listen(3000)
}

// Handler
func hello(c *fiber.Ctx) {
	c.JSON(fiber.Map{
		"return": "Hello, World 👋!",
	})
}

// Handler
func helloList(c *fiber.Ctx) {
	list := []fiber.Map{}
	list = append(list, fiber.Map{"return": "Hello, World 👋!"})
	list = append(list, fiber.Map{"return": "Hello, World 👋!"})
	c.JSON(list)
}

// Handler
func text(c *fiber.Ctx) {
	c.Send("Hello, World 👋!")
}
