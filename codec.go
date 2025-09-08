package nexora

import "io"

// EncoderFunc is a function that encodes a Go value into a byte slice.
// It is used for encoding JSON and XML payloads.
type EncoderFunc func(v any) ([]byte, error)

// DecoderFunc is a function that decodes a byte slice into a Go value.
// It is used for decoding JSON and XML payloads.
type DecoderFunc func(r io.Reader, v any) error

// IndentationEncoder is a function that encodes a Go value into an indented byte slice.
// It is used for encoding indented JSON and XML payloads.
type IndentationEncoder func(v any, prefix, indent string) ([]byte, error)
