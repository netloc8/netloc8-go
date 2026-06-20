package netloc8

import (
	"context"
	"fmt"
	"net/url"
)

// ──────────────────────────────────────────────────────
// Account response types
// ──────────────────────────────────────────────────────

// Profile represents the authenticated user's account.
type Profile struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name,omitempty"`
	CreatedAt string `json:"createdAt"`
}

// APIKey represents an API key's metadata (never the raw key value).
type APIKey struct {
	ID        string   `json:"id"`
	Prefix    string   `json:"prefix"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`              // "secret" or "publishable"
	Scopes    []string `json:"scopes"`
	Status    string   `json:"status"`             // "active", "expired", "revoked"
	CreatedAt string   `json:"createdAt"`
	ExpiresAt string   `json:"expiresAt,omitempty"`
}

// CreatedKey is returned when a new key is created.
// RawKey is the full key value — it is only returned once.
type CreatedKey struct {
	APIKey
	RawKey string `json:"rawKey"`
}

// Usage holds API usage statistics for the current billing period.
type Usage struct {
	TotalKeys     int         `json:"totalKeys"`
	ActiveKeys    int         `json:"activeKeys"`
	TotalRequests int         `json:"totalRequests"`
	MonthlyCap    *int        `json:"monthlyCap"`
	DailyUsage    []DailyUsage `json:"dailyUsage"`
	Keys          []KeyUsage  `json:"keys"`
}

// DailyUsage holds request counts for a single day.
type DailyUsage struct {
	Date          string `json:"date"`
	TotalRequests int    `json:"totalRequests"`
}

// KeyUsage holds per-key usage and rate limit status.
type KeyUsage struct {
	KeyPrefix           string  `json:"keyPrefix"`
	KeyName             string  `json:"keyName"`
	IsActive            bool    `json:"isActive"`
	LastUsedAt          *string `json:"lastUsedAt"`
	RateLimitRemaining  int     `json:"rateLimitRemaining"`
	RateLimitMax        int     `json:"rateLimitMax"`
}

// AuditEntry represents a single audit log event.
type AuditEntry struct {
	ID         string `json:"id"`
	Action     string `json:"action"`
	ActorID    string `json:"actorId"`
	ActorLabel string `json:"actorLabel"`
	TargetType string `json:"targetType,omitempty"`
	TargetID   string `json:"targetId,omitempty"`
	CreatedAt  string `json:"createdAt"`
}

// AuditLog is the paginated response from the audit endpoint.
type AuditLog struct {
	Entries []AuditEntry `json:"entries"`
	Total   int          `json:"total"`
}

// Site represents an allowed origin for publishable keys.
type Site struct {
	ID     string `json:"id"`
	Origin string `json:"origin"`
}

// ──────────────────────────────────────────────────────
// Option types
// ──────────────────────────────────────────────────────

// CreateKeyOption configures key creation.
type CreateKeyOption func( *createKeyParams )

type createKeyParams struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

// WithKeyType sets the key type ("secret" or "publishable").
// If not specified, the API default is used.
func WithKeyType( t string ) CreateKeyOption {
	return func( p *createKeyParams ) {
		p.Type = t
	}
}

// AuditLogOption configures audit log queries.
type AuditLogOption func( *auditLogParams )

type auditLogParams struct {
	Limit  int
	Offset int
	Action string
}

// WithLimit sets the maximum number of audit entries to return.
func WithLimit( n int ) AuditLogOption {
	return func( p *auditLogParams ) {
		p.Limit = n
	}
}

// WithOffset sets the pagination offset for audit log queries.
func WithOffset( n int ) AuditLogOption {
	return func( p *auditLogParams ) {
		p.Offset = n
	}
}

// WithAction filters audit log entries by action type.
func WithAction( action string ) AuditLogOption {
	return func( p *auditLogParams ) {
		p.Action = action
	}
}

// ──────────────────────────────────────────────────────
// Profile
// ──────────────────────────────────────────────────────

// GetProfile returns the authenticated user's account profile.
func ( c *Client ) GetProfile( ctx context.Context ) ( *Profile, error ) {
	p := &Profile{}
	if err := c.doJSON( ctx, "/v1/account/me", p ); err != nil {
		return nil, err
	}
	return p, nil
}

// ──────────────────────────────────────────────────────
// API Keys
// ──────────────────────────────────────────────────────

// ListKeys returns all API keys for the authenticated account.
// The returned keys contain metadata only — raw key values are never
// returned after creation.
func ( c *Client ) ListKeys( ctx context.Context ) ( []APIKey, error ) {
	var keys []APIKey
	if err := c.doJSON( ctx, "/v1/account/me/keys", &keys ); err != nil {
		return nil, err
	}
	return keys, nil
}

// CreateKey creates a new API key with the given name.
// Use WithKeyType to specify "secret" or "publishable".
//
// The returned CreatedKey includes the RawKey field, which is the full
// key value. This is the only time the raw key is returned — store it
// securely.
func ( c *Client ) CreateKey( ctx context.Context, name string, opts ...CreateKeyOption ) ( *CreatedKey, error ) {
	params := &createKeyParams{ Name: name }
	for _, opt := range opts {
		opt( params )
	}

	key := &CreatedKey{}
	if err := c.doPost( ctx, "/v1/account/me/keys", params, key ); err != nil {
		return nil, err
	}
	return key, nil
}

// DeleteKey permanently deletes an API key by its ID.
// Returns nil on success (the API may respond with 204 No Content).
func ( c *Client ) DeleteKey( ctx context.Context, keyID string ) error {
	path := "/v1/account/me/keys/" + url.PathEscape( keyID )
	return c.doDelete( ctx, path, nil )
}

// RenewKey extends the expiration of an API key.
// Returns the updated key metadata.
func ( c *Client ) RenewKey( ctx context.Context, keyID string ) ( *APIKey, error ) {
	path := fmt.Sprintf( "/v1/account/me/keys/%s/renew", url.PathEscape( keyID ) )
	key := &APIKey{}
	if err := c.doPost( ctx, path, nil, key ); err != nil {
		return nil, err
	}
	return key, nil
}

// ──────────────────────────────────────────────────────
// Usage
// ──────────────────────────────────────────────────────

// GetUsage returns API usage statistics for the current billing period.
func ( c *Client ) GetUsage( ctx context.Context ) ( *Usage, error ) {
	u := &Usage{}
	if err := c.doJSON( ctx, "/v1/account/me/usage", u ); err != nil {
		return nil, err
	}
	return u, nil
}

// ──────────────────────────────────────────────────────
// Audit Log
// ──────────────────────────────────────────────────────

// GetAuditLog returns a paginated list of audit log entries for the
// authenticated account. Use WithLimit, WithOffset, and WithAction to
// filter and paginate results.
func ( c *Client ) GetAuditLog( ctx context.Context, opts ...AuditLogOption ) ( *AuditLog, error ) {
	params := &auditLogParams{}
	for _, opt := range opts {
		opt( params )
	}

	path := "/v1/account/me/audit"

	// Build query string from options.
	q := url.Values{}
	if params.Limit > 0 {
		q.Set( "limit", fmt.Sprintf( "%d", params.Limit ) )
	}
	if params.Offset > 0 {
		q.Set( "offset", fmt.Sprintf( "%d", params.Offset ) )
	}
	if params.Action != "" {
		q.Set( "action", params.Action )
	}
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	log := &AuditLog{}
	if err := c.doJSON( ctx, path, log ); err != nil {
		return nil, err
	}
	return log, nil
}

// ──────────────────────────────────────────────────────
// Sites (allowed origins for publishable keys)
// ──────────────────────────────────────────────────────

// ListSites returns all allowed origins for publishable keys.
func ( c *Client ) ListSites( ctx context.Context ) ( []Site, error ) {
	var sites []Site
	if err := c.doJSON( ctx, "/v1/account/me/sites", &sites ); err != nil {
		return nil, err
	}
	return sites, nil
}

// CreateSite adds an allowed origin for publishable keys.
// The origin should be a full URL (e.g. "https://example.com").
func ( c *Client ) CreateSite( ctx context.Context, origin string ) ( *Site, error ) {
	body := struct {
		Origin string `json:"origin"`
	}{ Origin: origin }

	site := &Site{}
	if err := c.doPost( ctx, "/v1/account/me/sites", body, site ); err != nil {
		return nil, err
	}
	return site, nil
}

// DeleteSite removes an allowed origin by its ID.
// Returns nil on success (the API may respond with 204 No Content).
func ( c *Client ) DeleteSite( ctx context.Context, siteID string ) error {
	path := "/v1/account/me/sites/" + url.PathEscape( siteID )
	return c.doDelete( ctx, path, nil )
}
