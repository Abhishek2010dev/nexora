package nexora

import (
	"sync"
)

// Taken from net/http package in Go standard library
const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH" // RFC 5789
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
)

type Handler func(c *Context) error

type Nexora struct {
	pool *sync.Pool // A pool of Context objects to reuse for performance optimization.

	IgnoreTrailingSlash bool    // If true, Nexora will ignore trailing slashes in the URL.
	useEscapeSlash      bool    // If true, Nexora will escape slashes in the URL.
	NotFoundHandler     Handler // A handler to call when no route matches the request.
	NotAllowedHandler   Handler // A handler to call when the method is not allowed for a route.
}

func New() *Nexora {
	nexora := &Nexora{}
	nexora.pool = &sync.Pool{
		New: func() any {
			return NewContext(nexora)
		},
	}
	return nexora
}
