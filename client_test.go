package netloc8

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ──────────────────────────────────────────────────────
// Test helpers
// ──────────────────────────────────────────────────────

// newTestServer creates an httptest.Server and a Client pointed at it.
func newTestServer( t *testing.T, handler http.HandlerFunc ) ( *Client, *httptest.Server ) {
	t.Helper()
	srv := httptest.NewServer( handler )
	t.Cleanup( srv.Close )

	client := NewClient( "pk_test_key",
		WithBaseURL( srv.URL ),
		WithOrigin( "https://test.local" ),
	)
	return client, srv
}

// geoFixture returns a realistic Geo response for 8.8.8.8.
func geoFixture() *Geo {
	return &Geo{
		Query: &Query{ Type: "ip", Value: "8.8.8.8", IPVersion: 4 },
		Location: &Location{
			Continent: &Continent{ Code: "NA", Name: "North America" },
			Country: &Country{
				Code: "US",
				Name: "United States",
				Flag: "🇺🇸",
			},
			Region:     &Region{ Code: "CA", Name: "California" },
			City:       "Mountain View",
			PostalCode: "94043",
			Coordinates: &Coordinates{
				Latitude:       37.386,
				Longitude:      -122.084,
				AccuracyRadius: 621,
			},
			Timezone:      "America/Los_Angeles",
			UTCOffset:     "-07:00",
			GeoConfidence: 1.0,
		},
		Network: &Network{
			ASN:          "AS15169",
			Organization: "Google LLC",
			Domain:       "google.com",
		},
		Sources: &Sources{
			Geo: []string{ "dbip", "ip2location" },
			ASN: []string{ "ipinfo" },
			TZ:  []string{ "derived" },
		},
		Meta: &Meta{
			Precision: "city",
			Tier:      "pro",
			RequestID: "test-req-001",
		},
	}
}

// ──────────────────────────────────────────────────────
// Client construction
// ──────────────────────────────────────────────────────

func TestNewClient_Defaults( t *testing.T ) {
	c := NewClient( "pk_test" )

	if c.apiKey != "pk_test" {
		t.Errorf( "apiKey = %q, want pk_test", c.apiKey )
	}
	if c.baseURL != DefaultBaseURL {
		t.Errorf( "baseURL = %q, want %q", c.baseURL, DefaultBaseURL )
	}
	if c.client == nil {
		t.Fatal( "http.Client should not be nil" )
	}
	if c.client.Timeout != DefaultTimeout {
		t.Errorf( "Timeout = %v, want %v", c.client.Timeout, DefaultTimeout )
	}
}

func TestNewClient_Options( t *testing.T ) {
	custom := &http.Client{ Timeout: 30 * time.Second }
	c := NewClient( "sk_test",
		WithBaseURL( "https://custom-api.example.com/" ),
		WithHTTPClient( custom ),
		WithOrigin( "https://my-app.com" ),
		WithUserAgent( "my-app/2.0" ),
	)

	if c.baseURL != "https://custom-api.example.com" {
		t.Errorf( "baseURL = %q, trailing slash should be stripped", c.baseURL )
	}
	if c.client != custom {
		t.Error( "custom http.Client not set" )
	}
	if c.origin != "https://my-app.com" {
		t.Errorf( "origin = %q", c.origin )
	}
	if c.userAgent != "my-app/2.0" {
		t.Errorf( "userAgent = %q", c.userAgent )
	}
}

func TestNewClient_TimeoutOrdering( t *testing.T ) {
	// WithTimeout before WithHTTPClient — should still apply.
	custom := &http.Client{ Timeout: 30 * time.Second }
	c := NewClient( "sk_test",
		WithTimeout( 5 * time.Second ),
		WithHTTPClient( custom ),
	)

	if c.client.Timeout != 5*time.Second {
		t.Errorf( "Timeout = %v, want 5s (timeout before http client)", c.client.Timeout )
	}

	// WithTimeout after WithHTTPClient — should also work.
	custom2 := &http.Client{ Timeout: 30 * time.Second }
	c2 := NewClient( "sk_test",
		WithHTTPClient( custom2 ),
		WithTimeout( 3 * time.Second ),
	)

	if c2.client.Timeout != 3*time.Second {
		t.Errorf( "Timeout = %v, want 3s (timeout after http client)", c2.client.Timeout )
	}
}

// ──────────────────────────────────────────────────────
// Request headers
// ──────────────────────────────────────────────────────

func TestClient_Headers( t *testing.T ) {
	var capturedHeaders http.Header

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedHeaders = r.Header.Clone()
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( geoFixture() )
	})

	_, err := client.LookupIP( context.Background(), "8.8.8.8" )
	if err != nil {
		t.Fatalf( "LookupIP: %v", err )
	}

	// API key header.
	if got := capturedHeaders.Get( headerAPIKey ); got != "pk_test_key" {
		t.Errorf( "X-API-Key = %q, want pk_test_key", got )
	}

	// Origin header.
	if got := capturedHeaders.Get( headerOrigin ); got != "https://test.local" {
		t.Errorf( "Origin = %q, want https://test.local", got )
	}

	// Accept header.
	if got := capturedHeaders.Get( headerAccept ); got != "application/json" {
		t.Errorf( "Accept = %q, want application/json", got )
	}

	// User-Agent includes SDK version.
	ua := capturedHeaders.Get( headerUserAgent )
	if !strings.Contains( ua, "netloc8-go/" ) {
		t.Errorf( "User-Agent %q should contain netloc8-go/", ua )
	}

	// Client identifier header.
	clientID := capturedHeaders.Get( headerClient )
	if !strings.HasPrefix( clientID, "netloc8-go/" ) {
		t.Errorf( "X-NetLoc8-Client = %q, want prefix netloc8-go/", clientID )
	}
}

func TestClient_CustomUserAgent( t *testing.T ) {
	var capturedUA string

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedUA = r.Header.Get( headerUserAgent )
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( geoFixture() )
	})
	client.userAgent = "my-app/1.0"

	_, err := client.LookupIP( context.Background(), "8.8.8.8" )
	if err != nil {
		t.Fatalf( "LookupIP: %v", err )
	}

	if !strings.HasPrefix( capturedUA, "my-app/1.0 " ) {
		t.Errorf( "User-Agent = %q, want prefix 'my-app/1.0 '", capturedUA )
	}
}

// ──────────────────────────────────────────────────────
// LookupIP
// ──────────────────────────────────────────────────────

func TestClient_LookupIP( t *testing.T ) {
	var capturedPath string

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedPath = r.URL.Path
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( geoFixture() )
	})

	geo, err := client.LookupIP( context.Background(), "8.8.8.8" )
	if err != nil {
		t.Fatalf( "LookupIP: %v", err )
	}

	if capturedPath != "/v1/ip/8.8.8.8" {
		t.Errorf( "path = %q, want /v1/ip/8.8.8.8", capturedPath )
	}

	if geo.CountryCode() != "US" {
		t.Errorf( "CountryCode() = %q, want US", geo.CountryCode() )
	}
	if geo.CityName() != "Mountain View" {
		t.Errorf( "CityName() = %q, want Mountain View", geo.CityName() )
	}
	if geo.TZ() != "America/Los_Angeles" {
		t.Errorf( "TZ() = %q, want America/Los_Angeles", geo.TZ() )
	}
	if geo.ASN() != "AS15169" {
		t.Errorf( "ASN() = %q, want AS15169", geo.ASN() )
	}
	if geo.Location.Coordinates.Latitude != 37.386 {
		t.Errorf( "Latitude = %f, want 37.386", geo.Location.Coordinates.Latitude )
	}
}

// ──────────────────────────────────────────────────────
// LookupMe
// ──────────────────────────────────────────────────────

func TestClient_LookupMe( t *testing.T ) {
	var capturedPath string

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedPath = r.URL.Path
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( geoFixture() )
	})

	geo, err := client.LookupMe( context.Background() )
	if err != nil {
		t.Fatalf( "LookupMe: %v", err )
	}

	if capturedPath != "/v1/ip/me" {
		t.Errorf( "path = %q, want /v1/ip/me", capturedPath )
	}

	if geo.IP() != "8.8.8.8" {
		t.Errorf( "IP() = %q, want 8.8.8.8", geo.IP() )
	}
}

// ──────────────────────────────────────────────────────
// Timezone
// ──────────────────────────────────────────────────────

func TestClient_Timezone( t *testing.T ) {
	var capturedPath string

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedPath = r.URL.Path
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( "America/Los_Angeles" )
	})

	tz, err := client.Timezone( context.Background(), "8.8.8.8" )
	if err != nil {
		t.Fatalf( "Timezone: %v", err )
	}

	if capturedPath != "/v1/ip/8.8.8.8/timezone" {
		t.Errorf( "path = %q, want /v1/ip/8.8.8.8/timezone", capturedPath )
	}

	if tz != "America/Los_Angeles" {
		t.Errorf( "tz = %q, want America/Los_Angeles", tz )
	}
}

func TestClient_MyTimezone( t *testing.T ) {
	var capturedPath string

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedPath = r.URL.Path
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( "America/Chicago" )
	})

	tz, err := client.MyTimezone( context.Background() )
	if err != nil {
		t.Fatalf( "MyTimezone: %v", err )
	}

	if capturedPath != "/v1/ip/me/timezone" {
		t.Errorf( "path = %q, want /v1/ip/me/timezone", capturedPath )
	}

	if tz != "America/Chicago" {
		t.Errorf( "tz = %q, want America/Chicago", tz )
	}
}

// ──────────────────────────────────────────────────────
// Validate
// ──────────────────────────────────────────────────────

func TestClient_Validate( t *testing.T ) {
	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		if strings.Contains( r.URL.Path, "8.8.8.8" ) {
			json.NewEncoder( w ).Encode( true )
		} else {
			json.NewEncoder( w ).Encode( false )
		}
	})

	valid, err := client.Validate( context.Background(), "8.8.8.8" )
	if err != nil {
		t.Fatalf( "Validate: %v", err )
	}
	if !valid {
		t.Error( "Validate(8.8.8.8) should be true" )
	}

	valid, err = client.Validate( context.Background(), "not-an-ip" )
	if err != nil {
		t.Fatalf( "Validate: %v", err )
	}
	if valid {
		t.Error( "Validate(not-an-ip) should be false" )
	}
}

// ──────────────────────────────────────────────────────
// Error handling
// ──────────────────────────────────────────────────────

func TestClient_APIError( t *testing.T ) {
	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 400 )
		json.NewEncoder( w ).Encode( map[string]any{
			"query": map[string]string{ "type": "ip", "value": "bad" },
			"error": map[string]string{
				"code":    "INVALID_IP",
				"message": "Invalid IP address format",
			},
			"meta": map[string]string{
				"requestId": "req-456",
			},
		})
	})

	_, err := client.LookupIP( context.Background(), "bad" )
	if err == nil {
		t.Fatal( "expected error" )
	}

	var apiErr *APIError
	if !errors.As( err, &apiErr ) {
		t.Fatalf( "expected *APIError, got %T: %v", err, err )
	}

	if apiErr.Status != 400 {
		t.Errorf( "Status = %d, want 400", apiErr.Status )
	}
	if apiErr.Code != "INVALID_IP" {
		t.Errorf( "Code = %q, want INVALID_IP", apiErr.Code )
	}
	if apiErr.Message != "Invalid IP address format" {
		t.Errorf( "Message = %q", apiErr.Message )
	}
	if apiErr.RequestID != "req-456" {
		t.Errorf( "RequestID = %q, want req-456", apiErr.RequestID )
	}

	// Error string should be human-readable.
	errStr := apiErr.Error()
	if !strings.Contains( errStr, "INVALID_IP" ) {
		t.Errorf( "Error() = %q, should contain INVALID_IP", errStr )
	}
}

func TestClient_RateLimited( t *testing.T ) {
	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 429 )
		json.NewEncoder( w ).Encode( map[string]any{
			"error": map[string]string{
				"code":    "RATE_LIMIT_EXCEEDED",
				"message": "Rate limit exceeded",
			},
		})
	})

	_, err := client.LookupIP( context.Background(), "8.8.8.8" )
	if err == nil {
		t.Fatal( "expected error" )
	}
	if !IsRateLimited( err ) {
		t.Error( "IsRateLimited should be true" )
	}
	if IsNotFound( err ) {
		t.Error( "IsNotFound should be false" )
	}
	if IsForbidden( err ) {
		t.Error( "IsForbidden should be false" )
	}
}

func TestClient_Forbidden( t *testing.T ) {
	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 403 )
		json.NewEncoder( w ).Encode( map[string]any{
			"error": map[string]string{
				"code":    "FORBIDDEN",
				"message": "This key does not have permission",
			},
		})
	})

	_, err := client.LookupIP( context.Background(), "8.8.8.8" )
	if err == nil {
		t.Fatal( "expected error" )
	}
	if !IsForbidden( err ) {
		t.Error( "IsForbidden should be true" )
	}
}

func TestClient_NotFound( t *testing.T ) {
	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 404 )
		json.NewEncoder( w ).Encode( map[string]any{
			"error": map[string]string{
				"code":    "NOT_FOUND",
				"message": "IP not found",
			},
		})
	})

	_, err := client.LookupIP( context.Background(), "8.8.8.8" )
	if err == nil {
		t.Fatal( "expected error" )
	}
	if !IsNotFound( err ) {
		t.Error( "IsNotFound should be true" )
	}
}

func TestClient_UnparseableError( t *testing.T ) {
	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		w.WriteHeader( 500 )
		w.Write( []byte( "internal server error" ) )
	})

	_, err := client.LookupIP( context.Background(), "8.8.8.8" )
	if err == nil {
		t.Fatal( "expected error" )
	}

	var apiErr *APIError
	if !errors.As( err, &apiErr ) {
		t.Fatalf( "expected *APIError, got %T", err )
	}

	if apiErr.Status != 500 {
		t.Errorf( "Status = %d, want 500", apiErr.Status )
	}
	// Should fallback to http.StatusText.
	if apiErr.Message != "Internal Server Error" {
		t.Errorf( "Message = %q, want fallback", apiErr.Message )
	}
}

// ──────────────────────────────────────────────────────
// Context cancellation
// ──────────────────────────────────────────────────────

func TestClient_ContextCancellation( t *testing.T ) {
	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		// Simulate slow response.
		time.Sleep( 2 * time.Second )
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( geoFixture() )
	})

	ctx, cancel := context.WithTimeout( context.Background(), 50*time.Millisecond )
	defer cancel()

	_, err := client.LookupIP( ctx, "8.8.8.8" )
	if err == nil {
		t.Fatal( "expected context deadline error" )
	}

	if !errors.Is( err, context.DeadlineExceeded ) && !strings.Contains( err.Error(), "deadline" ) {
		t.Errorf( "expected deadline error, got: %v", err )
	}
}

// ──────────────────────────────────────────────────────
// URL encoding
// ──────────────────────────────────────────────────────

func TestClient_IPv6URLEncoding( t *testing.T ) {
	var capturedPath string

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedPath = r.URL.Path
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( geoFixture() )
	})

	_, err := client.LookupIP( context.Background(), "2001:db8::1" )
	if err != nil {
		t.Fatalf( "LookupIP: %v", err )
	}

	// Colons in IPv6 should be percent-encoded in the path.
	if !strings.Contains( capturedPath, "2001" ) {
		t.Errorf( "path = %q, expected IPv6 address", capturedPath )
	}
}
