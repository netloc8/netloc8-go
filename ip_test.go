package netloc8

import (
	"testing"
)

func TestNormalizeIP( t *testing.T ) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{ "plain IPv4", "8.8.8.8", "8.8.8.8" },
		{ "whitespace", "  8.8.8.8  ", "8.8.8.8" },
		{ "IPv4-mapped IPv6", "::ffff:8.8.8.8", "8.8.8.8" },
		{ "IPv4-mapped uppercase", "::FFFF:1.2.3.4", "1.2.3.4" },
		{ "bracketed IPv6", "[2001:db8::1]", "2001:db8::1" },
		{ "plain IPv6", "2001:DB8::1", "2001:db8::1" },
		{ "empty", "", "" },
		{ "whitespace only", "   ", "" },
	}

	for _, tt := range tests {
		t.Run( tt.name, func( t *testing.T ) {
			got := NormalizeIP( tt.input )
			if got != tt.want {
				t.Errorf( "NormalizeIP(%q) = %q, want %q", tt.input, got, tt.want )
			}
		})
	}
}

func TestIsPublicIP( t *testing.T ) {
	tests := []struct {
		ip   string
		want bool
	}{
		// Public.
		{ "8.8.8.8", true },
		{ "1.1.1.1", true },
		{ "203.0.113.42", true },
		{ "2607:f8b0:4004:800::200e", true },

		// Loopback.
		{ "127.0.0.1", false },
		{ "127.255.255.255", false },
		{ "::1", false },

		// RFC1918.
		{ "10.0.0.1", false },
		{ "10.255.255.255", false },
		{ "172.16.0.1", false },
		{ "172.31.255.255", false },
		{ "172.15.0.1", true },  // Just outside /12.
		{ "172.32.0.1", true },  // Just outside /12.
		{ "192.168.0.1", false },
		{ "192.168.255.255", false },

		// CGNAT.
		{ "100.64.0.1", false },
		{ "100.127.255.255", false },
		{ "100.63.255.255", true },  // Just outside /10.
		{ "100.128.0.1", true },     // Just outside /10.

		// Link-local.
		{ "169.254.0.1", false },
		{ "169.254.255.255", false },
		{ "fe80::1", false },
		{ "fe90::1", false },          // fe80::/10 covers fe80–febf.
		{ "fea0::1", false },          // fe80::/10 covers fe80–febf.
		{ "febf::1", false },          // fe80::/10 upper bound.
		{ "fec0::1", true },           // Just outside fe80::/10.

		// ULA.
		{ "fc00::1", false },
		{ "fd12:3456::1", false },

		// Unspecified.
		{ "0.0.0.0", false },
		{ "::", false },

		// Invalid.
		{ "not-an-ip", false },
		{ "", false },
	}

	for _, tt := range tests {
		t.Run( tt.ip, func( t *testing.T ) {
			got := IsPublicIP( tt.ip )
			if got != tt.want {
				t.Errorf( "IsPublicIP(%q) = %v, want %v", tt.ip, got, tt.want )
			}
		})
	}
}

func TestSubnet( t *testing.T ) {
	tests := []struct {
		ip   string
		want string
	}{
		{ "203.0.113.42", "203.0.113.0/24" },
		{ "8.8.8.8", "8.8.8.0/24" },
		{ "192.168.1.100", "192.168.1.0/24" },
		{ "10.0.0.1", "10.0.0.0/24" },
		{ "2001:db8::1", "" }, // IPv6 — no /24.
		{ "not-an-ip", "" },
		{ "", "" },
	}

	for _, tt := range tests {
		t.Run( tt.ip, func( t *testing.T ) {
			got := Subnet( tt.ip )
			if got != tt.want {
				t.Errorf( "Subnet(%q) = %q, want %q", tt.ip, got, tt.want )
			}
		})
	}
}

func TestParseIP( t *testing.T ) {
	tests := []struct {
		input   string
		wantIP  string
		wantOK  bool
	}{
		{ "8.8.8.8", "8.8.8.8", true },
		{ "::ffff:1.2.3.4", "1.2.3.4", true },
		{ "2001:db8::1", "2001:db8::1", true },
		{ "not-an-ip", "", false },
		{ "", "", false },
	}

	for _, tt := range tests {
		t.Run( tt.input, func( t *testing.T ) {
			gotIP, gotOK := ParseIP( tt.input )
			if gotIP != tt.wantIP || gotOK != tt.wantOK {
				t.Errorf( "ParseIP(%q) = (%q, %v), want (%q, %v)",
					tt.input, gotIP, gotOK, tt.wantIP, tt.wantOK )
			}
		})
	}
}

func TestIsIPv4( t *testing.T ) {
	tests := []struct {
		ip   string
		want bool
	}{
		{ "8.8.8.8", true },
		{ "::ffff:8.8.8.8", true }, // Mapped → still IPv4.
		{ "2001:db8::1", false },
		{ "not-an-ip", false },
		{ "", false },
	}

	for _, tt := range tests {
		t.Run( tt.ip, func( t *testing.T ) {
			got := IsIPv4( tt.ip )
			if got != tt.want {
				t.Errorf( "IsIPv4(%q) = %v, want %v", tt.ip, got, tt.want )
			}
		})
	}
}

func TestIsIPv6( t *testing.T ) {
	tests := []struct {
		ip   string
		want bool
	}{
		{ "2001:db8::1", true },
		{ "::1", true },
		{ "8.8.8.8", false },
		{ "::ffff:8.8.8.8", false }, // Mapped → treated as IPv4.
		{ "not-an-ip", false },
		{ "", false },
	}

	for _, tt := range tests {
		t.Run( tt.ip, func( t *testing.T ) {
			got := IsIPv6( tt.ip )
			if got != tt.want {
				t.Errorf( "IsIPv6(%q) = %v, want %v", tt.ip, got, tt.want )
			}
		})
	}
}
