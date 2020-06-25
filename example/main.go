package main

import (
	"log"

	"github.com/gofiber/fiber"
)

func main() {
	// Fiber instance
	app := fiber.New()
	// Routes
	app.Get("/hello", hello)
	// Start server
	log.Fatal(app.Listen(3000))
}

// Handler
func hello(c *fiber.Ctx) {
	c.Send("Hello, World 👋!")
}
