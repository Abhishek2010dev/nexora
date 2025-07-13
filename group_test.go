package nexora

import (
	"testing"
)

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
		{"/files/{name}.{ext}", "/files/{name}.{ext}"},
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
