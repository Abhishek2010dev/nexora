package nexora

import (
	"log"
	"net/http"
)

// ResponseWriter is a wrapper around http.ResponseWriter that
// captures the status code and response size for logging and middleware.
type ResponseWriter struct {
	http.ResponseWriter      // underlying http.ResponseWriter
	status              int  // HTTP status code
	size                int  // number of bytes written
	wrote               bool // whether the header has been written
}

// Ensure ResponseWriter implements http.ResponseWriter.
var _ http.ResponseWriter = (*ResponseWriter)(nil)

// NewResponseWriter creates a new wrapped ResponseWriter.
// It sets the default status code to 200.
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		status:         200,
	}
}

// WriteHeader sets the HTTP status code for the response.
// If called multiple times with different codes, a warning is logged.
func (r *ResponseWriter) WriteHeader(status int) {
	if r.wrote && r.status != status {
		log.Printf("[nexora] warning: status code overwritten from %d to %d", r.status, status)
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
	r.wrote = true
}

// Write writes the response body and automatically sets the status code to 200
// if WriteHeader was not previously called.
func (r *ResponseWriter) Write(b []byte) (int, error) {
	if !r.wrote {
		r.WriteHeader(200)
	}
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

// Size returns the total number of bytes written to the response body.
func (r *ResponseWriter) Size() int {
	return r.size
}

// Status returns the HTTP status code of the response.
func (r *ResponseWriter) Status() int {
	return r.status
}
