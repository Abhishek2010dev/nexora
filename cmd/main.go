package main

import (
	"github.com/Abhishek2010dev/nexora"
	"github.com/Abhishek2010dev/nexora/middleware"
)

func main() {
	router := nexora.New()
	router.Use(middleware.Logger(&middleware.LoggerConfig{
		LogLatency: true,
		LogIP:      true,
	}))
	router.Get("/", func(c *nexora.Context) error {
		return c.SendString("Hello, World")
	})
	router.Run(":3000")
}
