package nexora

// EncoderFunc defines how a value is converted into bytes (e.g., JSON, XML, YAML).
type EncoderFunc func(v any) ([]byte, error)

// DecoderFunc defines how raw bytes are converted back into a Go value.
type DecoderFunc func([]byte) (any, error)
