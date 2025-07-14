package nexora

import (
	"net/http"
)

// Context represents the context of a single HTTP request in the Nexora framework.
//
// It contains the HTTP request, response writer, URL parameters, and the middleware handler chain.
// Context provides helper methods for accessing request data, sending responses, and controlling
// request flow (e.g., aborting or continuing handler execution).
type Context struct {
	params   map[string]string // URL parameters extracted from the request path.
	request  *http.Request     // The incoming HTTP request.
	writer   *ResponseWriter   // Custom response writer that wraps http.ResponseWriter.
	index    int               // Current index in the handler chain.
	handlers []Handler         // Middleware/handler chain.
	nexora   *Nexora           // Reference to the Nexora app instance.
}

// newContext creates and returns a new Context for the given Nexora instance.
// This is typically used internally by the Nexora router.
func newContext(nexora *Nexora) *Context {
	return &Context{
		nexora: nexora,
	}
}

// Nexora returns the parent Nexora instance associated with this context.
func (c *Context) Nexora() *Nexora {
	return c.nexora
}

// init initializes the context for a new HTTP request.
func (c *Context) init(request *http.Request, writer http.ResponseWriter) {
	c.request = request
	c.writer = NewResponseWriter(writer)
	c.index = -1
}

// Next executes the next handler in the middleware chain.
//
// If a handler returns an error, execution is halted and the error is returned.
// If all handlers run successfully, it returns nil.
func (c *Context) Next() error {
	c.index++
	for n := len(c.handlers); c.index < n; c.index++ {
		if err := c.handlers[c.index](c); err != nil {
			return err
		}
	}
	return nil
}

// Abort stops the execution of any remaining handlers in the chain.
func (c *Context) Abort() {
	c.index = len(c.handlers)
}

// Request returns the original *http.Request associated with this context.
func (c *Context) Request() *http.Request {
	return c.request
}

// ResponseWriter returns the custom ResponseWriter used to send the response.
func (c *Context) ResponseWriter() *ResponseWriter {
	return c.writer
}

// Params returns all route parameters as a map[string]string.
func (c *Context) Params() map[string]string {
	return c.params
}

// Param returns the value of a specific route parameter by name.
//
// If the parameter does not exist, it returns an empty string.
func (c *Context) Param(name string) string {
	return c.params[name]
}

// SendString sends a plain text response with the given string content.
//
// It writes directly to the response writer and returns any write error.
func (c *Context) SendString(s string) error {
	_, err := c.writer.Write([]byte(s))
	return err
}

// SendStatus sets the HTTP status code in the response without writing any body.
func (c *Context) SendStatus(code int) error {
	c.ResponseWriter().WriteHeader(code)
	return nil
}

// Status sets the HTTP status code and returns the context for method chaining.
//
// Example:
//
//	c.Status(404).SendString("Not found")
func (c *Context) Status(code int) *Context {
	c.ResponseWriter().WriteHeader(code)
	return c
}

// Method returns the HTTP method (GET, POST, etc.) of the request.
func (c *Context) Method() string {
	return c.request.Method
}

// Headers returns the request headers as a map.
func (c *Context) Headers() map[string][]string {
	return c.request.Header
}

// Path returns the URL path of the incoming request.
func (c *Context) Path() string {
	return c.request.URL.Path
}
