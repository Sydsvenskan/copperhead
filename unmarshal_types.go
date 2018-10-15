package copperhead

import (
	"net/url"
	"time"
)

// URL is an TextUnmarshaler-aware URL
type URL struct {
	url.URL
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (u *URL) UnmarshalText(text []byte) error {
	var u2 url.URL
	if err := u2.UnmarshalBinary(text); err != nil {
		return err
	}
	u.URL = u2
	return nil
}

// MustParseURL is a helper function for setting configuration
// defaults. Panics if the passed url is invalid.
func MustParseURL(rawURL string) *URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return &URL{URL: *u}
}

// Time is an TextUnmarshaler-aware Time
type Time struct {
	time.Time
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (t *Time) UnmarshalText(text []byte) error {
	pt, err := time.Parse(time.RFC3339, string(text))
	if err != nil {
		return err
	}

	t.Time = pt

	return nil
}

// Duration is an TextUnmarshaler-aware Duration
type Duration struct {
	time.Duration
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (d *Duration) UnmarshalText(text []byte) error {
	pd, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}

	d.Duration = pd

	return nil
}
