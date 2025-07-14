package nexora

import (
	"fmt"
)

// HttpError represents an HTTP error with a status code and message.
type HttpError struct {
	StatusCode int    // HTTP status code (e.g. 404, 500)
	Message    string // Human-readable message
}

// NewHttpError creates a new HttpError.
func NewHttpError(statusCode int, message string) *HttpError {
	return &HttpError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// Error implements the error interface.
func (e *HttpError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

var (
	ErrBadRequest                    = NewHttpError(StatusBadRequest, "Bad Request")
	ErrUnauthorized                  = NewHttpError(StatusUnauthorized, "Unauthorized")
	ErrPaymentRequired               = NewHttpError(StatusPaymentRequired, "Payment Required")
	ErrForbidden                     = NewHttpError(StatusForbidden, "Forbidden")
	ErrNotFound                      = NewHttpError(StatusNotFound, "Not Found")
	ErrMethodNotAllowed              = NewHttpError(StatusMethodNotAllowed, "Method Not Allowed")
	ErrNotAcceptable                 = NewHttpError(StatusNotAcceptable, "Not Acceptable")
	ErrProxyAuthRequired             = NewHttpError(StatusProxyAuthRequired, "Proxy Authentication Required")
	ErrRequestTimeout                = NewHttpError(StatusRequestTimeout, "Request Timeout")
	ErrConflict                      = NewHttpError(StatusConflict, "Conflict")
	ErrGone                          = NewHttpError(StatusGone, "Gone")
	ErrLengthRequired                = NewHttpError(StatusLengthRequired, "Length Required")
	ErrPreconditionFailed            = NewHttpError(StatusPreconditionFailed, "Precondition Failed")
	ErrRequestEntityTooLarge         = NewHttpError(StatusRequestEntityTooLarge, "Payload Too Large")
	ErrRequestURITooLong             = NewHttpError(StatusRequestURITooLong, "URI Too Long")
	ErrUnsupportedMediaType          = NewHttpError(StatusUnsupportedMediaType, "Unsupported Media Type")
	ErrRequestedRangeNotSatisfiable  = NewHttpError(StatusRequestedRangeNotSatisfiable, "Range Not Satisfiable")
	ErrExpectationFailed             = NewHttpError(StatusExpectationFailed, "Expectation Failed")
	ErrTeapot                        = NewHttpError(StatusTeapot, "I'm a teapot")
	ErrMisdirectedRequest            = NewHttpError(StatusMisdirectedRequest, "Misdirected Request")
	ErrUnprocessableEntity           = NewHttpError(StatusUnprocessableEntity, "Unprocessable Entity")
	ErrLocked                        = NewHttpError(StatusLocked, "Locked")
	ErrFailedDependency              = NewHttpError(StatusFailedDependency, "Failed Dependency")
	ErrTooEarly                      = NewHttpError(StatusTooEarly, "Too Early")
	ErrUpgradeRequired               = NewHttpError(StatusUpgradeRequired, "Upgrade Required")
	ErrPreconditionRequired          = NewHttpError(StatusPreconditionRequired, "Precondition Required")
	ErrTooManyRequests               = NewHttpError(StatusTooManyRequests, "Too Many Requests")
	ErrRequestHeaderFieldsTooLarge   = NewHttpError(StatusRequestHeaderFieldsTooLarge, "Request Header Fields Too Large")
	ErrUnavailableForLegalReasons    = NewHttpError(StatusUnavailableForLegalReasons, "Unavailable For Legal Reasons")
	ErrInternalServerError           = NewHttpError(StatusInternalServerError, "Internal Server Error")
	ErrNotImplemented                = NewHttpError(StatusNotImplemented, "Not Implemented")
	ErrBadGateway                    = NewHttpError(StatusBadGateway, "Bad Gateway")
	ErrServiceUnavailable            = NewHttpError(StatusServiceUnavailable, "Service Unavailable")
	ErrGatewayTimeout                = NewHttpError(StatusGatewayTimeout, "Gateway Timeout")
	ErrHTTPVersionNotSupported       = NewHttpError(StatusHTTPVersionNotSupported, "HTTP Version Not Supported")
	ErrVariantAlsoNegotiates         = NewHttpError(StatusVariantAlsoNegotiates, "Variant Also Negotiates")
	ErrInsufficientStorage           = NewHttpError(StatusInsufficientStorage, "Insufficient Storage")
	ErrLoopDetected                  = NewHttpError(StatusLoopDetected, "Loop Detected")
	ErrNotExtended                   = NewHttpError(StatusNotExtended, "Not Extended")
	ErrNetworkAuthenticationRequired = NewHttpError(StatusNetworkAuthenticationRequired, "Network Authentication Required")
)
