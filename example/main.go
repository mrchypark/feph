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