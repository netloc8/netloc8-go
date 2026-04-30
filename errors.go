package netloc8

import (
	"fmt"
)

// APIError represents a structured error response from the NetLoc8 API.
// It implements the error interface and supports errors.As for typed
// error handling.
type APIError struct {
	// Status is the HTTP status code (e.g. 400, 403, 404, 429, 500).
	Status int `json:"-"`

	// Code is the machine-readable error code (e.g. "INVALID_IP",
	// "FORBIDDEN", "RATE_LIMIT_EXCEEDED").
	Code string `json:"code"`

	// Message is the human-readable error description.
	Message string `json:"message"`

	// RequestID is the unique request identifier, if returned.
	RequestID string `json:"requestId,omitempty"`
}

// Error implements the error interface.
func ( e *APIError ) Error() string {
	if e.Code != "" {
		return fmt.Sprintf( "netloc8: %s — %s (HTTP %d)", e.Code, e.Message, e.Status )
	}
	return fmt.Sprintf( "netloc8: HTTP %d", e.Status )
}

// apiErrorResponse is the raw JSON envelope returned by the API on error.
type apiErrorResponse struct {
	Query *struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"query,omitempty"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
	Meta *struct {
		RequestID string `json:"requestId"`
	} `json:"meta,omitempty"`
}

// IsNotFound reports whether the error is a 404 Not Found response.
func IsNotFound( err error ) bool {
	if apiErr, ok := err.( *APIError ); ok {
		return apiErr.Status == 404
	}
	return false
}

// IsRateLimited reports whether the error is a 429 Rate Limit Exceeded response.
func IsRateLimited( err error ) bool {
	if apiErr, ok := err.( *APIError ); ok {
		return apiErr.Status == 429
	}
	return false
}

// IsForbidden reports whether the error is a 403 Forbidden response.
// This typically means the API key lacks the required scope or the
// origin does not match.
func IsForbidden( err error ) bool {
	if apiErr, ok := err.( *APIError ); ok {
		return apiErr.Status == 403
	}
	return false
}
