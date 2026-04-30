package netloc8

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// IsEU reports whether the geolocation result is in the European Union.
// Checks location.country.unions for "EU" membership.
//
// Safe to call on nil Geo — returns false.
func IsEU( geo *Geo ) bool {
	if geo == nil || geo.Location == nil || geo.Location.Country == nil {
		return false
	}
	for _, u := range geo.Location.Country.Unions {
		if strings.EqualFold( u, "EU" ) {
			return true
		}
	}
	return false
}

// Precision constants for Meta.Precision.
const (
	PrecisionCity      = "city"
	PrecisionRegion    = "region"
	PrecisionCountry   = "country"
	PrecisionContinent = "continent"
	PrecisionNone      = "none"
)

// GetClientIP extracts the most likely real client IP from HTTP
// request headers. Checks standard proxy headers in priority order
// and returns the first public IP found.
//
// Header priority:
//  1. X-Forwarded-For (first public IP in the chain)
//  2. CF-Connecting-IP (Cloudflare)
//  3. True-Client-IP (Akamai/Cloudflare)
//  4. X-Real-IP (nginx)
//  5. X-Client-IP
//  6. Fastly-Client-IP
//  7. Fly-Client-IP
//
// Falls back to the first candidate IP (even private) if no public
// IP is found. Returns empty string if no IP can be determined.
func GetClientIP( r *http.Request ) string {
	// Priority 1: X-Forwarded-For — multi-hop, pick first public IP.
	if xff := r.Header.Get( "X-Forwarded-For" ); xff != "" {
		for _, raw := range strings.Split( xff, "," ) {
			ip := NormalizeIP( strings.TrimSpace( raw ) )
			if ip != "" && IsPublicIP( ip ) {
				return ip
			}
		}
	}

	// Priority 2–7: single-IP headers.
	singleHeaders := []string{
		"CF-Connecting-IP",
		"True-Client-IP",
		"X-Real-IP",
		"X-Client-IP",
		"Fastly-Client-IP",
		"Fly-Client-IP",
	}

	for _, h := range singleHeaders {
		if val := r.Header.Get( h ); val != "" {
			ip := NormalizeIP( val )
			if ip != "" && IsPublicIP( ip ) {
				return ip
			}
		}
	}

	// Fallback: first candidate from XFF even if private.
	if xff := r.Header.Get( "X-Forwarded-For" ); xff != "" {
		for _, raw := range strings.Split( xff, "," ) {
			ip := NormalizeIP( strings.TrimSpace( raw ) )
			if ip != "" {
				return ip
			}
		}
	}

	// Last resort: single headers accepting private.
	for _, h := range singleHeaders {
		if val := r.Header.Get( h ); val != "" {
			ip := NormalizeIP( val )
			if ip != "" {
				return ip
			}
		}
	}

	return ""
}

// GeoFromPlatformHeaders extracts geolocation data from CDN/platform
// request headers without making an API call.
//
// Supports:
//   - Vercel: X-Vercel-IP-Country, X-Vercel-IP-Country-Region,
//     X-Vercel-IP-City, X-Vercel-IP-Latitude, X-Vercel-IP-Longitude,
//     X-Vercel-IP-Timezone
//   - Cloudflare: CF-IPCountry
//   - CloudFront: CloudFront-Viewer-Country
//
// Returns a partial Geo — only fields provided by the platform are set.
// Fields from higher-priority platforms take precedence.
func GeoFromPlatformHeaders( r *http.Request ) *Geo {
	geo := &Geo{}

	// --- Vercel headers ---

	if country := r.Header.Get( "X-Vercel-IP-Country" ); country != "" {
		ensureLocation( geo )
		if geo.Location.Country == nil {
			geo.Location.Country = &Country{}
		}
		geo.Location.Country.Code = country
	}

	if region := r.Header.Get( "X-Vercel-IP-Country-Region" ); region != "" {
		ensureLocation( geo )
		if geo.Location.Region == nil {
			geo.Location.Region = &Region{}
		}
		geo.Location.Region.Code = region
	}

	if city := r.Header.Get( "X-Vercel-IP-City" ); city != "" {
		ensureLocation( geo )
		decoded, err := url.QueryUnescape( city )
		if err != nil {
			decoded = city
		}
		geo.Location.City = decoded
	}

	if lat := r.Header.Get( "X-Vercel-IP-Latitude" ); lat != "" {
		if v, err := strconv.ParseFloat( lat, 64 ); err == nil {
			ensureLocation( geo )
			if geo.Location.Coordinates == nil {
				geo.Location.Coordinates = &Coordinates{}
			}
			geo.Location.Coordinates.Latitude = v
		}
	}

	if lng := r.Header.Get( "X-Vercel-IP-Longitude" ); lng != "" {
		if v, err := strconv.ParseFloat( lng, 64 ); err == nil {
			ensureLocation( geo )
			if geo.Location.Coordinates == nil {
				geo.Location.Coordinates = &Coordinates{}
			}
			geo.Location.Coordinates.Longitude = v
		}
	}

	if tz := r.Header.Get( "X-Vercel-IP-Timezone" ); tz != "" {
		ensureLocation( geo )
		geo.Location.Timezone = tz
	}

	// --- Cloudflare headers (lower priority — only if country not set) ---

	if cf := r.Header.Get( "CF-IPCountry" ); cf != "" {
		ensureLocation( geo )
		if geo.Location.Country == nil {
			geo.Location.Country = &Country{}
		}
		if geo.Location.Country.Code == "" {
			geo.Location.Country.Code = cf
		}
	}

	// --- CloudFront headers (lowest priority) ---

	if cfront := r.Header.Get( "CloudFront-Viewer-Country" ); cfront != "" {
		ensureLocation( geo )
		if geo.Location.Country == nil {
			geo.Location.Country = &Country{}
		}
		if geo.Location.Country.Code == "" {
			geo.Location.Country.Code = cfront
		}
	}

	// Return nil if no headers were found.
	if geo.Location == nil {
		return nil
	}

	return geo
}

// ensureLocation lazily initializes geo.Location.
func ensureLocation( geo *Geo ) {
	if geo.Location == nil {
		geo.Location = &Location{}
	}
}
