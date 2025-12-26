package server

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// DocumentUri represents a URI as defined in the LSP specification.
type DocumentUri struct {
	Scheme    string
	Authority string
	Path      string
	Query     string
	Fragment  string
}

// ParseURI parses a URI string into a DocumentUri struct.
func ParseURI(uriString string) (DocumentUri, error) {
	u, err := url.Parse(uriString)
	if err != nil {
		return DocumentUri{}, fmt.Errorf("failed to parse URI: %w", err)
	}

	return DocumentUri{
		Scheme:    u.Scheme,
		Authority: u.Host, // net/url.URL.Host contains authority
		Path:      u.Path,
		Query:     u.RawQuery,
		Fragment:  u.Fragment,
	}, nil
}

// String formats the DocumentUri struct back into a valid URI string.
func (d DocumentUri) String() string {
	u := url.URL{
		Scheme:   d.Scheme,
		Host:     d.Authority,
		Path:     d.Path,
		RawQuery: d.Query,
		Fragment: d.Fragment,
	}
	return u.String()
}

// MarshalJSON implements json.Marshaler.
func (d DocumentUri) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *DocumentUri) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	uri, err := ParseURI(s)
	if err != nil {
		return err
	}
	*d = uri
	return nil
}
