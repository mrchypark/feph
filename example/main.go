package main

import (
	"time"

	"github.com/gofiber/fiber"
)

func main() {
	// Fiber instance
	app := fiber.New()
	// Routes
	app.Post("/helloPostList", helloList)
	app.Post("/helloPost", hello)
	app.Get("/helloGet", hello)
	app.Get("/text", text)
	app.Get("/time", times)
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

func times(c *fiber.Ctx) {
	time.Sleep(30 * time.Second)
	c.Send("Hello, World 👋!")
}

// Handler
func text(c *fiber.Ctx) {
	c.Send("Hello, World 👋!")
}
