package nexora

import (
	"net/url"

	"github.com/go-playground/form/v4"
)

// Binder holds separate decoders for query and form data.
type Binder struct {
	queryDecoder *form.Decoder
	formDecoder  *form.Decoder
}

// newBinder creates a new Binder with decoders for "query" and "form" tags.
func newBinder() *Binder {
	queryDecoder := form.NewDecoder()
	queryDecoder.SetTagName("query")

	formDecoder := form.NewDecoder()
	formDecoder.SetTagName("form")

	return &Binder{
		queryDecoder: queryDecoder,
		formDecoder:  formDecoder,
	}
}

// DecodeQuery decodes URL query parameters into the provided struct.
func (b *Binder) DecodeQuery(values url.Values, v any) error {
	return b.queryDecoder.Decode(v, url.Values(values))
}

// DecodeForm decodes form data into the provided struct.
func (b *Binder) DecodeForm(values url.Values, v any) error {
	return b.formDecoder.Decode(v, url.Values(values))
}

/*
	Usage Example:

	// Define a struct with both query and form tags
	type CreateUserRequest struct {
		Name  string `query:"name" form:"name"`
		Email string `query:"email" form:"email"`
	}

	// In your handler:
	func CreateUser(c *nexora.Context) error {
		var req CreateUserRequest

		// Bind query parameters
		if err := c.BindQuery(&req); err != nil {
			return nexora.NewHTTPError(http.StatusBadRequest, "invalid query parameters")
		}

		// Bind form data (if any)
		if err := c.BindForm(&req); err != nil {
			return nexora.NewHTTPError(http.StatusBadRequest, "invalid form data")
		}

		// Now req is populated from either query or form data
		// ...
	}
*/
