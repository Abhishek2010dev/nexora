package nexora

import (
	"strings"
)

// RouteGroup represents a group of routes that share a common prefix and handlers.
type RouteGroup struct {
	prefix   string    // The prefix for the route group, used to create a common path for all routes in this group.
	nexora   *Nexora   // A reference to the Nexora instance that created this group, allowing access to shared resources and settings.
	handlers []Handler // A slice of handlers that will be applied to all routes in this group.
}

// newRouteGroup creates a new RouteGroup with the specified prefix and handlers.
func newRouteGroup(nexora *Nexora, prefix string, handlers []Handler) *RouteGroup {
	return &RouteGroup{
		prefix:   prefix,
		nexora:   nexora,
		handlers: handlers,
	}
}

func (g *RouteGroup) Use(handlers ...Handler) {
	g.handlers = append(g.handlers, handlers...)
}

// Group creates a RouteGroup with the given route path prefix and handlers.
// The new group will combine the existing path prefix with the new one.
// If no handler is provided, the new group will inherit the handlers registered
// with the current group.
func (g *RouteGroup) Group(prefix string, handlers ...Handler) *RouteGroup {
	if len(handlers) == 0 {
		handlers = make([]Handler, len(g.handlers))
		copy(handlers, g.handlers)
	}
	return newRouteGroup(g.nexora, g.prefix+prefix, handlers)
}

// Get registers a new GET route with the specified path and handlers.
func (g *RouteGroup) Get(path string, handler ...Handler) *Route {
	return g.add(MethodGet, path, handler)
}

// Post registers a new POST route with the specified path and handlers.
func (g *RouteGroup) Post(path string, handler ...Handler) *Route {
	return g.add(MethodPost, path, handler)
}

// Put registers a new PUT route with the specified path and handlers.
func (g *RouteGroup) Put(path string, handler ...Handler) *Route {
	return g.add(MethodPut, path, handler)
}

// Delete registers a new DELETE route with the specified path and handlers.
func (g *RouteGroup) Delete(path string, handler ...Handler) *Route {
	return g.add(MethodDelete, path, handler)
}

// Patch registers a new PATCH route with the specified path and handlers.
func (g *RouteGroup) Patch(path string, handler ...Handler) *Route {
	return g.add(MethodPatch, path, handler)
}

// Head registers a new HEAD route with the specified path and handlers.
func (g *RouteGroup) Head(path string, handler ...Handler) *Route {
	return g.add(MethodHead, path, handler)
}

// Options registers a new OPTIONS route with the specified path and handlers.
func (g *RouteGroup) Options(path string, handler ...Handler) *Route {
	return g.add(MethodOptions, path, handler)
}

// Trace registers a new TRACE route with the specified path and handlers.
func (g *RouteGroup) Trace(path string, handler ...Handler) *Route {
	return g.add(MethodTrace, path, handler)
}

// Connect registers a new CONNECT route with the specified path and handlers.
func (g *RouteGroup) Connect(path string, handler ...Handler) *Route {
	return g.add(MethodConnect, path, handler)
}

// add registers a new route with the specified method and path, combining the group's handlers with the provided handlers.
func (g *RouteGroup) add(method, path string, handler []Handler) *Route {
	r := g.newRoute(method, path)
	g.nexora.Handle(method, g.prefix+r.path, combineHandlers(g.handlers, handler)...)
	return r
}

// newRoute creates a new Route associated with this RouteGroup.
func (g *RouteGroup) newRoute(method, path string) *Route {
	return &Route{
		group:    g,
		method:   method,
		path:     path,
		template: buildURLTemplate(path),
	}
}

// combineHandlers merges two slices of handlers into one.
func combineHandlers(h1 []Handler, h2 []Handler) []Handler {
	hh := make([]Handler, 0, len(h1)+len(h2))
	hh = append(hh, h1...)
	hh = append(hh, h2...)
	return hh
}

func buildURLTemplate(path string) string {
	path = strings.TrimRight(path, "*")

	var (
		template strings.Builder
		start    = -1
		end      = -1
	)

	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '{':
			if start < 0 {
				start = i
			}
		case '}':
			if start >= 0 {
				name := path[start+1 : i]

				// Remove regex
				if colon := strings.IndexByte(name, ':'); colon >= 0 {
					name = name[:colon]
				}

				// Remove trailing * or ? if present (wildcard or optional)
				name = strings.TrimRight(name, "*?")

				template.WriteString(path[end+1 : start])
				template.WriteByte('{')
				template.WriteString(name)
				template.WriteByte('}')
				end = i
				start = -1
			}
		}
	}

	if end < 0 {
		return path
	}
	if end < len(path)-1 {
		template.WriteString(path[end+1:])
	}

	return template.String()
}
