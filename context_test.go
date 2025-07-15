package nexora

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContext_SendString(t *testing.T) {
	// Create a test HTTP request and response recorder
	req := httptest.NewRequest("GET", "/hello", nil)
	rec := httptest.NewRecorder()

	// Create and initialize a Context
	ctx := newContext(nil)
	ctx.init(req, rec)

	// Call SendString
	err := ctx.SendString("Hello, Nexora!")
	if err != nil {
		t.Fatalf("SendString failed: %v", err)
	}

	// Check response
	result := rec.Body.String()
	if result != "Hello, Nexora!" {
		t.Errorf("unexpected response body: got %q, want %q", result, "Hello, Nexora!")
	}
}

func TestContext_Status(t *testing.T) {
	req := httptest.NewRequest("GET", "/status", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	ctx.Status(http.StatusTeapot).SendString("I'm a teapot")

	if rec.Code != http.StatusTeapot {
		t.Errorf("unexpected status code: got %d, want %d", rec.Code, http.StatusTeapot)
	}

	if body := rec.Body.String(); body != "I'm a teapot" {
		t.Errorf("unexpected body: got %q, want %q", body, "I'm a teapot")
	}
}

func TestContext_Next(t *testing.T) {
	req := httptest.NewRequest("GET", "/next", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	var called []string
	ctx.handlers = []Handler{
		func(c *Context) error {
			called = append(called, "1")
			return c.Next()
		},
		func(c *Context) error {
			called = append(called, "2")
			return nil
		},
	}

	err := ctx.Next()
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}

	want := []string{"1", "2"}
	if strings.Join(called, ",") != strings.Join(want, ",") {
		t.Errorf("handlers called in wrong order: got %v, want %v", called, want)
	}
}

func TestContext_Param(t *testing.T) {
	req := httptest.NewRequest("GET", "/users/42", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	// Simulate route parameters
	ctx.params = map[string]string{
		"id":   "42",
		"name": "",
	}

	// Test existing param
	id := ctx.Param("id")
	if id != "42" {
		t.Errorf("Param id = %q; want %q", id, "42")
	}

	// Test missing param with default
	role := ctx.Param("role", "admin")
	if role != "admin" {
		t.Errorf("Param role with default = %q; want %q", role, "admin")
	}

	// Test missing param without default
	role = ctx.Param("role")
	if role != "" {
		t.Errorf("Param role without default = %q; want empty string", role)
	}

	// Test empty param (should not use default)
	name := ctx.Param("name", "guest")
	if name != "" {
		t.Errorf("Param name = %q; want empty string", name)
	}
}

func TestContext_ParamExists(t *testing.T) {
	req := httptest.NewRequest("GET", "/items/5", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	ctx.params = map[string]string{
		"item": "5",
	}

	val, ok := ctx.ParamExists("item")
	if !ok || val != "5" {
		t.Errorf("ParamExists(item) = (%q, %v); want (%q, true)", val, ok, "5")
	}

	val, ok = ctx.ParamExists("missing")
	if ok || val != "" {
		t.Errorf("ParamExists(missing) = (%q, %v); want (\"\", false)", val, ok)
	}
}

func TestContext_Queries(t *testing.T) {
	req := httptest.NewRequest("GET", "/search?q=golang&tag=web&tag=fast&empty=", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	values := ctx.Queries()
	if values.Get("q") != "golang" {
		t.Errorf("Queries()[q] = %q; want %q", values.Get("q"), "golang")
	}
	if got := values["tag"]; len(got) != 2 || got[0] != "web" || got[1] != "fast" {
		t.Errorf("Queries()[tag] = %v; want [web fast]", got)
	}
	if _, ok := values["empty"]; !ok {
		t.Errorf("Queries()[empty] missing; want present")
	}
}

func TestContext_QueryArray(t *testing.T) {
	req := httptest.NewRequest("GET", "/search?tag=web&tag=fast", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	arr := ctx.QueryArray("tag")
	if len(arr) != 2 || arr[0] != "web" || arr[1] != "fast" {
		t.Errorf("QueryArray(tag) = %v; want [web fast]", arr)
	}

	arr = ctx.QueryArray("missing")
	if arr != nil && len(arr) != 0 {
		t.Errorf("QueryArray(missing) = %v; want nil or []", arr)
	}
}

func TestContext_Query(t *testing.T) {
	req := httptest.NewRequest("GET", "/search?q=golang", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	// existing key
	val := ctx.Query("q")
	if val != "golang" {
		t.Errorf("Query(q) = %q; want %q", val, "golang")
	}

	// missing key with default
	val = ctx.Query("page", "1")
	if val != "1" {
		t.Errorf("Query(page,1) = %q; want %q", val, "1")
	}

	// missing key without default
	val = ctx.Query("missing")
	if val != "" {
		t.Errorf("Query(missing) = %q; want \"\"", val)
	}
}

func TestContext_QueryExists(t *testing.T) {
	req := httptest.NewRequest("GET", "/search?q=golang&empty=", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	// key exists with value
	val, ok := ctx.QueryExists("q")
	if !ok || val != "golang" {
		t.Errorf("QueryExists(q) = (%q, %v); want (%q, true)", val, ok, "golang")
	}

	// key exists with empty value
	val, ok = ctx.QueryExists("empty")
	if !ok || val != "" {
		t.Errorf("QueryExists(empty) = (%q, %v); want (\"\", true)", val, ok)
	}

	// key does not exist
	val, ok = ctx.QueryExists("missing")
	if ok || val != "" {
		t.Errorf("QueryExists(missing) = (%q, %v); want (\"\", false)", val, ok)
	}
}
