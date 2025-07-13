package nexora

import (
	"reflect"
	"testing"
)

func TestRouteGroup_Use(t *testing.T) {
	n := &Nexora{}
	g := newRouteGroup(n, "/api", nil)

	h1 := dummyHandler("one")
	h2 := dummyHandler("two")

	g.Use(h1, h2)

	if len(g.handlers) != 2 {
		t.Fatalf("expected 2 handlers, got %d", len(g.handlers))
	}

	if reflect.ValueOf(g.handlers[0]).Pointer() != reflect.ValueOf(h1).Pointer() ||
		reflect.ValueOf(g.handlers[1]).Pointer() != reflect.ValueOf(h2).Pointer() {
		t.Error("handlers not stored in order or incorrectly")
	}
}

func TestRouteGroup_Group_InheritsHandlers(t *testing.T) {
	n := &Nexora{}
	parent := newRouteGroup(n, "/api", []Handler{dummyHandler("parent")})

	child := parent.Group("/v1")

	if child.prefix != "/api/v1" {
		t.Errorf("expected prefix '/api/v1', got '%s'", child.prefix)
	}

	if len(child.handlers) != len(parent.handlers) {
		t.Errorf("expected child to inherit %d handlers, got %d", len(parent.handlers), len(child.handlers))
	}

	if reflect.ValueOf(child.handlers[0]).Pointer() != reflect.ValueOf(parent.handlers[0]).Pointer() {
		t.Error("handler not correctly inherited")
	}
}

func TestRouteGroup_Group_CustomHandlers(t *testing.T) {
	n := &Nexora{}
	parent := newRouteGroup(n, "/api", []Handler{dummyHandler("parent")})

	custom := dummyHandler("custom")
	child := parent.Group("/v2", custom)

	if child.prefix != "/api/v2" {
		t.Errorf("expected prefix '/api/v2', got '%s'", child.prefix)
	}

	if len(child.handlers) != 1 {
		t.Errorf("expected 1 custom handler, got %d", len(child.handlers))
	}

	if reflect.ValueOf(child.handlers[0]).Pointer() != reflect.ValueOf(custom).Pointer() {
		t.Error("custom handler not set correctly")
	}
}

func TestCombineHandlers(t *testing.T) {
	h1 := []Handler{
		dummyHandler("h1-1"),
		dummyHandler("h1-2"),
	}
	h2 := []Handler{
		dummyHandler("h2-1"),
		dummyHandler("h2-2"),
	}

	combined := combineHandlers(h1, h2)

	if len(combined) != 4 {
		t.Fatalf("expected 4 handlers, got %d", len(combined))
	}
}

func TestBuildURLTemplate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/user/{id}", "/user/{id}"},
		{"/user/{id:[0-9]+}", "/user/{id}"},
		{"/{category}/{id:[0-9]+}/view", "/{category}/{id}/view"},
		{"/download/{file*}", "/download/{file}"},
		{"/static/*", "/static/"},
		{"/", "/"},
		{"/test/{slug}-{id:[0-9]+}", "/test/{slug}-{id}"},
	}

	for _, tt := range tests {
		got := buildURLTemplate(tt.input)
		if got != tt.expected {
			t.Errorf("buildURLTemplate(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}
