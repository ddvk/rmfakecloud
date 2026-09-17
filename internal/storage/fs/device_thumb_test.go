package fs

import "testing"

func TestMatchDeviceThumb(t *testing.T) {
	page := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	cases := []struct {
		name string
		want bool
	}{
		{page + ".jpg", true},
		{page + ".png", true},
		{"thumbnails/" + page + ".jpg", true},
		{page + ".rm", false},
		{page + ".content", false},
		{"other.jpg", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := matchDeviceThumb(tc.name, page); got != tc.want {
			t.Fatalf("%q: got %v want %v", tc.name, got, tc.want)
		}
	}
}
