package nexora

import (
	"net/http"
)

// Context represents the context of a single HTTP request.
// It holds the request, response writer, URL parameters, and a stack of handlers.
// It provides methods to access these values and to control the flow of request handling.
type Context struct {
	params   map[string]string   // URL parameters extracted from the request path.
	request  *http.Request       // The original HTTP request object.
	writer   http.ResponseWriter // The response writer to send the response back to the client.
	index    int                 // The current index in the handlers slice, used to track which handler is currently being executed.
	handlers []Handler           // A slice of handlers to be executed in order for this request.
	nexora   *Nexora             // A reference to the Nexora instance that created this context, allowing access to shared resources and settings.
}

func newContext(nexora *Nexora) *Context {
	return &Context{
		nexora: nexora,
	}
}

// Nexora returns the Nexora instance for the current request.
func (c *Context) Nexora() *Nexora {
	return c.nexora
}

// Init initializes the context with the given request and response writer.
func (c *Context) init(request *http.Request, writer http.ResponseWriter) {
	c.request = request
	c.writer = writer
	c.index = -1
}

// Next executes the next handler in the context's handlers slice.
// It increments the index and calls the handler at the new index.
// If the handler returns an error, it stops execution and returns the error.
// If all handlers are executed successfully, it returns nil.
func (c *Context) Next() error {
	c.index++
	for n := len(c.handlers); c.index < n; c.index++ {
		if err := c.handlers[c.index](c); err != nil {
			return err
		}
	}
	return nil
}

// Request returns the original HTTP request associated with this context.
func (c *Context) Request() *http.Request {
	return c.request
}

// ResponseWriter returns the response writer associated with this context.
func (c *Context) ResponseWriter() http.ResponseWriter {
	return c.writer
}

// Params returns the URL parameters associated with this context.
func (c *Context) Params() map[string]string {
	return c.params
}

// Param retrieves the value of a URL parameter by its name.
func (c *Context) Param(name string) string {
	return c.params[name]
}

// Abort stops the execution of any further handlers in the context's handlers slice.
func (c *Context) Abort() {
	c.index = len(c.handlers) // Skip all remaining handlers
}

func (c *Context) SendString(s string) error {
	_, err := c.writer.Write([]byte(s))
	if err != nil {
		return err
	}
	return nil
}
