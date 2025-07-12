package nexora

type Context struct{}

type Handler func(c *Context) error
