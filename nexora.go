package nexora

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/valyala/bytebufferpool"
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

// MethodWild wild HTTP method
const MethodWild = "*"

var questionMark = byte('?')

type Handler func(c *Context) error

type Nexora struct {
	trees              []*tree
	treeMutable        bool
	customMethodsIndex map[string]int
	registeredPaths    map[string][]string
	namedRoutes        map[string]*Route // Maps route names to paths

	RouteGroup // Default route group for new routes

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 308 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 308 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// If enabled, the router automatically replies to OPTIONS requests.
	// Custom OPTIONS handlers take priority over automatic replies.
	HandleOPTIONS bool

	// An optional http.Handler that is called on automatic OPTIONS requests.
	// The handler is only called if HandleOPTIONS is true and no OPTIONS
	// handler for the specific path was set.
	// The "Allowed" header is set before calling the handler.
	GlobalOPTIONS Handler

	// Configurable http.Handler which is called when no matching route is
	// found. If it is not set, default NotFound is used.
	NotFound Handler

	// Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, ctx.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	MethodNotAllowed Handler

	// Cached value of global (*) allowed methods
	globalAllowed string

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
	PanicHandler func(c *Context, v any) error

	ErrorHandler func(c *Context, err error) error

	pool *sync.Pool // Pool for Context objects
}

// New creates a new instance of Nexora with default settings.
func New() *Nexora {
	nexora := &Nexora{
		trees:                  make([]*tree, 10),
		customMethodsIndex:     make(map[string]int),
		registeredPaths:        make(map[string][]string),
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
		namedRoutes:            make(map[string]*Route),
	}
	nexora.RouteGroup = *newRouteGroup(nexora, "", make([]Handler, 0))
	nexora.pool = &sync.Pool{
		New: func() any {
			return newContext(nexora)
		},
	}
	return nexora
}

// Route returns the named route.
// Nil is returned if the named route cannot be found.
func (r *Nexora) Route(name string) *Route {
	return r.namedRoutes[name]
}

// RegisterCustomMethod registers a custom HTTP method with an index.
func (n *Nexora) methodIndexOf(method string) int {
	switch method {
	case http.MethodGet:
		return 0
	case http.MethodHead:
		return 1
	case http.MethodPost:
		return 2
	case http.MethodPut:
		return 3
	case http.MethodPatch:
		return 4
	case http.MethodDelete:
		return 5
	case http.MethodConnect:
		return 6
	case http.MethodOptions:
		return 7
	case http.MethodTrace:
		return 8
	case MethodWild:
		return 9
	}

	if i, ok := n.customMethodsIndex[method]; ok {
		return i
	}

	return -1
}

// PanicHandler is the default panic handler that recovers from panics in handlers.
func (n *Nexora) recv(c *Context) {
	if rcv := recover(); rcv != nil {
		if err := n.PanicHandler(c, rcv); err != nil {
			n.handleError(c, err)
		}
	}
}

// Handle registers a handler for a specific HTTP method and path.
// It is used under Get, Post, Put, Patch, Delete, Connect, Options, Trace and Wild methods.
// If the path contains optional parameters, it will register all possible paths.
// It panics if the method is empty or no handlers are provided.
// The path must start with a '/' character.
// If the path is invalid, it panics with an error message.
func (n *Nexora) Handle(method, path string, handlers ...Handler) {
	switch {
	case len(method) == 0:
		panic("nexora: method must not be empty")
	case len(handlers) == 0:
		panic("nexora: at least one handler must be provided")
	default:
		validatePath(path)
	}

	path = parseConstraintsRoute(path)

	n.registeredPaths[method] = append(n.registeredPaths[method], path)

	methodIndex := n.methodIndexOf(method)
	if methodIndex == -1 {
		tree := newTree()
		tree.Mutable = n.treeMutable

		n.trees = append(n.trees, tree)
		methodIndex = len(n.trees) - 1
		n.globalAllowed = n.allowed("*", "")
	}

	tree := n.trees[methodIndex]
	if tree == nil {
		tree = newTree()
		tree.Mutable = n.treeMutable
		n.trees[methodIndex] = tree
		n.globalAllowed = n.allowed("*", "")
	}

	optionalPaths := getOptionalPaths(path)
	if len(optionalPaths) == 0 {
		// No optional paths, add the path as is
		tree.Add(path, handlers)
	} else {
		// Add all optional paths
		for _, p := range optionalPaths {
			tree.Add(p, handlers)
		}
	}
}

// allowed returns a comma-separated string of allowed HTTP methods for the given path.
// If the path is "*" or "/*", it returns all registered methods except OPTIONS.
// If the path is not found, it returns an empty string.
// If the request method is specified, it checks if that method is allowed for the path.
// If the path is not found, it returns an empty string.
// The returned string is sorted in ascending order of HTTP methods.
func (n *Nexora) allowed(path, reqMethod string) (allow string) {
	allowed := make([]string, 0, 9)

	if path == "*" || path == "/*" {
		if reqMethod == "" {
			for method := range n.registeredPaths {
				if method == http.MethodOptions {
					continue
				}
				allowed = append(allowed, method)
			}
		} else {
			return n.globalAllowed
		}
	} else {
		for method := range n.registeredPaths {
			if method == reqMethod || method == http.MethodOptions {
				continue
			}

			methodIndex := n.methodIndexOf(method)
			if methodIndex < 0 || methodIndex >= len(n.trees) {
				continue // skip invalid or uninitialized method
			}

			tree := n.trees[methodIndex]
			if tree == nil {
				continue // skip nil tree
			}

			handle, _, _ := tree.Get(path)
			if handle != nil {
				allowed = append(allowed, method)
			}
		}
	}

	if len(allowed) > 0 {
		allowed = append(allowed, http.MethodOptions)

		// Insertion sort to avoid unnecessary allocations
		for i := 1; i < len(allowed); i++ {
			for j := i; j > 0 && allowed[j] < allowed[j-1]; j-- {
				allowed[j], allowed[j-1] = allowed[j-1], allowed[j]
			}
		}

		return strings.Join(allowed, ", ")
	}

	return
}

// validatePath checks if the provided path is valid.
// It panics if the path does not start with a '/' or is empty.
func validatePath(path string) {
	switch {
	case len(path) == 0 || !strings.HasPrefix(path, "/"):
		panic("nexora: path must begin with '/' in path '" + path + "'")
	}
}

// tryRedirect attempts to redirect the request if the path is not found.
// It checks if the path has a trailing slash and redirects accordingly.
// If the request method is GET, it uses a 301 Moved Permanently status code.
// For other methods, it uses a 308 Permanent Redirect status code.
// If the RedirectTrailingSlash option is enabled, it redirects to the path without the trailing slash.
// If the RedirectFixedPath option is enabled, it tries to fix the request path
// by removing superfluous elements like '../' or '//' and redirects to the corrected path.
func (n *Nexora) tryRedirect(w http.ResponseWriter, r *http.Request, tree *tree, tsr bool, method, path string) bool {
	// Moved Permanently, request with GET method
	code := http.StatusMovedPermanently
	if method != http.MethodGet {
		// Permanent Redirect, request with same method
		code = http.StatusPermanentRedirect
	}

	if tsr && n.RedirectTrailingSlash {
		uri := bytebufferpool.Get()

		if len(path) > 1 && path[len(path)-1] == '/' {
			uri.SetString(path[:len(path)-1])
		} else {
			uri.SetString(path)
			uri.WriteByte('/')
		}

		if queryBuf := r.URL.RawQuery; len(queryBuf) > 0 {
			uri.WriteByte(questionMark)
			uri.Write([]byte(queryBuf))
		}

		http.Redirect(w, r, uri.String(), code)
		// ctx.Redirect(uri.String(), code)
		bytebufferpool.Put(uri)

		return true
	}

	// Try to fix the request path
	if n.RedirectFixedPath {
		path2 := r.URL.RawPath

		uri := bytebufferpool.Get()
		found := tree.FindCaseInsensitivePath(
			cleanPath(path2),
			n.RedirectTrailingSlash,
			uri,
		)

		if found {
			if queryBuf := r.URL.RawQuery; len(queryBuf) > 0 {
				uri.WriteByte(questionMark)
				uri.Write([]byte(queryBuf))
			}

			// ctx.Redirect(uri.String(), code)
			http.Redirect(w, r, uri.String(), code)
			bytebufferpool.Put(uri)

			return true
		}

		bytebufferpool.Put(uri)
	}

	return false
}

func (n *Nexora) handleError(c *Context, err error) {
	if n.ErrorHandler != nil {
		if handlerErr := n.ErrorHandler(c, err); handlerErr != nil {
			// NOTE: Replace it later with nexora custom logger
			log.Printf("ErrorHandler failed: %v", handlerErr)
			http.Error(c.ResponseWriter(), "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	if httpErr, ok := err.(*HTTPError); ok {
		http.Error(c.ResponseWriter(), httpErr.Message, httpErr.StatusCode)
	} else {
		// NOTE: Replace it later with nexora custom logger
		log.Printf("Unhandled error: %v", err)
		http.Error(c.ResponseWriter(), "Internal Server Error", http.StatusInternalServerError)
	}
}

// ServeHTTP implements the http.Handler interface for Nexora.
// It processes incoming HTTP requests, routing them to the appropriate handlers
func (n *Nexora) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := n.pool.Get().(*Context)
	defer func() {
		n.recv(c)
		n.pool.Put(c)
	}()

	c.init(r, w)

	path := r.URL.Path
	method := r.Method
	methodIndex := n.methodIndexOf(method)

	if methodIndex > -1 {
		if tree := n.trees[methodIndex]; tree != nil {
			if handlers, params, tsr := tree.Get(path); handlers != nil {
				c.params = params
				c.handlers = handlers
				if err := c.Next(); err != nil {
					n.handleError(c, err)
				}
				return
			} else if method != MethodConnect && path != "/" {
				if ok := n.tryRedirect(w, r, tree, tsr, method, path); ok {
					return
				}
			}
		}
	}

	if tree := n.trees[n.methodIndexOf(MethodWild)]; tree != nil {
		if handler, params, tsr := tree.Get(path); handler != nil {
			c.params = params
			c.handlers = handler
			if err := c.Next(); err != nil {
				n.handleError(c, err)
				return
			} else if method != MethodConnect && path != "/" {
				if ok := n.tryRedirect(w, r, tree, tsr, method, path); ok {
					return
				}
			}

		}
	}

	if n.HandleOPTIONS && method == MethodOptions {
		allow := n.allowed(path, MethodOptions)
		if allow == "" {
			allow = n.allowed("*", MethodOptions)
		}

		if allow != "" {
			w.Header().Set("Allow", allow)
			if n.GlobalOPTIONS != nil {
				if err := n.GlobalOPTIONS(c); err != nil {
					n.handleError(c, err)
					return
				}
			}
			return
		}
	} else if n.HandleMethodNotAllowed {
		if allow := n.allowed(path, method); allow != "" {
			w.Header().Set("Allow", allow)
			if n.MethodNotAllowed != nil {
				if err := n.MethodNotAllowed(c); err != nil {
					n.handleError(c, err)
				}
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}
	}

	if n.NotFound != nil {
		if err := n.NotFound(c); err != nil {
			n.handleError(c, err)
			return
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// Run starts the HTTP server on the specified address.
// It uses the ServeHTTP method to handle incoming requests.
// The address should be in the format "host:port", e.g., ":8080" or "localhost:8080".
// It returns an error if the server fails to start.
// Example usage: nexora.Run(":8080")
// If the address is empty, it defaults to ":8080".
func (n *Nexora) Run(addr string) error {
	server := &http.Server{
		Addr:    addr,
		Handler: n,
	}

	return server.ListenAndServe()
}
