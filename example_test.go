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
