package netloc8

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// ──────────────────────────────────────────────────────
// GetProfile
// ──────────────────────────────────────────────────────

func TestClient_GetProfile( t *testing.T ) {
	var capturedPath string
	var capturedMethod string

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedPath = r.URL.Path
		capturedMethod = r.Method
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( Profile{
			ID:        "usr_abc123",
			Email:     "tom@example.com",
			Name:      "Tom Voss",
			CreatedAt: "2024-01-15T10:30:00Z",
		})
	})

	p, err := client.GetProfile( context.Background() )
	if err != nil {
		t.Fatalf( "GetProfile: %v", err )
	}

	if capturedMethod != http.MethodGet {
		t.Errorf( "method = %q, want GET", capturedMethod )
	}
	if capturedPath != "/v1/account/me" {
		t.Errorf( "path = %q, want /v1/account/me", capturedPath )
	}
	if p.ID != "usr_abc123" {
		t.Errorf( "ID = %q, want usr_abc123", p.ID )
	}
	if p.Email != "tom@example.com" {
		t.Errorf( "Email = %q, want tom@example.com", p.Email )
	}
	if p.Name != "Tom Voss" {
		t.Errorf( "Name = %q, want Tom Voss", p.Name )
	}
}

func TestClient_GetProfile_Error( t *testing.T ) {
	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 401 )
		json.NewEncoder( w ).Encode( map[string]any{
			"error": map[string]string{
				"code":    "UNAUTHORIZED",
				"message": "Invalid API key",
			},
		})
	})

	_, err := client.GetProfile( context.Background() )
	if err == nil {
		t.Fatal( "expected error" )
	}

	var apiErr *APIError
	if !errors.As( err, &apiErr ) {
		t.Fatalf( "expected *APIError, got %T: %v", err, err )
	}
	if apiErr.Status != 401 {
		t.Errorf( "Status = %d, want 401", apiErr.Status )
	}
	if apiErr.Code != "UNAUTHORIZED" {
		t.Errorf( "Code = %q, want UNAUTHORIZED", apiErr.Code )
	}
}

// ──────────────────────────────────────────────────────
// ListKeys
// ──────────────────────────────────────────────────────

func TestClient_ListKeys( t *testing.T ) {
	var capturedPath string

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedPath = r.URL.Path
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( []APIKey{
			{
				ID:        "key_001",
				Prefix:    "sk_live_",
				Name:      "Production",
				Type:      "secret",
				Scopes:    []string{ "geo:read", "account:read" },
				Status:    "active",
				CreatedAt: "2024-01-15T10:30:00Z",
			},
			{
				ID:        "key_002",
				Prefix:    "pk_live_",
				Name:      "Frontend",
				Type:      "publishable",
				Scopes:    []string{ "geo:read" },
				Status:    "active",
				CreatedAt: "2024-02-01T08:00:00Z",
				ExpiresAt: "2025-02-01T08:00:00Z",
			},
		})
	})

	keys, err := client.ListKeys( context.Background() )
	if err != nil {
		t.Fatalf( "ListKeys: %v", err )
	}

	if capturedPath != "/v1/account/me/keys" {
		t.Errorf( "path = %q, want /v1/account/me/keys", capturedPath )
	}
	if len( keys ) != 2 {
		t.Fatalf( "len(keys) = %d, want 2", len( keys ) )
	}
	if keys[0].Name != "Production" {
		t.Errorf( "keys[0].Name = %q, want Production", keys[0].Name )
	}
	if keys[1].Type != "publishable" {
		t.Errorf( "keys[1].Type = %q, want publishable", keys[1].Type )
	}
	if len( keys[0].Scopes ) != 2 {
		t.Errorf( "len(keys[0].Scopes) = %d, want 2", len( keys[0].Scopes ) )
	}
}

func TestClient_ListKeys_Error( t *testing.T ) {
	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 403 )
		json.NewEncoder( w ).Encode( map[string]any{
			"error": map[string]string{
				"code":    "FORBIDDEN",
				"message": "Insufficient permissions",
			},
		})
	})

	_, err := client.ListKeys( context.Background() )
	if err == nil {
		t.Fatal( "expected error" )
	}
	if !IsForbidden( err ) {
		t.Error( "IsForbidden should be true" )
	}
}

// ──────────────────────────────────────────────────────
// CreateKey
// ──────────────────────────────────────────────────────

func TestClient_CreateKey( t *testing.T ) {
	var capturedPath string
	var capturedMethod string
	var capturedBody map[string]any

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedPath = r.URL.Path
		capturedMethod = r.Method

		body, _ := io.ReadAll( r.Body )
		json.Unmarshal( body, &capturedBody )

		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 201 )
		json.NewEncoder( w ).Encode( CreatedKey{
			APIKey: APIKey{
				ID:        "key_003",
				Prefix:    "sk_live_",
				Name:      "CI Pipeline",
				Type:      "secret",
				Scopes:    []string{ "geo:read" },
				Status:    "active",
				CreatedAt: "2024-06-13T12:00:00Z",
			},
			RawKey: "sk_live_abc123def456",
		})
	})

	key, err := client.CreateKey( context.Background(), "CI Pipeline", WithKeyType( "secret" ) )
	if err != nil {
		t.Fatalf( "CreateKey: %v", err )
	}

	if capturedMethod != http.MethodPost {
		t.Errorf( "method = %q, want POST", capturedMethod )
	}
	if capturedPath != "/v1/account/me/keys" {
		t.Errorf( "path = %q, want /v1/account/me/keys", capturedPath )
	}

	// Verify request body.
	if capturedBody["name"] != "CI Pipeline" {
		t.Errorf( "body.name = %q, want CI Pipeline", capturedBody["name"] )
	}
	if capturedBody["type"] != "secret" {
		t.Errorf( "body.type = %q, want secret", capturedBody["type"] )
	}

	// Verify response.
	if key.ID != "key_003" {
		t.Errorf( "ID = %q, want key_003", key.ID )
	}
	if key.RawKey != "sk_live_abc123def456" {
		t.Errorf( "RawKey = %q, want sk_live_abc123def456", key.RawKey )
	}
	if key.Name != "CI Pipeline" {
		t.Errorf( "Name = %q, want CI Pipeline", key.Name )
	}
}

func TestClient_CreateKey_NoType( t *testing.T ) {
	var capturedBody map[string]any

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		body, _ := io.ReadAll( r.Body )
		json.Unmarshal( body, &capturedBody )

		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 201 )
		json.NewEncoder( w ).Encode( CreatedKey{
			APIKey: APIKey{
				ID:   "key_004",
				Name: "Default Type",
			},
			RawKey: "sk_live_xyz",
		})
	})

	_, err := client.CreateKey( context.Background(), "Default Type" )
	if err != nil {
		t.Fatalf( "CreateKey: %v", err )
	}

	// type should be omitted when WithKeyType is not called.
	if _, ok := capturedBody["type"]; ok {
		val := capturedBody["type"]
		if val != nil && val != "" {
			t.Errorf( "body.type should be empty or omitted, got %q", val )
		}
	}
}

func TestClient_CreateKey_Error( t *testing.T ) {
	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 400 )
		json.NewEncoder( w ).Encode( map[string]any{
			"error": map[string]string{
				"code":    "INVALID_REQUEST",
				"message": "Name is required",
			},
		})
	})

	_, err := client.CreateKey( context.Background(), "" )
	if err == nil {
		t.Fatal( "expected error" )
	}

	var apiErr *APIError
	if !errors.As( err, &apiErr ) {
		t.Fatalf( "expected *APIError, got %T", err )
	}
	if apiErr.Status != 400 {
		t.Errorf( "Status = %d, want 400", apiErr.Status )
	}
}

// ──────────────────────────────────────────────────────
// DeleteKey
// ──────────────────────────────────────────────────────

func TestClient_DeleteKey( t *testing.T ) {
	var capturedPath string
	var capturedMethod string

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedPath = r.URL.Path
		capturedMethod = r.Method
		w.WriteHeader( 204 )
	})

	err := client.DeleteKey( context.Background(), "key_001" )
	if err != nil {
		t.Fatalf( "DeleteKey: %v", err )
	}

	if capturedMethod != http.MethodDelete {
		t.Errorf( "method = %q, want DELETE", capturedMethod )
	}
	if capturedPath != "/v1/account/me/keys/key_001" {
		t.Errorf( "path = %q, want /v1/account/me/keys/key_001", capturedPath )
	}
}

func TestClient_DeleteKey_Error( t *testing.T ) {
	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 404 )
		json.NewEncoder( w ).Encode( map[string]any{
			"error": map[string]string{
				"code":    "NOT_FOUND",
				"message": "Key not found",
			},
		})
	})

	err := client.DeleteKey( context.Background(), "key_nonexistent" )
	if err == nil {
		t.Fatal( "expected error" )
	}
	if !IsNotFound( err ) {
		t.Error( "IsNotFound should be true" )
	}
}

// ──────────────────────────────────────────────────────
// RenewKey
// ──────────────────────────────────────────────────────

func TestClient_RenewKey( t *testing.T ) {
	var capturedPath string
	var capturedMethod string

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedPath = r.URL.Path
		capturedMethod = r.Method
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( APIKey{
			ID:        "key_001",
			Prefix:    "sk_live_",
			Name:      "Production",
			Type:      "secret",
			Status:    "active",
			CreatedAt: "2024-01-15T10:30:00Z",
			ExpiresAt: "2025-06-13T10:30:00Z",
		})
	})

	key, err := client.RenewKey( context.Background(), "key_001" )
	if err != nil {
		t.Fatalf( "RenewKey: %v", err )
	}

	if capturedMethod != http.MethodPost {
		t.Errorf( "method = %q, want POST", capturedMethod )
	}
	if capturedPath != "/v1/account/me/keys/key_001/renew" {
		t.Errorf( "path = %q, want /v1/account/me/keys/key_001/renew", capturedPath )
	}
	if key.ExpiresAt != "2025-06-13T10:30:00Z" {
		t.Errorf( "ExpiresAt = %q, want 2025-06-13T10:30:00Z", key.ExpiresAt )
	}
}

func TestClient_RenewKey_Error( t *testing.T ) {
	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 404 )
		json.NewEncoder( w ).Encode( map[string]any{
			"error": map[string]string{
				"code":    "NOT_FOUND",
				"message": "Key not found",
			},
		})
	})

	_, err := client.RenewKey( context.Background(), "key_nonexistent" )
	if err == nil {
		t.Fatal( "expected error" )
	}
	if !IsNotFound( err ) {
		t.Error( "IsNotFound should be true" )
	}
}

// ──────────────────────────────────────────────────────
// GetUsage
// ──────────────────────────────────────────────────────

func TestClient_GetUsage( t *testing.T ) {
	var capturedPath string

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedPath = r.URL.Path
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( Usage{
			Total:  15432,
			Cap:    100000,
			Period: "2024-06",
			Daily: map[string]int{
				"2024-06-12": 512,
				"2024-06-13": 489,
			},
			ByKey: map[string]int{
				"key_001": 12000,
				"key_002": 3432,
			},
		})
	})

	u, err := client.GetUsage( context.Background() )
	if err != nil {
		t.Fatalf( "GetUsage: %v", err )
	}

	if capturedPath != "/v1/account/me/usage" {
		t.Errorf( "path = %q, want /v1/account/me/usage", capturedPath )
	}
	if u.Total != 15432 {
		t.Errorf( "Total = %d, want 15432", u.Total )
	}
	if u.Cap != 100000 {
		t.Errorf( "Cap = %d, want 100000", u.Cap )
	}
	if u.Period != "2024-06" {
		t.Errorf( "Period = %q, want 2024-06", u.Period )
	}
	if len( u.Daily ) != 2 {
		t.Errorf( "len(Daily) = %d, want 2", len( u.Daily ) )
	}
	if u.ByKey["key_001"] != 12000 {
		t.Errorf( "ByKey[key_001] = %d, want 12000", u.ByKey["key_001"] )
	}
}

func TestClient_GetUsage_Error( t *testing.T ) {
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

	_, err := client.GetUsage( context.Background() )
	if err == nil {
		t.Fatal( "expected error" )
	}
	if !IsRateLimited( err ) {
		t.Error( "IsRateLimited should be true" )
	}
}

// ──────────────────────────────────────────────────────
// GetAuditLog
// ──────────────────────────────────────────────────────

func TestClient_GetAuditLog( t *testing.T ) {
	var capturedPath string
	var capturedQuery string

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedPath = r.URL.Path
		capturedQuery = r.URL.RawQuery
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( AuditLog{
			Entries: []AuditEntry{
				{
					ID:         "aud_001",
					Action:     "key.created",
					ActorID:    "usr_abc123",
					ActorLabel: "tom@example.com",
					TargetType: "api_key",
					TargetID:   "key_003",
					CreatedAt:  "2024-06-13T12:00:00Z",
				},
			},
			Total: 1,
		})
	})

	log, err := client.GetAuditLog( context.Background(),
		WithLimit( 10 ),
		WithOffset( 5 ),
		WithAction( "key.created" ),
	)
	if err != nil {
		t.Fatalf( "GetAuditLog: %v", err )
	}

	if capturedPath != "/v1/account/me/audit" {
		t.Errorf( "path = %q, want /v1/account/me/audit", capturedPath )
	}

	// Verify query parameters.
	if !strings.Contains( capturedQuery, "limit=10" ) {
		t.Errorf( "query = %q, should contain limit=10", capturedQuery )
	}
	if !strings.Contains( capturedQuery, "offset=5" ) {
		t.Errorf( "query = %q, should contain offset=5", capturedQuery )
	}
	if !strings.Contains( capturedQuery, "action=key.created" ) {
		t.Errorf( "query = %q, should contain action=key.created", capturedQuery )
	}

	if log.Total != 1 {
		t.Errorf( "Total = %d, want 1", log.Total )
	}
	if len( log.Entries ) != 1 {
		t.Fatalf( "len(Entries) = %d, want 1", len( log.Entries ) )
	}
	if log.Entries[0].Action != "key.created" {
		t.Errorf( "Entries[0].Action = %q, want key.created", log.Entries[0].Action )
	}
	if log.Entries[0].TargetType != "api_key" {
		t.Errorf( "Entries[0].TargetType = %q, want api_key", log.Entries[0].TargetType )
	}
}

func TestClient_GetAuditLog_NoOptions( t *testing.T ) {
	var capturedQuery string

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( AuditLog{
			Entries: []AuditEntry{},
			Total:   0,
		})
	})

	_, err := client.GetAuditLog( context.Background() )
	if err != nil {
		t.Fatalf( "GetAuditLog: %v", err )
	}

	if capturedQuery != "" {
		t.Errorf( "query = %q, want empty (no options)", capturedQuery )
	}
}

func TestClient_GetAuditLog_Error( t *testing.T ) {
	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 403 )
		json.NewEncoder( w ).Encode( map[string]any{
			"error": map[string]string{
				"code":    "FORBIDDEN",
				"message": "Insufficient permissions",
			},
		})
	})

	_, err := client.GetAuditLog( context.Background() )
	if err == nil {
		t.Fatal( "expected error" )
	}
	if !IsForbidden( err ) {
		t.Error( "IsForbidden should be true" )
	}
}

// ──────────────────────────────────────────────────────
// ListSites
// ──────────────────────────────────────────────────────

func TestClient_ListSites( t *testing.T ) {
	var capturedPath string

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedPath = r.URL.Path
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( []Site{
			{ ID: "site_001", Origin: "https://example.com" },
			{ ID: "site_002", Origin: "https://staging.example.com" },
		})
	})

	sites, err := client.ListSites( context.Background() )
	if err != nil {
		t.Fatalf( "ListSites: %v", err )
	}

	if capturedPath != "/v1/account/me/sites" {
		t.Errorf( "path = %q, want /v1/account/me/sites", capturedPath )
	}
	if len( sites ) != 2 {
		t.Fatalf( "len(sites) = %d, want 2", len( sites ) )
	}
	if sites[0].Origin != "https://example.com" {
		t.Errorf( "sites[0].Origin = %q, want https://example.com", sites[0].Origin )
	}
}

// ──────────────────────────────────────────────────────
// CreateSite
// ──────────────────────────────────────────────────────

func TestClient_CreateSite( t *testing.T ) {
	var capturedPath string
	var capturedMethod string
	var capturedBody map[string]any

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedPath = r.URL.Path
		capturedMethod = r.Method

		body, _ := io.ReadAll( r.Body )
		json.Unmarshal( body, &capturedBody )

		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 201 )
		json.NewEncoder( w ).Encode( Site{
			ID:     "site_003",
			Origin: "https://app.example.com",
		})
	})

	site, err := client.CreateSite( context.Background(), "https://app.example.com" )
	if err != nil {
		t.Fatalf( "CreateSite: %v", err )
	}

	if capturedMethod != http.MethodPost {
		t.Errorf( "method = %q, want POST", capturedMethod )
	}
	if capturedPath != "/v1/account/me/sites" {
		t.Errorf( "path = %q, want /v1/account/me/sites", capturedPath )
	}
	if capturedBody["origin"] != "https://app.example.com" {
		t.Errorf( "body.origin = %q, want https://app.example.com", capturedBody["origin"] )
	}
	if site.ID != "site_003" {
		t.Errorf( "ID = %q, want site_003", site.ID )
	}
	if site.Origin != "https://app.example.com" {
		t.Errorf( "Origin = %q, want https://app.example.com", site.Origin )
	}
}

func TestClient_CreateSite_Error( t *testing.T ) {
	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 400 )
		json.NewEncoder( w ).Encode( map[string]any{
			"error": map[string]string{
				"code":    "INVALID_REQUEST",
				"message": "Invalid origin format",
			},
		})
	})

	_, err := client.CreateSite( context.Background(), "not-a-url" )
	if err == nil {
		t.Fatal( "expected error" )
	}

	var apiErr *APIError
	if !errors.As( err, &apiErr ) {
		t.Fatalf( "expected *APIError, got %T", err )
	}
	if apiErr.Status != 400 {
		t.Errorf( "Status = %d, want 400", apiErr.Status )
	}
}

// ──────────────────────────────────────────────────────
// DeleteSite
// ──────────────────────────────────────────────────────

func TestClient_DeleteSite( t *testing.T ) {
	var capturedPath string
	var capturedMethod string

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedPath = r.URL.Path
		capturedMethod = r.Method
		w.WriteHeader( 204 )
	})

	err := client.DeleteSite( context.Background(), "site_001" )
	if err != nil {
		t.Fatalf( "DeleteSite: %v", err )
	}

	if capturedMethod != http.MethodDelete {
		t.Errorf( "method = %q, want DELETE", capturedMethod )
	}
	if capturedPath != "/v1/account/me/sites/site_001" {
		t.Errorf( "path = %q, want /v1/account/me/sites/site_001", capturedPath )
	}
}

func TestClient_DeleteSite_Error( t *testing.T ) {
	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 404 )
		json.NewEncoder( w ).Encode( map[string]any{
			"error": map[string]string{
				"code":    "NOT_FOUND",
				"message": "Site not found",
			},
		})
	})

	err := client.DeleteSite( context.Background(), "site_nonexistent" )
	if err == nil {
		t.Fatal( "expected error" )
	}
	if !IsNotFound( err ) {
		t.Error( "IsNotFound should be true" )
	}
}

// ──────────────────────────────────────────────────────
// doPost / doDelete transport: request shape
// ──────────────────────────────────────────────────────

func TestClient_PostHeaders( t *testing.T ) {
	var capturedHeaders http.Header

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedHeaders = r.Header.Clone()
		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 201 )
		json.NewEncoder( w ).Encode( Site{ ID: "site_001", Origin: "https://example.com" } )
	})

	_, err := client.CreateSite( context.Background(), "https://example.com" )
	if err != nil {
		t.Fatalf( "CreateSite: %v", err )
	}

	// Content-Type should be set for POST requests.
	if got := capturedHeaders.Get( "Content-Type" ); got != "application/json" {
		t.Errorf( "Content-Type = %q, want application/json", got )
	}

	// API key should still be sent.
	if got := capturedHeaders.Get( headerAPIKey ); got != "pk_test_key" {
		t.Errorf( "X-API-Key = %q, want pk_test_key", got )
	}
}

func TestClient_DeleteHeaders( t *testing.T ) {
	var capturedHeaders http.Header

	client, _ := newTestServer( t, func( w http.ResponseWriter, r *http.Request ) {
		capturedHeaders = r.Header.Clone()
		w.WriteHeader( 204 )
	})

	err := client.DeleteKey( context.Background(), "key_001" )
	if err != nil {
		t.Fatalf( "DeleteKey: %v", err )
	}

	// API key should be sent on DELETE requests.
	if got := capturedHeaders.Get( headerAPIKey ); got != "pk_test_key" {
		t.Errorf( "X-API-Key = %q, want pk_test_key", got )
	}

	// Accept header should be set.
	if got := capturedHeaders.Get( headerAccept ); got != "application/json" {
		t.Errorf( "Accept = %q, want application/json", got )
	}
}
