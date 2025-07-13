package nexora

import (
	"fmt"
	"net/url"
	"strings"
)

// Route represents a single route in the routing tree.
type Route struct {
	group          *RouteGroup // The group this route belongs to, which contains shared settings and handlers.
	method, path   string      // The HTTP method (GET, POST, etc.) and the path for this route.
	name, template string      // The name of the route and a template for generating URLs.
	tags           []any       // Custom data associated with the route, which can be used for various purposes.
	routes         []*Route    // Nested routes, which can be used to create more complex routing structures.
}

// Name sets the name of the route.
// It also registers the route in the namedRoutes map of the Nexora instance.
func (r *Route) Name(name string) *Route {
	r.name = name
	r.group.nexora.namedRoutes[name] = r
	return r
}

// Tag associates some custom data with the route.
func (r *Route) Tag(value any) *Route {
	if len(r.routes) > 0 {
		for _, route := range r.routes {
			// If the route has a template, we apply the tags to it.
			route.Tag(value)
		}
		return r
	}
	if r.tags == nil {
		r.tags = []any{}
	}
	r.tags = append(r.tags, value)
	return r
}

// Method returns the HTTP method that this route is associated with.
func (r *Route) Method() string {
	return r.method
}

// Path returns the request path that this route should match.
func (r *Route) Path() string {
	return r.group.prefix + r.path
}

// Tags returns all custom data associated with the route.
func (r *Route) Tags() []any {
	return r.tags
}

// Get adds the route to the router using the GET HTTP method.
func (r *Route) Get(handlers ...Handler) *Route {
	return r.group.add(MethodGet, r.path, handlers)
}

// Post adds the route to the router using the POST HTTP method.
func (r *Route) Post(handlers ...Handler) *Route {
	return r.group.add(MethodPost, r.path, handlers)
}

// Put adds the route to the router using the PUT HTTP method.
func (r *Route) Put(handlers ...Handler) *Route {
	return r.group.add(MethodPut, r.path, handlers)
}

// Patch adds the route to the router using the PATCH HTTP method.
func (r *Route) Patch(handlers ...Handler) *Route {
	return r.group.add(MethodPatch, r.path, handlers)
}

// Delete adds the route to the router using the DELETE HTTP method.
func (r *Route) Delete(handlers ...Handler) *Route {
	return r.group.add(MethodDelete, r.path, handlers)
}

// Connect adds the route to the router using the CONNECT HTTP method.
func (r *Route) Connect(handlers ...Handler) *Route {
	return r.group.add(MethodConnect, r.path, handlers)
}

// Head adds the route to the router using the HEAD HTTP method.
func (r *Route) Head(handlers ...Handler) *Route {
	return r.group.add(MethodHead, r.path, handlers)
}

// Options adds the route to the router using the OPTIONS HTTP method.
func (r *Route) Options(handlers ...Handler) *Route {
	return r.group.add(MethodOptions, r.path, handlers)
}

// Trace adds the route to the router using the TRACE HTTP method.
func (r *Route) Trace(handlers ...Handler) *Route {
	return r.group.add(MethodTrace, r.path, handlers)
}

// Handle adds the route to the router using the specified HTTP method.
func (r *Route) Handle(method string, handlers ...Handler) *Route {
	return r.group.add(method, r.path, handlers)
}

// URL creates a URL using the current route and the given parameters.
// The parameters should be given in the sequence of name1, value1, name2, value2, and so on.
// If a parameter in the route is not provided a value, the parameter token will remain in the resulting URL.
// The method will perform URL encoding for all given parameter values.
func (r *Route) URL(pairs ...any) (s string) {
	s = r.template
	for i := range pairs {
		name := fmt.Sprintf("{%v}", pairs[i])
		value := ""
		if i < len(pairs)-1 {
			value = url.QueryEscape(fmt.Sprint(pairs[i+1]))
		}
		s = strings.ReplaceAll(s, name, value)
	}
	return
}

// String returns the string representation of the route.
func (r *Route) String() string {
	return r.method + " " + r.group.prefix + r.path
}
