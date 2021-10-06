package common

import "testing"

func Test_sanitized(t *testing.T) {
	tests := []struct {
		name  string
		param string
		want  string
	}{
		{
			name:  "blah",
			param: "./.../..\\./fuu",
			want:  "fuu",
		},
		{
			name:  "blah2",
			param: "./..//../..\\./fuu.\\",
			want:  "fuu",
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitized(tt.param); got != tt.want {
				t.Errorf("sanitized() = %v, want %v", got, tt.want)
			}
		})
	}
}
