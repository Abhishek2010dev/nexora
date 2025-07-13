package nexora

import (
	"testing"
)

func setupRoute() (*Nexora, *RouteGroup, *Route) {
	n := New()
	group := &RouteGroup{
		prefix:   "/api",
		nexora:   n,
		handlers: nil,
	}
	route := &Route{
		group:    group,
		method:   MethodGet,
		path:     "/user/{id}",
		template: "/api/user/{id}",
	}
	return n, group, route
}

func TestRoute_Name(t *testing.T) {
	n, _, route := setupRoute()
	route.Name("get_user")

	if n.namedRoutes["get_user"] != route {
		t.Errorf("route not registered correctly in namedRoutes map")
	}
	if route.name != "get_user" {
		t.Errorf("route name not set correctly: got %s", route.name)
	}
}

func TestRoute_Tag(t *testing.T) {
	_, _, route := setupRoute()

	route.Tag("auth").Tag("admin")

	expected := []any{"auth", "admin"}
	got := route.Tags()

	if len(got) != len(expected) {
		t.Fatalf("expected %d tags, got %d", len(expected), len(got))
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("expected tag %v, got %v", expected[i], got[i])
		}
	}
}

func TestRoute_Path(t *testing.T) {
	_, _, route := setupRoute()
	if route.Path() != "/api/user/{id}" {
		t.Errorf("unexpected route path: got %s", route.Path())
	}
}

func TestRoute_URL(t *testing.T) {
	_, _, route := setupRoute()
	url := route.URL("id", 42)

	expected := "/api/user/42"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

func TestRoute_String(t *testing.T) {
	_, _, route := setupRoute()
	str := route.String()
	expected := "GET /api/user/{id}"

	if str != expected {
		t.Errorf("expected route string %q, got %q", expected, str)
	}
}

func TestRoute_NestedTag(t *testing.T) {
	_, group := New(), &RouteGroup{}
	parent := &Route{
		group: group,
		routes: []*Route{
			{group: group},
			{group: group},
		},
	}
	parent.Tag("global")

	for i, r := range parent.routes {
		if len(r.tags) != 1 || r.tags[0] != "global" {
			t.Errorf("child route %d did not receive tag", i)
		}
	}
}
