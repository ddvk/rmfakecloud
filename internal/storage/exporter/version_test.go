package exporter

import (
	"bytes"
	"strings"
	"testing"
)

func TestDetectRmVersion(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    RmVersion
		wantErr bool
	}{
		{
			name:    "v6 format",
			header:  "reMarkable .lines file, version=6          ", // 43 bytes
			want:    VersionV6,
			wantErr: false,
		},
		{
			name:    "v5 format",
			header:  "reMarkable .lines file, version=5\n",
			want:    VersionV5,
			wantErr: false,
		},
		{
			name:    "v3 format",
			header:  "reMarkable .lines file, version=3\n",
			want:    VersionV3,
			wantErr: false,
		},
		{
			name:    "unknown format",
			header:  "Some other format",
			want:    VersionUnknown,
			wantErr: true,
		},
		{
			name:    "empty input",
			header:  "",
			want:    VersionUnknown,
			wantErr: true,
		},
		{
			name:    "truncated v6 header",
			header:  "reMarkable .lines file, version=6",
			want:    VersionV6,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.header)
			got, err := DetectRmVersion(reader)

			if (err != nil) != tt.wantErr {
				t.Errorf("DetectRmVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("DetectRmVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectRmVersionFromBytes(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    RmVersion
		wantErr bool
	}{
		{
			name:    "v6 bytes",
			data:    []byte("reMarkable .lines file, version=6          "),
			want:    VersionV6,
			wantErr: false,
		},
		{
			name:    "v5 bytes",
			data:    []byte("reMarkable .lines file, version=5\nsomedata"),
			want:    VersionV5,
			wantErr: false,
		},
		{
			name:    "v3 bytes",
			data:    []byte("reMarkable .lines file, version=3\nsomedata"),
			want:    VersionV3,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectRmVersionFromBytes(tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("DetectRmVersionFromBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("DetectRmVersionFromBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRmVersionString(t *testing.T) {
	tests := []struct {
		version RmVersion
		want    string
	}{
		{VersionV3, "v3"},
		{VersionV5, "v5"},
		{VersionV6, "v6"},
		{VersionUnknown, "unknown"},
		{RmVersion(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.version.String(); got != tt.want {
				t.Errorf("RmVersion.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectRmVersionPriorityV6(t *testing.T) {
	// Ensure v6 is detected correctly even with similar v5/v3 strings
	reader := bytes.NewReader([]byte("reMarkable .lines file, version=6          "))
	got, err := DetectRmVersion(reader)

	if err != nil {
		t.Errorf("DetectRmVersion() unexpected error = %v", err)
	}

	if got != VersionV6 {
		t.Errorf("DetectRmVersion() = %v, want %v", got, VersionV6)
	}
}