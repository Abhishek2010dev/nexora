package nexora

// EncoderFunc defines a function type responsible for encoding a Go value
// into a slice of bytes. Implementations of this function may produce
// serialized representations such as JSON, XML, or YAML.
//
// The function takes one parameter:
//
//   - v: The Go value to be encoded. This can be any type supported
//     by the chosen encoding format.
//
// The function returns:
//
//   - []byte: The encoded representation of the value.
//   - error:  An error if the encoding fails, otherwise nil.
//
// Example:
//
//	var encode EncoderFunc = json.Marshal
//	data, err := encode(map[string]any{"foo": "bar"})
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(string(data))
type EncoderFunc func(v any) ([]byte, error)

// DecoderFunc defines a function type responsible for decoding a slice
// of bytes into a Go value. Implementations of this function are expected
// to populate the provided pointer with the decoded representation.
//
// The function takes two parameters:
//
//   - data: A slice of bytes representing the encoded value (e.g., JSON, XML, YAML).
//   - v:    A pointer to the Go value where the decoded data will be stored.
//
// The function returns:
//
//   - error: An error if the decoding fails, otherwise nil.
//
// Example:
//
//	var decode DecoderFunc = json.Unmarshal
//	var result map[string]any
//	err := decode([]byte(`{"foo":"bar"}`), &result)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(result["foo"]) // Output: bar
type DecoderFunc func(data []byte, v any) error

// IndentationEncoder defines a function type for encoding a Go value
// into a slice of bytes, while also applying custom indentation rules.
//
// This is commonly used for producing human-readable, pretty-printed
// JSON or similar structured formats.
//
// The function takes three parameters:
//
//   - v:      The Go value to encode.
//   - prefix: A string to prepend to each line of output (e.g., for alignment).
//   - indent: A string used to represent one level of indentation
//     (e.g., spaces or tabs).
//
// The function returns:
//
//   - []byte: The encoded, indented representation of the value.
//   - error:  An error if the encoding fails, otherwise nil.
//
// Example:
//
//	var encode IndentationEncoder = json.MarshalIndent
//	data := map[string]any{"foo": "bar"}
//	pretty, err := encode(data, "", "  ")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(string(pretty))
//
// Output:
//
//	{
//	  "foo": "bar"
//	}
type IndentationEncoder func(v any, prefix, indent string) ([]byte, error)

