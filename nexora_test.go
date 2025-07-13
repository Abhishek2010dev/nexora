package nexora

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Dummy handler to use in tests
func dummyHandler(text string) Handler {
	return func(c *Context) error {
		_, _ = c.ResponseWriter().Write([]byte(text))
		return nil
	}
}

func TestHandleAndServeHTTP(t *testing.T) {
	router := New()

	// Register route
	router.Handle(http.MethodGet, "/hello", dummyHandler("Hello, Nexora!"))

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
	if string(body) != "Hello, Nexora!" {
		t.Fatalf("Expected body 'Hello, Nexora!', got '%s'", string(body))
	}
}

func TestNotFound(t *testing.T) {
	router := New()

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 Not Found, got %d", w.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	router := New()

	router.Handle(http.MethodGet, "/onlyget", dummyHandler("GET only"))

	req := httptest.NewRequest(http.MethodPost, "/onlyget", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Expected 405 Method Not Allowed, got %d", w.Code)
	}

	allow := w.Header().Get("Allow")
	if allow != "GET, OPTIONS" {
		t.Fatalf("Expected Allow header to include GET and OPTIONS, got %q", allow)
	}
}

func TestOPTIONS(t *testing.T) {
	router := New()

	router.Handle(http.MethodGet, "/opt", dummyHandler("GET ok"))
	router.Handle(http.MethodPost, "/opt", dummyHandler("POST ok"))

	req := httptest.NewRequest(http.MethodOptions, "/opt", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for OPTIONS, got %d", w.Code)
	}

	allow := strings.Split(w.Header().Get("Allow"), ", ")
	expected := map[string]bool{"GET": true, "POST": true, "OPTIONS": true}

	for _, method := range allow {
		if !expected[method] {
			t.Fatalf("Unexpected method %q in Allow header", method)
		}
		delete(expected, method)
	}

	for missing := range expected {
		t.Fatalf("Missing method %q in Allow header", missing)
	}
}

func TestRedirectTrailingSlash(t *testing.T) {
	router := New()

	router.Handle(http.MethodGet, "/slash", dummyHandler("No slash"))

	req := httptest.NewRequest(http.MethodGet, "/slash/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMovedPermanently && w.Code != http.StatusPermanentRedirect {
		t.Fatalf("Expected redirect for trailing slash, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/slash" {
		t.Fatalf("Expected redirect location '/slash', got %q", location)
	}
}

func TestOptionalParams(t *testing.T) {
	router := New()

	// Register route with optional parameters
	router.Handle(http.MethodGet, "/user/{name?}", func(c *Context) error {
		name := c.Param("name")
		if name == "" {
			name = "Guest"
		}
		_, _ = c.ResponseWriter().Write([]byte("Hello, " + name))
		return nil
	})

	// Case 1: With param
	reqWith := httptest.NewRequest(http.MethodGet, "/user/Abhishek", nil)
	wWith := httptest.NewRecorder()
	router.ServeHTTP(wWith, reqWith)

	bodyWith, _ := io.ReadAll(wWith.Body)
	if string(bodyWith) != "Hello, Abhishek" {
		t.Fatalf("Expected 'Hello, Abhishek', got '%s'", string(bodyWith))
	}

	// Case 2: Without param (optional)
	reqWithout := httptest.NewRequest(http.MethodGet, "/user", nil)
	wWithout := httptest.NewRecorder()
	router.ServeHTTP(wWithout, reqWithout)

	bodyWithout, _ := io.ReadAll(wWithout.Body)
	if string(bodyWithout) != "Hello, Guest" {
		t.Fatalf("Expected 'Hello, Guest', got '%s'", string(bodyWithout))
	}
}
