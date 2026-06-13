package netloc8

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestGeo_JSONRoundtrip( t *testing.T ) {
	original := &Geo{
		Query: &Query{
			Type:      "ip",
			Value:     "8.8.8.8",
			IPVersion: 4,
		},
		Location: &Location{
			Continent: &Continent{ Code: "NA", Name: "North America" },
			Country: &Country{
				Code:   "US",
				Name:   "United States",
				Flag:   "🇺🇸",
				Unions: []string{},
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
		},
	}

	data, err := json.Marshal( original )
	if err != nil {
		t.Fatalf( "Marshal: %v", err )
	}

	var decoded Geo
	if err := json.Unmarshal( data, &decoded ); err != nil {
		t.Fatalf( "Unmarshal: %v", err )
	}

	if decoded.CountryCode() != "US" {
		t.Errorf( "CountryCode() = %q, want US", decoded.CountryCode() )
	}
	if decoded.CityName() != "Mountain View" {
		t.Errorf( "CityName() = %q, want Mountain View", decoded.CityName() )
	}
	if decoded.TZ() != "America/Los_Angeles" {
		t.Errorf( "TZ() = %q, want America/Los_Angeles", decoded.TZ() )
	}
	if decoded.ASN() != "AS15169" {
		t.Errorf( "ASN() = %q, want AS15169", decoded.ASN() )
	}
	if decoded.Org() != "Google LLC" {
		t.Errorf( "Org() = %q, want Google LLC", decoded.Org() )
	}
	if decoded.IP() != "8.8.8.8" {
		t.Errorf( "IP() = %q, want 8.8.8.8", decoded.IP() )
	}
	if decoded.RegionName() != "California" {
		t.Errorf( "RegionName() = %q, want California", decoded.RegionName() )
	}
	if decoded.CountryName() != "United States" {
		t.Errorf( "CountryName() = %q, want United States", decoded.CountryName() )
	}
	if decoded.RegionCode() != "CA" {
		t.Errorf( "RegionCode() = %q, want CA", decoded.RegionCode() )
	}
	if decoded.Lat() != 37.386 {
		t.Errorf( "Lat() = %f, want 37.386", decoded.Lat() )
	}
	if decoded.Lng() != -122.084 {
		t.Errorf( "Lng() = %f, want -122.084", decoded.Lng() )
	}
	if !decoded.HasCoordinates() {
		t.Error( "HasCoordinates() should be true for full geo" )
	}
}

func TestGeo_NilSafety( t *testing.T ) {
	// All accessors should return zero values on nil Geo.
	var g *Geo
	if g.CountryCode() != "" {
		t.Error( "nil Geo.CountryCode() should be empty" )
	}
	if g.CountryName() != "" {
		t.Error( "nil Geo.CountryName() should be empty" )
	}
	if g.RegionName() != "" {
		t.Error( "nil Geo.RegionName() should be empty" )
	}
	if g.CityName() != "" {
		t.Error( "nil Geo.CityName() should be empty" )
	}
	if g.TZ() != "" {
		t.Error( "nil Geo.TZ() should be empty" )
	}
	if g.IP() != "" {
		t.Error( "nil Geo.IP() should be empty" )
	}
	if g.ASN() != "" {
		t.Error( "nil Geo.ASN() should be empty" )
	}
	if g.Org() != "" {
		t.Error( "nil Geo.Org() should be empty" )
	}
	if g.RegionCode() != "" {
		t.Error( "nil Geo.RegionCode() should be empty" )
	}
	if g.Lat() != 0 {
		t.Error( "nil Geo.Lat() should be 0" )
	}
	if g.Lng() != 0 {
		t.Error( "nil Geo.Lng() should be 0" )
	}
	if g.HasCoordinates() {
		t.Error( "nil Geo.HasCoordinates() should be false" )
	}

	// Partial Geo with nil sub-structs.
	partial := &Geo{}
	if partial.CountryCode() != "" {
		t.Error( "empty Geo.CountryCode() should be empty" )
	}
	if partial.CityName() != "" {
		t.Error( "empty Geo.CityName() should be empty" )
	}
	if partial.RegionCode() != "" {
		t.Error( "empty Geo.RegionCode() should be empty" )
	}
	if partial.Lat() != 0 {
		t.Error( "empty Geo.Lat() should be 0" )
	}
	if partial.Lng() != 0 {
		t.Error( "empty Geo.Lng() should be 0" )
	}
	if partial.HasCoordinates() {
		t.Error( "empty Geo.HasCoordinates() should be false" )
	}
}

func TestGeo_ParseAPIResponse( t *testing.T ) {
	// Simulate a raw API response.
	raw := `{
		"query": { "type": "ip", "value": "8.8.8.8", "ipVersion": 4 },
		"location": {
			"continent": { "code": "NA", "name": "North America" },
			"country": { "code": "US", "name": "United States", "flag": "🇺🇸" },
			"region": { "code": "CA", "name": "California" },
			"city": "Mountain View",
			"postalCode": "94043",
			"coordinates": { "latitude": 37.386, "longitude": -122.084, "accuracyRadius": 621 },
			"timezone": "America/Los_Angeles",
			"utcOffset": "-07:00",
			"geoConfidence": 1
		},
		"network": { "asn": "AS15169", "organization": "Google LLC", "domain": "google.com" },
		"sources": { "geo": ["dbip", "ip2location"], "asn": ["ipinfo"], "tz": ["derived"] },
		"meta": { "precision": "city", "tier": "pro", "requestId": "abc-123" }
	}`

	var geo Geo
	if err := json.Unmarshal( []byte( raw ), &geo ); err != nil {
		t.Fatalf( "Unmarshal: %v", err )
	}

	if geo.Meta.RequestID != "abc-123" {
		t.Errorf( "RequestID = %q, want abc-123", geo.Meta.RequestID )
	}
	if geo.Location.Coordinates.AccuracyRadius != 621 {
		t.Errorf( "AccuracyRadius = %d, want 621", geo.Location.Coordinates.AccuracyRadius )
	}
	if len( geo.Sources.Geo ) != 2 {
		t.Errorf( "Sources.Geo length = %d, want 2", len( geo.Sources.Geo ) )
	}
}

func TestIsEU( t *testing.T ) {
	tests := []struct {
		name string
		geo  *Geo
		want bool
	}{
		{
			"EU member",
			&Geo{ Location: &Location{ Country: &Country{ Code: "DE", Unions: []string{ "EU" } } } },
			true,
		},
		{
			"EU member lowercase",
			&Geo{ Location: &Location{ Country: &Country{ Code: "FR", Unions: []string{ "eu" } } } },
			true,
		},
		{
			"non-EU",
			&Geo{ Location: &Location{ Country: &Country{ Code: "US" } } },
			false,
		},
		{
			"no unions field",
			&Geo{ Location: &Location{ Country: &Country{ Code: "GB" } } },
			false,
		},
		{
			"nil geo",
			nil,
			false,
		},
		{
			"empty geo",
			&Geo{},
			false,
		},
	}

	for _, tt := range tests {
		t.Run( tt.name, func( t *testing.T ) {
			got := IsEU( tt.geo )
			if got != tt.want {
				t.Errorf( "IsEU() = %v, want %v", got, tt.want )
			}
		})
	}
}

func TestGetClientIP( t *testing.T ) {
	tests := []struct {
		name    string
		headers map[string]string
		want    string
	}{
		{
			"XFF single public",
			map[string]string{ "X-Forwarded-For": "203.0.113.42" },
			"203.0.113.42",
		},
		{
			"XFF chain — first public",
			map[string]string{ "X-Forwarded-For": "10.0.0.1, 203.0.113.42, 198.51.100.1" },
			"203.0.113.42",
		},
		{
			"CF-Connecting-IP",
			map[string]string{ "CF-Connecting-IP": "8.8.8.8" },
			"8.8.8.8",
		},
		{
			"X-Real-IP",
			map[string]string{ "X-Real-IP": "1.1.1.1" },
			"1.1.1.1",
		},
		{
			"all private — fallback to first XFF",
			map[string]string{ "X-Forwarded-For": "192.168.1.1, 10.0.0.1" },
			"192.168.1.1",
		},
		{
			"no headers",
			map[string]string{},
			"",
		},
	}

	for _, tt := range tests {
		t.Run( tt.name, func( t *testing.T ) {
			r := &http.Request{ Header: http.Header{} }
			for k, v := range tt.headers {
				r.Header.Set( k, v )
			}
			got := GetClientIP( r )
			if got != tt.want {
				t.Errorf( "GetClientIP() = %q, want %q", got, tt.want )
			}
		})
	}
}

func TestGeoFromPlatformHeaders( t *testing.T ) {
	t.Run( "Vercel full headers", func( t *testing.T ) {
		r := &http.Request{ Header: http.Header{} }
		r.Header.Set( "X-Vercel-IP-Country", "US" )
		r.Header.Set( "X-Vercel-IP-Country-Region", "CA" )
		r.Header.Set( "X-Vercel-IP-City", "Mountain View" )
		r.Header.Set( "X-Vercel-IP-Latitude", "37.386" )
		r.Header.Set( "X-Vercel-IP-Longitude", "-122.084" )
		r.Header.Set( "X-Vercel-IP-Timezone", "America/Los_Angeles" )

		geo := GeoFromPlatformHeaders( r )
		if geo == nil {
			t.Fatal( "expected non-nil Geo" )
		}
		if geo.CountryCode() != "US" {
			t.Errorf( "CountryCode() = %q, want US", geo.CountryCode() )
		}
		if geo.RegionCode() != "CA" {
			t.Errorf( "RegionCode() = %q, want CA", geo.RegionCode() )
		}
		if geo.CityName() != "Mountain View" {
			t.Errorf( "CityName() = %q, want Mountain View", geo.CityName() )
		}
		if geo.TZ() != "America/Los_Angeles" {
			t.Errorf( "TZ() = %q, want America/Los_Angeles", geo.TZ() )
		}
		if geo.Lat() != 37.386 {
			t.Errorf( "Lat() = %f, want 37.386", geo.Lat() )
		}
		if geo.Lng() != -122.084 {
			t.Errorf( "Lng() = %f, want -122.084", geo.Lng() )
		}
	})

	t.Run( "Cloudflare country only", func( t *testing.T ) {
		r := &http.Request{ Header: http.Header{} }
		r.Header.Set( "CF-IPCountry", "DE" )

		geo := GeoFromPlatformHeaders( r )
		if geo == nil {
			t.Fatal( "expected non-nil Geo" )
		}
		if geo.CountryCode() != "DE" {
			t.Errorf( "CountryCode() = %q, want DE", geo.CountryCode() )
		}
	})

	t.Run( "CloudFront country only", func( t *testing.T ) {
		r := &http.Request{ Header: http.Header{} }
		r.Header.Set( "CloudFront-Viewer-Country", "JP" )

		geo := GeoFromPlatformHeaders( r )
		if geo == nil {
			t.Fatal( "expected non-nil Geo" )
		}
		if geo.CountryCode() != "JP" {
			t.Errorf( "CountryCode() = %q, want JP", geo.CountryCode() )
		}
	})

	t.Run( "Vercel takes priority over Cloudflare", func( t *testing.T ) {
		r := &http.Request{ Header: http.Header{} }
		r.Header.Set( "X-Vercel-IP-Country", "US" )
		r.Header.Set( "CF-IPCountry", "DE" )

		geo := GeoFromPlatformHeaders( r )
		if geo == nil {
			t.Fatal( "expected non-nil Geo" )
		}
		if geo.CountryCode() != "US" {
			t.Errorf( "CountryCode() = %q, want US (Vercel priority)", geo.CountryCode() )
		}
	})

	t.Run( "URL-encoded city", func( t *testing.T ) {
		r := &http.Request{ Header: http.Header{} }
		r.Header.Set( "X-Vercel-IP-City", "S%C3%A3o+Paulo" )

		geo := GeoFromPlatformHeaders( r )
		if geo == nil {
			t.Fatal( "expected non-nil Geo" )
		}
		if geo.CityName() != "São Paulo" {
			t.Errorf( "CityName() = %q, want São Paulo", geo.CityName() )
		}
	})

	t.Run( "no headers returns nil", func( t *testing.T ) {
		r := &http.Request{ Header: http.Header{} }

		geo := GeoFromPlatformHeaders( r )
		if geo != nil {
			t.Errorf( "expected nil Geo, got %+v", geo )
		}
	})
}
