package main

import (
	"github.com/Abhishek2010dev/nexora"
)

func main() {
	router := nexora.New(&nexora.Config{
		LoggerConfig: &nexora.LoggerConfig{
			Production: true,
		},
	})
	router.Get("/", func(c *nexora.Context) error {
		return c.SendString("Hello, World")
	})
	router.Run(":3000")
}
