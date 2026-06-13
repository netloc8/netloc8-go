// Package netloc8 provides an idiomatic Go client for the NetLoc8 IP
// geolocation API.
//
// NetLoc8 resolves IP addresses to geographic locations, timezones,
// coordinates, ASN data, and EU membership — with city-level precision.
//
// # Quick Start
//
// Create a client with your API key and look up any IP address:
//
//	client := netloc8.NewClient( "sk_your_secret_key" )
//
//
//	geo, err := client.LookupIP( ctx, "8.8.8.8" )
//	if err != nil {
//	    log.Fatal( err )
//	}
//
//	fmt.Println( geo.Location.Country.Code ) // "US"
//	fmt.Println( geo.Location.City )          // "Mountain View"
//	fmt.Println( geo.Location.Timezone )      // "America/Los_Angeles"
//	fmt.Println( geo.Network.Organization )   // "Google LLC"
//
// # Self-Lookup
//
// Discover the caller's own IP geolocation (useful for proxy exit-IP
// detection):
//
//	geo, err := client.LookupMe( ctx )
//
// # Timezone Only
//
// Fetch just the IANA timezone string for lightweight use:
//
//	tz, err := client.Timezone( ctx, "8.8.8.8" )
//	// tz == "America/Los_Angeles"
//
// # IP Utilities
//
// Standalone functions for IP validation, normalization, and classification
// that work without an API client:
//
//	netloc8.IsPublicIP( "192.168.1.1" )           // false
//	netloc8.IsPublicIP( "8.8.8.8" )               // true
//	netloc8.NormalizeIP( "::ffff:8.8.8.8" )       // "8.8.8.8"
//	netloc8.Subnet( "203.0.113.42" )              // "203.0.113.0/24"
//
// # EU Detection
//
// Check whether a geolocation result is in the European Union:
//
//	if netloc8.IsEU( geo ) {
//	    showCookieConsent()
//	}
//
// # Error Handling
//
// API errors are returned as *APIError, which includes the HTTP status
// code and structured error code from the API:
//
//	geo, err := client.LookupIP( ctx, "not-an-ip" )
//	var apiErr *netloc8.APIError
//	if errors.As( err, &apiErr ) {
//	    fmt.Println( apiErr.Code )   // "INVALID_IP"
//	    fmt.Println( apiErr.Status ) // 400
//	}
//
// # Proxy-Aware
//
// Pass a custom http.Client with a proxy transport for exit-IP discovery:
//
//	transport := &http.Transport{
//	    Proxy: http.ProxyURL( proxyURL ),
//	}
//	client := netloc8.NewClient( "sk_key",
//	    netloc8.WithHTTPClient( &http.Client{ Transport: transport } ),
//	)
//
//	geo, err := client.LookupMe( ctx )
//	// geo.Query.Value is the proxy's exit IP
//
// # Account Management
//
// Manage your account, API keys, usage, and allowed origins with a
// secret key (sk_):
//
//	client := netloc8.NewClient( "sk_your_secret_key" )
//
//	// Fetch your profile.
//	profile, err := client.GetProfile( ctx )
//	fmt.Println( profile.Email )
//
//	// List all API keys.
//	keys, err := client.ListKeys( ctx )
//	for _, k := range keys {
//	    fmt.Printf( "%s (%s) — %s\n", k.Name, k.Prefix, k.Status )
//	}
//
//	// Create a new secret key.
//	created, err := client.CreateKey( ctx, "CI Pipeline",
//	    netloc8.WithKeyType( "secret" ),
//	)
//	fmt.Println( created.RawKey ) // store this — only shown once
//
//	// Delete a key.
//	err = client.DeleteKey( ctx, "key_id" )
//
//	// Renew a key before it expires.
//	renewed, err := client.RenewKey( ctx, "key_id" )
//
// # Usage Statistics
//
// Check API consumption for the current billing period:
//
//	usage, err := client.GetUsage( ctx )
//	fmt.Printf( "%d / %d requests (%s)\n", usage.Total, usage.Cap, usage.Period )
//
// # Audit Log
//
// Query account activity with optional filters:
//
//	log, err := client.GetAuditLog( ctx,
//	    netloc8.WithLimit( 25 ),
//	    netloc8.WithAction( "key.created" ),
//	)
//	for _, entry := range log.Entries {
//	    fmt.Printf( "%s: %s by %s\n", entry.CreatedAt, entry.Action, entry.ActorLabel )
//	}
//
// # Allowed Origins
//
// Manage allowed origins for publishable keys:
//
//	sites, err := client.ListSites( ctx )
//	site, err := client.CreateSite( ctx, "https://my-app.com" )
//	err = client.DeleteSite( ctx, site.ID )
package netloc8
