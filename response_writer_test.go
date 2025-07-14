package nexora

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResponseWriter_Basic(t *testing.T) {
	rr := httptest.NewRecorder()
	w := NewResponseWriter(rr)

	// Simulate a handler writing a response
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Hello, World!"))

	if w.Status() != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Status())
	}

	expectedSize := len("Hello, World!")
	if w.Size() != expectedSize {
		t.Errorf("expected size %d, got %d", expectedSize, w.Size())
	}

	if rr.Code != http.StatusCreated {
		t.Errorf("expected recorder code %d, got %d", http.StatusCreated, rr.Code)
	}

	if body := rr.Body.String(); body != "Hello, World!" {
		t.Errorf("expected body %q, got %q", "Hello, World!", body)
	}
}

func TestResponseWriter_OverwriteStatus(t *testing.T) {
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(nil) // restore default log output after test

	rr := httptest.NewRecorder()
	w := NewResponseWriter(rr)

	w.WriteHeader(http.StatusCreated)
	w.WriteHeader(http.StatusBadRequest) // should log warning

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "status code overwritten") {
		t.Errorf("expected warning log for overwritten status code, got: %q", logOutput)
	}
}
