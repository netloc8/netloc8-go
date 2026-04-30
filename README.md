# NetLoc8 Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/netloc8/netloc8-go.svg)](https://pkg.go.dev/github.com/netloc8/netloc8-go)
[![License: ELv2](https://img.shields.io/badge/license-ELv2-blue)](LICENSE)

The easiest way to add **IP geolocation** to Go applications. Look up any IP
address to get country, city, region, timezone, coordinates, ASN, and EU
membership — with city-level precision.

Built for developers who need geolocation without wiring up raw API calls:
geo-aware proxy routing, exit-IP verification, EU compliance gating,
timezone-accurate scheduling, and regional content delivery.

## Why NetLoc8?

- **Zero dependencies** — only the Go standard library. No transitive dependency tree to audit
- **Proxy-aware** — plug in any `*http.Client` with a proxy transport to discover exit IPs through tunnels
- **Context-native** — every API call takes `context.Context` for cancellation, timeouts, and tracing
- **Nil-safe accessors** — `geo.CountryCode()`, `geo.CityName()`, `geo.TZ()` never panic on partial responses
- **Typed errors** — `*APIError` with machine-readable codes, `errors.As` support, and convenience predicates
- **Privacy compliance ready** — `IsEU()` helper for GDPR, country and region fields for CCPA and other regulations

### Common Use Cases

| Use Case | How |
|----------|-----|
| Detect visitor country | `geo.CountryCode()` → `"US"` |
| Show cookie consent to EU users | `netloc8.IsEU( geo )` |
| Discover proxy exit IP | `client.LookupMe( ctx )` with proxy transport |
| Derive subnet for rotation | `netloc8.Subnet( geo.IP() )` → `"203.0.113.0/24"` |
| Display local timezone | `geo.TZ()` → `"America/Los_Angeles"` |
| Classify IP as public/private | `netloc8.IsPublicIP( ip )` |
| Extract client IP from headers | `netloc8.GetClientIP( r )` |

## Install

```bash
go get github.com/netloc8/netloc8-go
```

[Documentation](https://netloc8.com/docs) · [API Reference](https://netloc8.com/docs/api) · [Dashboard](https://netloc8.com)

## Quick Start

```go
import (
    "context"
    "fmt"

    "github.com/netloc8/netloc8-go"
)

client := netloc8.NewClient( "sk_your_secret_key" )

geo, err := client.LookupIP( context.Background(), "8.8.8.8" )
if err != nil {
    log.Fatal( err )
}

fmt.Println( geo.CountryCode() )             // "US"
fmt.Println( geo.CityName() )                // "Mountain View"
fmt.Println( geo.TZ() )                      // "America/Los_Angeles"
fmt.Println( geo.ASN() )                     // "AS15169"
fmt.Println( geo.Org() )                     // "Google LLC"
```

## Response Shape

All lookup methods return a `*Geo` struct matching the API's `GeolocationResult`:

```go
geo, _ := client.LookupIP( ctx, "8.8.8.8" )

geo.IP()                                // "8.8.8.8"
geo.Query.IPVersion                     // 4
geo.CountryCode()                       // "US"
geo.CountryName()                       // "United States"
geo.Location.Country.Flag               // "🇺🇸"
geo.Location.Country.Unions             // ["EU"] or []
geo.Location.Region.Code                // "CA"
geo.RegionName()                        // "California"
geo.CityName()                          // "Mountain View"
geo.Location.PostalCode                 // "94043"
geo.Location.Coordinates.Latitude       // 37.386
geo.Location.Coordinates.Longitude      // -122.084
geo.TZ()                                // "America/Los_Angeles"
geo.Location.UTCOffset                  // "-07:00"
geo.Location.GeoConfidence              // 1.0
geo.ASN()                               // "AS15169"
geo.Org()                               // "Google LLC"
geo.Network.Domain                      // "google.com"
geo.Meta.Precision                      // "city"
```

All accessor methods (`CountryCode()`, `CityName()`, `TZ()`, etc.) are nil-safe
— they return empty strings on nil or partial responses, so you never need to
nil-check nested structs.

## API Methods

All methods accept `context.Context` for cancellation and timeouts.

| Method | Description |
|--------|-------------|
| `LookupIP(ctx, ip)` | Full geolocation for a specific IP address |
| `LookupMe(ctx)` | Full geolocation for the caller's own IP |
| `Timezone(ctx, ip)` | IANA timezone string for a specific IP |
| `MyTimezone(ctx)` | IANA timezone string for the caller's own IP |
| `Validate(ctx, ip)` | Check whether a string is a valid IP address (via API) |

## Client Options

| Option | Description |
|--------|-------------|
| `WithTimeout(d)` | Request timeout (default: 10s) |
| `WithHTTPClient(c)` | Custom `*http.Client` for proxy transports, custom TLS, connection pooling |
| `WithBaseURL(url)` | Override the API base URL (for testing or self-hosted deployments) |
| `WithUserAgent(ua)` | Prepend a custom User-Agent string |
| `WithOrigin(url)` | Set `Origin` header — only needed for publishable keys (`pk_`) |

## IP Utilities

Standalone functions that work locally — no API client or network access needed:

| Function | Description |
|----------|-------------|
| `NormalizeIP(ip)` | Strip IPv4-mapped prefix, brackets, whitespace, lowercase |
| `IsPublicIP(ip)` | Check if IP is publicly routable (rejects RFC1918, CGNAT, loopback, link-local, ULA) |
| `IsIPv4(ip)` / `IsIPv6(ip)` | IP version detection |
| `ParseIP(ip)` | Validate and normalize an IP string |
| `Subnet(ip)` | Derive `/24` CIDR prefix from IPv4 |
| `IsEU(geo)` | Check EU membership via `country.unions` |
| `GetClientIP(r)` | Extract the real client IP from HTTP request headers |

## Error Handling

All fetch methods return errors — they never panic. API errors are returned as
`*APIError` with structured error codes:

```go
geo, err := client.LookupIP( ctx, "not-an-ip" )

var apiErr *netloc8.APIError
if errors.As( err, &apiErr ) {
    fmt.Println( apiErr.Code )      // "INVALID_IP"
    fmt.Println( apiErr.Message )   // "Invalid IP address format"
    fmt.Println( apiErr.Status )    // 400
    fmt.Println( apiErr.RequestID ) // "req-abc-123"
}
```

Convenience predicates:

```go
netloc8.IsNotFound( err )    // 404 — IP not in database
netloc8.IsRateLimited( err ) // 429 — rate limit exceeded
netloc8.IsForbidden( err )   // 403 — invalid key or origin mismatch
```

## Proxy-Aware Usage

Discover proxy exit IPs by passing a custom transport. The request tunnels
through the proxy — `LookupMe` returns the proxy's exit IP, not yours:

```go
proxyURL, _ := url.Parse( "http://user:pass@proxy:8080" )

client := netloc8.NewClient( "pk_key",
    netloc8.WithOrigin( "https://your-app.com" ),
    netloc8.WithHTTPClient( &http.Client{
        Transport: &http.Transport{ Proxy: http.ProxyURL( proxyURL ) },
    }),
)

geo, _ := client.LookupMe( ctx )

fmt.Println( geo.IP() )                          // proxy's exit IP
fmt.Println( geo.CountryCode() )                 // proxy's exit country
fmt.Println( netloc8.Subnet( geo.IP() ) )        // "203.0.113.0/24"
```

## Server-Side IP Extraction

Extract the real client IP from reverse proxy headers in HTTP handlers:

```go
func handler( w http.ResponseWriter, r *http.Request ) {
    ip := netloc8.GetClientIP( r )
    // Checks in order: X-Forwarded-For, CF-Connecting-IP,
    //   True-Client-IP, X-Real-IP, X-Client-IP,
    //   Fastly-Client-IP, Fly-Client-IP
}
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `NETLOC8_API_KEY` | No | Fallback API key if not passed to `NewClient` |

### API Key Types

| Key type | Prefix | Usage |
|----------|--------|-------|
| Secret | `sk_` | Server-side — full access, no origin restriction |
| Publishable | `pk_` | Origin-restricted — requires `WithOrigin` |

## License

[Elastic License 2.0 (ELv2)](LICENSE)
