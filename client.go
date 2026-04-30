package netloc8

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the production API endpoint.
	DefaultBaseURL = "https://api.netloc8.com"

	// DefaultTimeout is the default request timeout.
	DefaultTimeout = 10 * time.Second

	headerAPIKey    = "X-API-Key"
	headerOrigin    = "Origin"
	headerAccept    = "Accept"
	headerUserAgent = "User-Agent"
	headerClient    = "X-NetLoc8-Client"

	// Response headers exposed by the API.
	HeaderRateLimitLimit     = "X-RateLimit-Limit"
	HeaderRateLimitRemaining = "X-RateLimit-Remaining"
	HeaderPrecision          = "X-NetLoc8-Precision"
	HeaderCached             = "X-NetLoc8-Cached"
	HeaderProvider           = "X-NetLoc8-Provider"
)

// Client is a NetLoc8 API client. Create one with NewClient and reuse
// it across goroutines — Client is safe for concurrent use.
type Client struct {
	apiKey    string
	baseURL   string
	origin    string
	userAgent string
	client    *http.Client
}

// Option configures a Client.
type Option func( *Client )

// WithBaseURL overrides the default API base URL.
// Useful for testing or self-hosted deployments.
func WithBaseURL( url string ) Option {
	return func( c *Client ) {
		c.baseURL = strings.TrimRight( url, "/" )
	}
}

// WithHTTPClient sets a custom http.Client for all API requests.
// Use this to configure proxy transports, custom TLS, or connection pooling.
//
// Example:
//
//	transport := &http.Transport{
//	    Proxy: http.ProxyURL( proxyURL ),
//	}
//	client := netloc8.NewClient( "pk_key",
//	    netloc8.WithHTTPClient( &http.Client{ Transport: transport } ),
//	)
func WithHTTPClient( hc *http.Client ) Option {
	return func( c *Client ) {
		c.client = hc
	}
}

// WithTimeout sets the request timeout for API calls.
// Overrides the timeout on the default http.Client. If a custom
// http.Client is provided via WithHTTPClient, this option sets
// its Timeout field.
func WithTimeout( d time.Duration ) Option {
	return func( c *Client ) {
		c.client.Timeout = d
	}
}

// WithOrigin sets the Origin header for publishable key (pk_) authentication.
// Required when using a publishable key; the API validates that the
// Origin matches an allowed origin for the key.
func WithOrigin( origin string ) Option {
	return func( c *Client ) {
		c.origin = origin
	}
}

// WithUserAgent sets a custom User-Agent string. The SDK version is
// always appended (e.g. "my-app/1.0 netloc8-go/0.1.0").
func WithUserAgent( ua string ) Option {
	return func( c *Client ) {
		c.userAgent = ua
	}
}

// NewClient creates a NetLoc8 API client.
//
// apiKey is your publishable (pk_) or secret (sk_) API key.
// For publishable keys, you must also set WithOrigin.
//
// The returned Client is safe for concurrent use and should be reused.
func NewClient( apiKey string, opts ...Option ) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: DefaultBaseURL,
		client: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	for _, opt := range opts {
		opt( c )
	}

	return c
}

// ──────────────────────────────────────────────────────
// Core API methods
// ──────────────────────────────────────────────────────

// LookupIP returns full geolocation data for the given IP address.
//
// The IP must be a valid IPv4 or IPv6 address. Use NormalizeIP to
// clean user input before passing it here.
func ( c *Client ) LookupIP( ctx context.Context, ip string ) ( *Geo, error ) {
	path := "/v1/ip/" + url.PathEscape( ip )
	geo := &Geo{}
	if err := c.doJSON( ctx, path, geo ); err != nil {
		return nil, err
	}
	return geo, nil
}

// LookupMe returns full geolocation data for the caller's own IP.
//
// The API auto-detects the IP from the connection. This is useful for:
//   - Browser-side geolocation (with a publishable key)
//   - Proxy exit-IP discovery (when using a custom http.Client with
//     a proxy transport)
func ( c *Client ) LookupMe( ctx context.Context ) ( *Geo, error ) {
	geo := &Geo{}
	if err := c.doJSON( ctx, "/v1/ip/me", geo ); err != nil {
		return nil, err
	}
	return geo, nil
}

// Timezone returns the IANA timezone string for the given IP address.
//
// This is a lightweight alternative to LookupIP when you only need
// the timezone (e.g. for session alignment).
func ( c *Client ) Timezone( ctx context.Context, ip string ) ( string, error ) {
	path := "/v1/ip/" + url.PathEscape( ip ) + "/timezone"
	var tz string
	if err := c.doJSON( ctx, path, &tz ); err != nil {
		return "", err
	}
	return tz, nil
}

// MyTimezone returns the IANA timezone string for the caller's own IP.
func ( c *Client ) MyTimezone( ctx context.Context ) ( string, error ) {
	var tz string
	if err := c.doJSON( ctx, "/v1/ip/me/timezone", &tz ); err != nil {
		return "", err
	}
	return tz, nil
}

// Validate checks whether the given string is a valid IP address.
// Returns the API's validation result.
//
// For local-only validation without an API call, use ParseIP instead.
func ( c *Client ) Validate( ctx context.Context, ip string ) ( bool, error ) {
	path := "/v1/ip/" + url.PathEscape( ip ) + "/validation"
	var valid bool
	if err := c.doJSON( ctx, path, &valid ); err != nil {
		return false, err
	}
	return valid, nil
}

// ──────────────────────────────────────────────────────
// HTTP transport
// ──────────────────────────────────────────────────────

// doJSON performs a GET request and decodes the JSON response into dst.
// Returns *APIError for non-2xx responses.
func ( c *Client ) doJSON( ctx context.Context, path string, dst any ) error {
	reqURL := c.baseURL + path

	req, err := http.NewRequestWithContext( ctx, http.MethodGet, reqURL, nil )
	if err != nil {
		return fmt.Errorf( "netloc8: create request: %w", err )
	}

	c.setHeaders( req )

	resp, err := c.client.Do( req )
	if err != nil {
		return fmt.Errorf( "netloc8: request failed: %w", err )
	}
	defer resp.Body.Close()

	body, err := io.ReadAll( resp.Body )
	if err != nil {
		return fmt.Errorf( "netloc8: read response: %w", err )
	}

	// Non-2xx → parse structured API error.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.parseError( resp.StatusCode, body )
	}

	if err := json.Unmarshal( body, dst ); err != nil {
		return fmt.Errorf( "netloc8: decode response: %w", err )
	}

	return nil
}

// setHeaders adds authentication and metadata headers to a request.
func ( c *Client ) setHeaders( req *http.Request ) {
	req.Header.Set( headerAccept, "application/json" )

	if c.apiKey != "" {
		req.Header.Set( headerAPIKey, c.apiKey )
	}

	if c.origin != "" {
		req.Header.Set( headerOrigin, c.origin )
	}

	// User-Agent: custom prefix + SDK identifier.
	ua := fmt.Sprintf( "netloc8-go/%s", Version )
	if c.userAgent != "" {
		ua = c.userAgent + " " + ua
	}
	req.Header.Set( headerUserAgent, ua )

	// Client identifier header (matches JS SDK's X-NetLoc8-Client).
	req.Header.Set( headerClient, fmt.Sprintf( "netloc8-go/%s", Version ) )
}

// parseError extracts a structured APIError from a non-2xx response body.
func ( c *Client ) parseError( status int, body []byte ) *APIError {
	apiErr := &APIError{
		Status: status,
	}

	var raw apiErrorResponse
	if err := json.Unmarshal( body, &raw ); err == nil {
		if raw.Error != nil {
			apiErr.Code = raw.Error.Code
			apiErr.Message = raw.Error.Message
		}
		if raw.Meta != nil {
			apiErr.RequestID = raw.Meta.RequestID
		}
	}

	// Fallback message if the body wasn't parseable.
	if apiErr.Message == "" {
		apiErr.Message = http.StatusText( status )
	}

	return apiErr
}
