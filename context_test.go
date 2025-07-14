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
