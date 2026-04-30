package netloc8

// Geo is the top-level geolocation response from the NetLoc8 API.
// All fields are pointers to distinguish between absent and zero values.
// Fields omitted by the API are nil.
type Geo struct {
	Query    *Query    `json:"query,omitempty"`
	Location *Location `json:"location,omitempty"`
	Network  *Network  `json:"network,omitempty"`
	Sources  *Sources  `json:"sources,omitempty"`
	Meta     *Meta     `json:"meta,omitempty"`
}

// Query describes the IP address that was looked up.
type Query struct {
	// Type is always "ip" for geolocation lookups.
	Type string `json:"type,omitempty"`

	// Value is the resolved IP address (e.g. "8.8.8.8").
	Value string `json:"value,omitempty"`

	// IPVersion is 4 or 6.
	IPVersion int `json:"ipVersion,omitempty"`
}

// Location holds geographic data for an IP address.
type Location struct {
	Continent   *Continent   `json:"continent,omitempty"`
	Country     *Country     `json:"country,omitempty"`
	Region      *Region      `json:"region,omitempty"`
	District    string       `json:"district,omitempty"`
	City        string       `json:"city,omitempty"`
	PostalCode  string       `json:"postalCode,omitempty"`
	Coordinates *Coordinates `json:"coordinates,omitempty"`

	// Timezone is an IANA timezone string (e.g. "America/Los_Angeles").
	Timezone string `json:"timezone,omitempty"`

	// UTCOffset is the current UTC offset (e.g. "-07:00").
	UTCOffset string `json:"utcOffset,omitempty"`

	// GeoConfidence is the confidence score for the geolocation result,
	// ranging from 0.0 (no confidence) to 1.0 (high confidence).
	GeoConfidence float64 `json:"geoConfidence,omitempty"`
}

// Continent identifies a continent.
type Continent struct {
	// Code is a two-letter continent code (e.g. "NA", "EU").
	Code string `json:"code,omitempty"`

	// Name is the full continent name (e.g. "North America").
	Name string `json:"name,omitempty"`
}

// Country identifies a country.
type Country struct {
	// Code is the ISO 3166-1 alpha-2 country code (e.g. "US").
	Code string `json:"code,omitempty"`

	// Name is the full country name (e.g. "United States").
	Name string `json:"name,omitempty"`

	// Flag is a flag emoji (e.g. "🇺🇸").
	Flag string `json:"flag,omitempty"`

	// Unions lists supranational memberships (e.g. ["EU"]).
	// Used by IsEU to detect European Union membership.
	Unions []string `json:"unions,omitempty"`
}

// Region identifies a state, province, or administrative region.
type Region struct {
	// Code is the ISO 3166-2 subdivision code (e.g. "CA").
	Code string `json:"code,omitempty"`

	// Name is the full region name (e.g. "California").
	Name string `json:"name,omitempty"`
}

// Coordinates holds latitude, longitude, and accuracy data.
type Coordinates struct {
	Latitude       float64 `json:"latitude,omitempty"`
	Longitude      float64 `json:"longitude,omitempty"`
	AccuracyRadius int     `json:"accuracyRadius,omitempty"`
}

// Network holds autonomous system and organization data.
type Network struct {
	// ASN is the autonomous system number (e.g. "AS15169").
	ASN string `json:"asn,omitempty"`

	// Organization is the AS organization name (e.g. "Google LLC").
	Organization string `json:"organization,omitempty"`

	// Domain is the organization's primary domain (e.g. "google.com").
	Domain string `json:"domain,omitempty"`
}

// Sources indicates which data providers contributed to the result.
type Sources struct {
	Geo []string `json:"geo,omitempty"`
	ASN []string `json:"asn,omitempty"`
	TZ  []string `json:"tz,omitempty"`
}

// Meta holds response metadata.
type Meta struct {
	// Precision indicates how specific the geolocation data is.
	// One of: "city", "region", "country", "continent", "none".
	Precision string `json:"precision,omitempty"`

	// Tier is the plan tier that served the request (e.g. "free", "pro").
	Tier string `json:"tier,omitempty"`

	// RequestID is a unique identifier for the API request.
	// Present when server-timing is enabled.
	RequestID string `json:"requestId,omitempty"`

	// Degraded is true when the response has been reduced due to
	// plan limits being exceeded.
	Degraded bool `json:"degraded,omitempty"`
}

// CountryCode returns the ISO 3166-1 alpha-2 country code, or empty string.
// Safe to call on nil Geo.
func ( g *Geo ) CountryCode() string {
	if g == nil || g.Location == nil || g.Location.Country == nil {
		return ""
	}
	return g.Location.Country.Code
}

// CountryName returns the full country name, or empty string.
// Safe to call on nil Geo.
func ( g *Geo ) CountryName() string {
	if g == nil || g.Location == nil || g.Location.Country == nil {
		return ""
	}
	return g.Location.Country.Name
}

// RegionName returns the region/state name, or empty string.
// Safe to call on nil Geo.
func ( g *Geo ) RegionName() string {
	if g == nil || g.Location == nil || g.Location.Region == nil {
		return ""
	}
	return g.Location.Region.Name
}

// CityName returns the city name, or empty string.
// Safe to call on nil Geo.
func ( g *Geo ) CityName() string {
	if g == nil || g.Location == nil {
		return ""
	}
	return g.Location.City
}

// TZ returns the IANA timezone string, or empty string.
// Safe to call on nil Geo.
func ( g *Geo ) TZ() string {
	if g == nil || g.Location == nil {
		return ""
	}
	return g.Location.Timezone
}

// IP returns the resolved IP address, or empty string.
// Safe to call on nil Geo.
func ( g *Geo ) IP() string {
	if g == nil || g.Query == nil {
		return ""
	}
	return g.Query.Value
}

// ASN returns the autonomous system number, or empty string.
// Safe to call on nil Geo.
func ( g *Geo ) ASN() string {
	if g == nil || g.Network == nil {
		return ""
	}
	return g.Network.ASN
}

// Org returns the AS organization name, or empty string.
// Safe to call on nil Geo.
func ( g *Geo ) Org() string {
	if g == nil || g.Network == nil {
		return ""
	}
	return g.Network.Organization
}
