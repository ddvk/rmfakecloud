package ui

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientFormFactor(t *testing.T) {
	cases := []struct {
		name string
		ch   string
		ua   string
		want string
	}{
		{"desktop empty", "", "", "desktop"},
		{"ch mobile", "?1", "Mozilla/5.0", "mobile"},
		{"ch desktop", "?0", "Mozilla/5.0 (iPhone)", "desktop"},
		{"ua iphone", "", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)", "mobile"},
		{"ua android", "", "Mozilla/5.0 (Linux; Android 14)", "mobile"},
		{"ua desktop chrome", "", "Mozilla/5.0 (X11; Linux x86_64) Chrome/120.0.0.0", "desktop"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.ch != "" {
				r.Header.Set("Sec-CH-UA-Mobile", tc.ch)
			}
			if tc.ua != "" {
				r.Header.Set("User-Agent", tc.ua)
			}
			if got := clientFormFactor(r); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}
