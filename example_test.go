package netloc8_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/netloc8/netloc8-go"
)

func ExampleNewClient() {
	// Create a client with a publishable key and origin.
	client := netloc8.NewClient( "pk_your_key",
		netloc8.WithOrigin( "https://your-app.com" ),
	)

	geo, err := client.LookupIP( context.Background(), "8.8.8.8" )
	if err != nil {
		fmt.Println( "error:", err )
		return
	}

	fmt.Println( geo.CountryCode() )
	fmt.Println( geo.CityName() )
}

func ExampleClient_LookupIP() {
	// Simulate the API with a test server.
	srv := httptest.NewServer( http.HandlerFunc( func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( map[string]any{
			"query":    map[string]any{ "type": "ip", "value": "8.8.8.8", "ipVersion": 4 },
			"location": map[string]any{ "country": map[string]any{ "code": "US" }, "city": "Mountain View" },
			"network":  map[string]any{ "asn": "AS15169", "organization": "Google LLC" },
		})
	}))
	defer srv.Close()

	client := netloc8.NewClient( "pk_test", netloc8.WithBaseURL( srv.URL ) )

	geo, err := client.LookupIP( context.Background(), "8.8.8.8" )
	if err != nil {
		fmt.Println( "error:", err )
		return
	}

	fmt.Println( geo.CountryCode() )
	fmt.Println( geo.CityName() )
	fmt.Println( geo.ASN() )
	// Output:
	// US
	// Mountain View
	// AS15169
}

func ExampleNormalizeIP() {
	fmt.Println( netloc8.NormalizeIP( "::ffff:8.8.8.8" ) )
	fmt.Println( netloc8.NormalizeIP( "[2001:DB8::1]" ) )
	fmt.Println( netloc8.NormalizeIP( "  1.2.3.4  " ) )
	// Output:
	// 8.8.8.8
	// 2001:db8::1
	// 1.2.3.4
}

func ExampleIsPublicIP() {
	fmt.Println( netloc8.IsPublicIP( "8.8.8.8" ) )
	fmt.Println( netloc8.IsPublicIP( "192.168.1.1" ) )
	fmt.Println( netloc8.IsPublicIP( "10.0.0.1" ) )
	fmt.Println( netloc8.IsPublicIP( "100.64.0.1" ) )
	// Output:
	// true
	// false
	// false
	// false
}

func ExampleSubnet() {
	fmt.Println( netloc8.Subnet( "203.0.113.42" ) )
	fmt.Println( netloc8.Subnet( "8.8.8.8" ) )
	fmt.Println( netloc8.Subnet( "2001:db8::1" ) )
	// Output:
	// 203.0.113.0/24
	// 8.8.8.0/24
	//
}

func ExampleIsEU() {
	eu := &netloc8.Geo{
		Location: &netloc8.Location{
			Country: &netloc8.Country{
				Code:   "DE",
				Name:   "Germany",
				Unions: []string{ "EU" },
			},
		},
	}

	nonEU := &netloc8.Geo{
		Location: &netloc8.Location{
			Country: &netloc8.Country{
				Code: "US",
				Name: "United States",
			},
		},
	}

	fmt.Println( netloc8.IsEU( eu ) )
	fmt.Println( netloc8.IsEU( nonEU ) )
	fmt.Println( netloc8.IsEU( nil ) )
	// Output:
	// true
	// false
	// false
}

func ExampleClient_CreateKey() {
	// Simulate the API with a test server.
	srv := httptest.NewServer( http.HandlerFunc( func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		w.WriteHeader( 201 )
		json.NewEncoder( w ).Encode( map[string]any{
			"id":        "key_abc123",
			"prefix":    "sk_live_",
			"name":      "CI Pipeline",
			"type":      "secret",
			"scopes":    []string{ "geo:read" },
			"status":    "active",
			"createdAt": "2024-06-13T12:00:00Z",
			"rawKey":    "sk_live_abc123def456",
		})
	}))
	defer srv.Close()

	client := netloc8.NewClient( "sk_test", netloc8.WithBaseURL( srv.URL ) )

	key, err := client.CreateKey( context.Background(), "CI Pipeline",
		netloc8.WithKeyType( "secret" ),
	)
	if err != nil {
		fmt.Println( "error:", err )
		return
	}

	fmt.Println( key.Name )
	fmt.Println( key.Type )
	fmt.Println( key.RawKey )
	// Output:
	// CI Pipeline
	// secret
	// sk_live_abc123def456
}

func ExampleClient_GetUsage() {
	monthlyCap := 100000
	srv := httptest.NewServer( http.HandlerFunc( func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( map[string]any{
			"totalKeys":     3,
			"activeKeys":    2,
			"totalRequests": 15432,
			"monthlyCap":    monthlyCap,
			"dailyUsage":    []any{},
			"keys":          []any{},
		})
	}))
	defer srv.Close()

	client := netloc8.NewClient( "sk_test", netloc8.WithBaseURL( srv.URL ) )

	usage, err := client.GetUsage( context.Background() )
	if err != nil {
		fmt.Println( "error:", err )
		return
	}

	if usage.MonthlyCap != nil {
		fmt.Printf( "%d / %d requests (%d keys)\n", usage.TotalRequests, *usage.MonthlyCap, usage.TotalKeys )
	} else {
		fmt.Printf( "%d requests, unlimited (%d keys)\n", usage.TotalRequests, usage.TotalKeys )
	}
	// Output:
	// 15432 / 100000 requests (3 keys)
}

func ExampleClient_GetAuditLog() {
	srv := httptest.NewServer( http.HandlerFunc( func( w http.ResponseWriter, r *http.Request ) {
		w.Header().Set( "Content-Type", "application/json" )
		json.NewEncoder( w ).Encode( map[string]any{
			"entries": []map[string]any{
				{
					"id":         "aud_001",
					"action":     "key.created",
					"actorId":    "usr_abc",
					"actorLabel": "tom@example.com",
					"createdAt":  "2024-06-13T12:00:00Z",
				},
			},
			"total": 1,
		})
	}))
	defer srv.Close()

	client := netloc8.NewClient( "sk_test", netloc8.WithBaseURL( srv.URL ) )

	log, err := client.GetAuditLog( context.Background(),
		netloc8.WithLimit( 10 ),
		netloc8.WithAction( "key.created" ),
	)
	if err != nil {
		fmt.Println( "error:", err )
		return
	}

	fmt.Println( log.Total )
	fmt.Println( log.Entries[0].Action )
	fmt.Println( log.Entries[0].ActorLabel )
	// Output:
	// 1
	// key.created
	// tom@example.com
}

