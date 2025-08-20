package nexora

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

const ContentTypeOctetStream = "application/octet-stream"
const ContentTypeJson = "application/json"

// Context represents the context of a single HTTP request in the Nexora framework.
//
// It contains the HTTP request, response writer, URL parameters, and the middleware handler chain.
// Context provides helper methods for accessing request data, sending responses, and controlling
// request flow (e.g., aborting or continuing handler execution).
type Context struct {
	params      map[string]string // URL parameters extracted from the request path.
	request     *http.Request     // The incoming HTTP request.
	writer      *ResponseWriter   // Custom response writer that wraps http.ResponseWriter.
	index       int               // Current index in the handler chain.
	handlers    []Handler         // Middleware/handler chain.
	nexora      *Nexora           // Reference to the Nexora app instance.
	queryValues url.Values        // query cached
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
	c.queryValues = nil
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

// Param returns the value of a route parameter by name.
//
// If the parameter is not present and a defaultValue is provided,
// the first element of defaultValue is returned instead.
//
// Example usage:
//
//	id := ctx.Param("id")              // returns "" if not found
//	id := ctx.Param("id", "default")   // returns "default" if not found.
func (c *Context) Param(name string, defaultValue ...string) string {
	if value, ok := c.params[name]; ok {
		return value
	}
	if 0 < len(defaultValue) {
		return defaultValue[0]
	}
	return ""
}

// ParamExists returns the value of a route parameter and a boolean indicating
// whether the parameter was present in the route.
//
// This is useful when you need to distinguish between a parameter that is
// missing and one that is present with an empty value.
//
// Example usage:
//
//	id, ok := ctx.ParamExists("id")
//	if ok {
//	    // Use id
//	} else {
//	    // Handle missing parameter
//	}
func (c *Context) ParamExists(name string) (string, bool) {
	val, ok := c.params[name]
	return val, ok
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

// Path returns the URL path of the incoming request.
func (c *Context) Path() string {
	return c.request.URL.Path
}

// cacheQuery lazily parses and caches the URL query parameters
// from the underlying *http.Request. It is called internally by
// other query-related methods to avoid repeated parsing.
func (c *Context) cacheQuery() {
	if c.queryValues == nil {
		if c.request != nil && c.request.URL != nil {
			c.queryValues = c.request.URL.Query()
		} else {
			c.queryValues = url.Values{}
		}
	}
}

// Queries returns all URL query parameters as a url.Values map.
// It ensures the query parameters are parsed and cached first.
//
// Example:
//
//	values := c.Queries()
//	name := values.Get("name")
func (c *Context) Queries() url.Values {
	c.cacheQuery()
	return c.queryValues
}

// QueryArray returns all values for a given query parameter key.
// If the key is not present, it returns nil.
//
// Example:
//
//	tags := c.QueryArray("tag")
//	// ?tag=go&tag=web → []string{"go", "web"}
func (c *Context) QueryArray(key string) []string {
	c.cacheQuery()
	if vals, ok := c.queryValues[key]; ok {
		return vals
	}
	return nil
}

// Query returns the first value for a given query parameter key.
// If the key is not present, it returns the optional defaultValue
// if provided, or an empty string otherwise.
//
// Example:
//
//	q := c.Query("q")
//	page := c.Query("page", "1") // default to "1" if missing
func (c *Context) Query(key string, defaultValue ...string) string {
	c.cacheQuery()
	if vals, ok := c.queryValues[key]; ok && len(vals) > 0 {
		return vals[0]
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// QueryExists returns the first value for the given query key
// and a boolean indicating whether the key exists.
//
// If the key is present in the query parameters (even if its value is empty),
// it returns (value, true). If the key is not present, it returns ("", false).
//
// Example:
//
//	?foo=bar     -> ("bar", true)
//	?foo=        -> ("", true)
//	(no foo)     -> ("", false)
func (c *Context) QueryExists(key string) (string, bool) {
	c.cacheQuery() // make sure queryValues is initialized
	vals, ok := c.queryValues[key]
	if ok && len(vals) > 0 {
		return vals[0], true
	}
	// key not found
	return "", false
}

// Port returns the server port on which the request was received.
// It parses the Host field of the request to extract the port.
// If no explicit port is present, it falls back to 443 for HTTPS or 80 for HTTP.
func (c *Context) Port() string {
	_, port, err := net.SplitHostPort(c.request.Host)
	if err != nil {
		if c.request.TLS != nil {
			return "443"
		}
		return "80"
	}
	return port
}

// RemotePort returns the remote TCP port from which the client
// is connected. If the remote address cannot be parsed, it returns an empty string.
func (c *Context) RemotePort() string {
	_, port, err := net.SplitHostPort(c.request.RemoteAddr)
	if err != nil {
		return ""
	}
	return port
}

// IP returns the remote IP address of the client that made the request.
// If the remote address cannot be parsed, it returns an empty string.
func (c *Context) IP() string {
	host, _, err := net.SplitHostPort(c.request.RemoteAddr)
	if err != nil {
		return ""
	}
	return host
}

// Headers returns all the HTTP request headers as an http.Header map.
// The returned map can be iterated or queried for multiple values.
func (c *Context) Headers() http.Header {
	return c.request.Header
}

// GetHeader retrieves the value of the specified request header field.
// If the header is not present, it returns an empty string.
func (c *Context) GetHeader(key string) string {
	return c.request.Header.Get(key)
}

// SetHeader sets a header field on the HTTP response.
// It replaces any existing values associated with the key.
func (c *Context) SetHeader(key, value string) {
	c.writer.Header().Set(key, value)
}

// DelHeader deletes the specified header field from the HTTP response.
// If the header is not present, this is a no-op.
func (c *Context) DelHeader(key string) {
	c.writer.Header().Del(key)
}

// AddHeader adds the specified value to the given header field in the HTTP response.
// It appends to any existing values associated with the key.
func (c *Context) AddHeader(key, value string) {
	c.writer.Header().Add(key, value)
}

// SendHeader sets an HTTP header key-value pair on the response.
//
// This method is **sugar syntax**: it always returns `error` (currently always `nil`),
// which matches the typical handler signature in this framework
// (e.g., `func(c *Context) error`).
// That means you can directly return it from your handler without extra wrapping.
//
// Example:
//
//	func H(c *nexora.Context) error {
//	    // Set a custom header and directly return
//	    return c.SendHeader("X-Custom-Header", "my-value")
//	}
//
// Parameters:
//   - key:   The header name (e.g., "X-Custom-Header").
//   - value: The header value.
//
// Returns:
//   - error: Always returns nil (reserved for future use).
func (c *Context) SendHeader(key string, value string) error {
	c.writer.Header().Set(key, value)
	return nil
}

// SetContentType sets the "Content-Type" header on the response.
// This defines the media type of the response body.
//
// Example:
//
//	c.SetContentType("application/json")
//	c.SetContentType("text/html; charset=utf-8")
//
// Parameters:
//   - ct: The content type string (e.g., "application/json").
func (c *Context) SetContentType(ct string) {
	c.writer.Header().Set(HeaderContentType, ct)
}

// RealIP returns the client's real IP address, considering X-Forwarded-For.
func (c *Context) RealIP() string {
	xForwardedFor := c.GetHeader(HeaderForwardedFor)
	if xForwardedFor != "" {
		parts := strings.Split(xForwardedFor, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	return c.IP()
}

// Body returns the raw request body as []byte.
// It reads and caches the body so multiple calls won't re-read the stream.
func (c *Context) Body() []byte {
	data, err := io.ReadAll(c.request.Body)
	if err != nil {
		return nil
	}
	return data
}

// IsAJAX returns true if the request was made via AJAX.
func (c *Context) IsAJAX() bool {
	return c.GetHeader(HeaderXRequestedWith) == "XMLHttpRequest"
}

// BindJson parses the request body as JSON into v.
func (c *Context) BindJson(v any) error {
	err := c.nexora.JsonDecoder(c.Body(), v)
	if err == nil {
		return nil
	}
	// Decode JSON directly from the request body
	if err := c.nexora.JsonDecoder(c.Body(), v); err != nil {
		// Zero-cost, minimal allocation error handling
		switch e := err.(type) {
		case *json.SyntaxError:
			return NewHTTPError(StatusBadRequest, fmt.Sprintf("Malformed JSON at byte offset %d", e.Offset))
		case *json.UnmarshalTypeError:
			return NewHTTPError(StatusBadRequest,
				fmt.Sprintf("Wrong type for field '%s', expected %s at byte offset %d", e.Field, e.Type, e.Offset))
		default:
			return NewHTTPError(StatusBadRequest, err.Error())
		}
	}

	return nil
}

// SendJsonp encodes v as JSON and wraps it in a JavaScript function call
// using the given callback name (e.g., callback(v);). This is commonly
// used to support cross-domain requests without CORS.
//
// Example output:
//
//	callback({"message":"hello"});
//
// If encoding fails or writing fails, an HTTP error is returned.
func (c *Context) SendJsonp(callback string, v any) error {
	body, err := c.nexora.JsonEncoder(v)
	if err != nil {
		return NewHTTPError(StatusInternalServerError,
			fmt.Sprintf("failed to encode JSON: %v", err))
	}

	c.SetContentType("application/javascript")

	// Wrap JSON with callback(...)
	full := []byte(callback + "(" + string(body) + ");")
	return c.setBody(full)
}

// Json is a helper function that parses the JSON body of the request
// into a new instance of type T, reducing boilerplate code.
func Json[T any](c *Context) (*T, error) {
	var value = new(T)
	if err := c.BindJson(value); err != nil {
		return nil, err
	}
	return value, nil
}

// SendJson encodes v as JSON and writes it to the response body.
// Returns an error if encoding fails.
func (c *Context) SendJson(v any) error {
	body, err := c.nexora.JsonEncoder(v)
	if err != nil {
		return NewHTTPError(StatusInternalServerError, fmt.Sprintf("failed to encode JSON: %v", err))
	}
	c.SetContentType(ContentTypeJson)
	return c.setBody(body)
}

// SendPrettyJson encodes the given value `v` into a human-readable JSON format
// (pretty-printed with a single space as indentation). It sets the response
// Content-Type to "application/json" and writes the JSON body to the response.
//
// Returns an error if JSON encoding or writing to the response fails.
func (c *Context) SendPrettyJson(v any) error {
	body, err := c.nexora.JsonIndentationEncoder(v, "", " ")
	if err != nil {
		return NewHTTPError(StatusInternalServerError, fmt.Sprintf("failed to encode JSON: %v", err))
	}
	c.SetContentType(ContentTypeJson)
	return c.setBody(body)
}

// SendIndentJson encodes the given value `v` into indented JSON using the
// provided prefix and indent string. It sets the response Content-Type to
// "application/json" and writes the JSON body to the response.
//
// Parameters:
//   - v: The value to encode as JSON.
//   - prefix: A string to place before each JSON line.
//   - indent: A string used for indentation (e.g., "\t" or "  ").
//
// Returns an error if JSON encoding or writing to the response fails.
func (c *Context) SendIndentJson(v any, prefix, indent string) error {
	body, err := c.nexora.JsonIndentationEncoder(v, prefix, indent)
	if err != nil {
		return NewHTTPError(StatusInternalServerError, fmt.Sprintf("failed to encode JSON: %v", err))
	}
	c.SetContentType(ContentTypeJson)
	return c.setBody(body)
}

// SendSecureJson encodes the given value as JSON and writes it to the response
// with the "application/json" Content-Type header. The output is prefixed with
// c.nexora.secureJsonPrefix (e.g., "while(1);") to mitigate JSON hijacking
// attacks when serving untrusted clients.
//
// If JSON encoding fails or writing to the response fails, an HTTP error is
// returned.
func (c *Context) SendSecureJson(v any) error {
	body, err := c.nexora.JsonEncoder(v)
	if err != nil {
		return NewHTTPError(StatusInternalServerError,
			fmt.Sprintf("failed to encode JSON: %v", err))
	}

	c.SetContentType(ContentTypeJson)

	full := append(c.nexora.secureJsonPrefix, body...)
	return c.setBody(full)
}

// SendBytes sets the Content-Type to "application/octet-stream" and writes
// the given byte slice to the HTTP response. Returns an HTTP error if writing fails
// or not all bytes are sent.
func (c *Context) SendBytes(b []byte) error {
	c.SetContentType(ContentTypeOctetStream)
	return c.setBody(b)
}

// SendByte writes the given byte slice to the HTTP response without
// modifying the Content-Type. Returns an HTTP error if writing fails
// or not all bytes are sent.
//
// For sending a single byte, consider using SendBytes([]byte{b}).
func (c *Context) SendByte(b []byte) error {
	return c.setBody(b)
}

// setBody is an internal helper that writes the provided byte slice
// to the HTTP response. It returns an HTTP error if writing fails
// or if the number of bytes written does not match the slice length.
//
// This function is private and should not be called directly outside
// the Context implementation.
func (c *Context) setBody(b []byte) error {
	n, err := c.ResponseWriter().Write(b)
	if err != nil {
		return NewHTTPError(StatusInternalServerError, fmt.Sprintf("failed to send bytes: %v", err))
	}
	if n != len(b) {
		return NewHTTPError(StatusInternalServerError, "failed to send complete bytes")
	}
	return nil
}
