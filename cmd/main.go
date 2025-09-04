package main

import (
	"fmt"

	"github.com/Abhishek2010dev/nexora"
)

type User struct {
	ID   string `param:"id" json:"id"`
	Name string `param:"name" json:"name"`
}

func main() {
	app := nexora.New()
	app.Get("/{id:int}/{name}", func(c *nexora.Context) error {
		var u User
		if err := c.BindParams(&u); err != nil {
			return err
		}
		fmt.Println(u.ID)
		return c.SendJSON(u)
	})
	app.Run(":8080")
}
