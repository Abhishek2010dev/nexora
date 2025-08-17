package main

import "github.com/Abhishek2010dev/nexora"

func main() {
	router := nexora.New()
	router.Get("/", func(c *nexora.Context) error {
		return c.SendString("Ok")
	})
	router.Run(":3000")
}
