package main

import (
	"github.com/Abhishek2010dev/nexora"
)

func main() {
	router := nexora.New()
	router.Get("/", func(c *nexora.Context) error {
		return c.SendFile("codex.go")
	})
	router.Run(":3000")
}
