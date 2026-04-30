package netloc8

import (
	"fmt"
	"net"
	"strings"
)

// NormalizeIP cleans an IP string: strips IPv4-mapped IPv6 prefixes,
// brackets, and lowercases the result.
//
// Examples:
//
//	NormalizeIP( "::ffff:8.8.8.8" )  → "8.8.8.8"
//	NormalizeIP( "[2001:db8::1]" )   → "2001:db8::1"
//	NormalizeIP( "  8.8.8.8  " )     → "8.8.8.8"
func NormalizeIP( ip string ) string {
	s := strings.TrimSpace( ip )
	if s == "" {
		return ""
	}

	// Strip brackets (IPv6 in URLs).
	if strings.HasPrefix( s, "[" ) && strings.HasSuffix( s, "]" ) {
		s = s[1 : len( s )-1]
	}

	// Strip IPv4-mapped IPv6 prefix.
	lower := strings.ToLower( s )
	if strings.HasPrefix( lower, "::ffff:" ) {
		s = s[7:]
	}

	return strings.ToLower( s )
}

// IsPublicIP reports whether the given IP address is publicly routable.
// Returns false for RFC1918, loopback, link-local, CGNAT, and ULA addresses.
//
// Examples:
//
//	IsPublicIP( "8.8.8.8" )       → true
//	IsPublicIP( "192.168.1.1" )   → false
//	IsPublicIP( "10.0.0.1" )      → false
//	IsPublicIP( "100.64.0.1" )    → false  (CGNAT)
//	IsPublicIP( "::1" )           → false
//	IsPublicIP( "fd00::1" )       → false  (ULA)
func IsPublicIP( ip string ) bool {
	normalized := NormalizeIP( ip )
	if normalized == "" {
		return false
	}

	parsed := net.ParseIP( normalized )
	if parsed == nil {
		return false
	}

	// IPv4.
	if ipv4 := parsed.To4(); ipv4 != nil {
		a, b := ipv4[0], ipv4[1]

		// 0.0.0.0
		if ipv4.Equal( net.IPv4zero ) {
			return false
		}

		// Loopback: 127.0.0.0/8
		if a == 127 {
			return false
		}

		// RFC1918: 10.0.0.0/8
		if a == 10 {
			return false
		}

		// RFC1918: 172.16.0.0/12
		if a == 172 && b >= 16 && b <= 31 {
			return false
		}

		// RFC1918: 192.168.0.0/16
		if a == 192 && b == 168 {
			return false
		}

		// CGNAT: 100.64.0.0/10
		if a == 100 && b >= 64 && b <= 127 {
			return false
		}

		// Link-local: 169.254.0.0/16
		if a == 169 && b == 254 {
			return false
		}

		return true
	}

	// IPv6.

	// Loopback: ::1
	if parsed.Equal( net.IPv6loopback ) {
		return false
	}

	// Unspecified: ::
	if parsed.Equal( net.IPv6unspecified ) {
		return false
	}

	// ULA: fc00::/7
	if normalized[0] == 'f' && ( normalized[1] == 'c' || normalized[1] == 'd' ) {
		return false
	}

	// Link-local: fe80::/10
	if strings.HasPrefix( normalized, "fe80" ) {
		return false
	}

	return true
}

// Subnet derives the /24 CIDR prefix from an IPv4 address.
// Returns an empty string for IPv6 or invalid addresses.
//
// Examples:
//
//	Subnet( "203.0.113.42" )  → "203.0.113.0/24"
//	Subnet( "8.8.8.8" )      → "8.8.8.0/24"
//	Subnet( "2001:db8::1" )  → ""
func Subnet( ip string ) string {
	normalized := NormalizeIP( ip )
	if normalized == "" {
		return ""
	}

	parsed := net.ParseIP( normalized )
	if parsed == nil {
		return ""
	}

	ipv4 := parsed.To4()
	if ipv4 == nil {
		return "" // IPv6 — no /24 subnet
	}

	return fmt.Sprintf( "%d.%d.%d.0/24", ipv4[0], ipv4[1], ipv4[2] )
}

// ParseIP parses and validates an IP address string.
// Returns the normalized IP string and true if valid, or empty string
// and false if invalid.
func ParseIP( ip string ) ( string, bool ) {
	normalized := NormalizeIP( ip )
	if normalized == "" {
		return "", false
	}

	parsed := net.ParseIP( normalized )
	if parsed == nil {
		return "", false
	}

	return normalized, true
}

// IsIPv4 reports whether the given string is a valid IPv4 address.
func IsIPv4( ip string ) bool {
	normalized := NormalizeIP( ip )
	if normalized == "" {
		return false
	}

	parsed := net.ParseIP( normalized )
	if parsed == nil {
		return false
	}

	return parsed.To4() != nil
}

// IsIPv6 reports whether the given string is a valid IPv6 address.
func IsIPv6( ip string ) bool {
	normalized := NormalizeIP( ip )
	if normalized == "" {
		return false
	}

	parsed := net.ParseIP( normalized )
	if parsed == nil {
		return false
	}

	return parsed.To4() == nil
}
