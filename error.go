package nexora

import (
	"fmt"
)

// HTTPError represents an HTTP error with a status code and message.
type HTTPError struct {
	StatusCode int    // HTTP status code (e.g. 404, 500)
	Message    string // Human-readable message
}

// NewHTTPError creates a new HTTPError.
func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

var (
	ErrBadRequest                    = NewHTTPError(StatusBadRequest, "Bad Request")
	ErrUnauthorized                  = NewHTTPError(StatusUnauthorized, "Unauthorized")
	ErrPaymentRequired               = NewHTTPError(StatusPaymentRequired, "Payment Required")
	ErrForbidden                     = NewHTTPError(StatusForbidden, "Forbidden")
	ErrNotFound                      = NewHTTPError(StatusNotFound, "Not Found")
	ErrMethodNotAllowed              = NewHTTPError(StatusMethodNotAllowed, "Method Not Allowed")
	ErrNotAcceptable                 = NewHTTPError(StatusNotAcceptable, "Not Acceptable")
	ErrProxyAuthRequired             = NewHTTPError(StatusProxyAuthRequired, "Proxy Authentication Required")
	ErrRequestTimeout                = NewHTTPError(StatusRequestTimeout, "Request Timeout")
	ErrConflict                      = NewHTTPError(StatusConflict, "Conflict")
	ErrGone                          = NewHTTPError(StatusGone, "Gone")
	ErrLengthRequired                = NewHTTPError(StatusLengthRequired, "Length Required")
	ErrPreconditionFailed            = NewHTTPError(StatusPreconditionFailed, "Precondition Failed")
	ErrRequestEntityTooLarge         = NewHTTPError(StatusRequestEntityTooLarge, "Payload Too Large")
	ErrRequestURITooLong             = NewHTTPError(StatusRequestURITooLong, "URI Too Long")
	ErrUnsupportedMediaType          = NewHTTPError(StatusUnsupportedMediaType, "Unsupported Media Type")
	ErrRequestedRangeNotSatisfiable  = NewHTTPError(StatusRequestedRangeNotSatisfiable, "Range Not Satisfiable")
	ErrExpectationFailed             = NewHTTPError(StatusExpectationFailed, "Expectation Failed")
	ErrTeapot                        = NewHTTPError(StatusTeapot, "I'm a teapot")
	ErrMisdirectedRequest            = NewHTTPError(StatusMisdirectedRequest, "Misdirected Request")
	ErrUnprocessableEntity           = NewHTTPError(StatusUnprocessableEntity, "Unprocessable Entity")
	ErrLocked                        = NewHTTPError(StatusLocked, "Locked")
	ErrFailedDependency              = NewHTTPError(StatusFailedDependency, "Failed Dependency")
	ErrTooEarly                      = NewHTTPError(StatusTooEarly, "Too Early")
	ErrUpgradeRequired               = NewHTTPError(StatusUpgradeRequired, "Upgrade Required")
	ErrPreconditionRequired          = NewHTTPError(StatusPreconditionRequired, "Precondition Required")
	ErrTooManyRequests               = NewHTTPError(StatusTooManyRequests, "Too Many Requests")
	ErrRequestHeaderFieldsTooLarge   = NewHTTPError(StatusRequestHeaderFieldsTooLarge, "Request Header Fields Too Large")
	ErrUnavailableForLegalReasons    = NewHTTPError(StatusUnavailableForLegalReasons, "Unavailable For Legal Reasons")
	ErrInternalServerError           = NewHTTPError(StatusInternalServerError, "Internal Server Error")
	ErrNotImplemented                = NewHTTPError(StatusNotImplemented, "Not Implemented")
	ErrBadGateway                    = NewHTTPError(StatusBadGateway, "Bad Gateway")
	ErrServiceUnavailable            = NewHTTPError(StatusServiceUnavailable, "Service Unavailable")
	ErrGatewayTimeout                = NewHTTPError(StatusGatewayTimeout, "Gateway Timeout")
	ErrHTTPVersionNotSupported       = NewHTTPError(StatusHTTPVersionNotSupported, "HTTP Version Not Supported")
	ErrVariantAlsoNegotiates         = NewHTTPError(StatusVariantAlsoNegotiates, "Variant Also Negotiates")
	ErrInsufficientStorage           = NewHTTPError(StatusInsufficientStorage, "Insufficient Storage")
	ErrLoopDetected                  = NewHTTPError(StatusLoopDetected, "Loop Detected")
	ErrNotExtended                   = NewHTTPError(StatusNotExtended, "Not Extended")
	ErrNetworkAuthenticationRequired = NewHTTPError(StatusNetworkAuthenticationRequired, "Network Authentication Required")
)
