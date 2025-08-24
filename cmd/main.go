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
	counter := 0
	router.Get("/", func(c *nexora.Context) error {
		counter++
		if counter == 1 {
			return c.SendStatus(200)
		} else if counter == 2 {
			return c.SendStatus(404)
		}
		return c.SendStatus(500)
	})
	router.Run(":3000")
}
