package fs

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
)

var pngMagic = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}

// validRasterImage reports whether b looks like a usable page image (PNG/JPEG/GIF/WebP).
// Used for sibling page images and device thumbs before serving or staging for rmc.
func validRasterImage(b []byte) bool {
	if len(b) < 12 {
		return false
	}
	if _, _, err := image.DecodeConfig(bytes.NewReader(b)); err == nil {
		return true
	}
	if bytes.HasPrefix(b, pngMagic) {
		if _, err := png.DecodeConfig(bytes.NewReader(b)); err == nil {
			return true
		}
	}
	// WebP: RIFF....WEBP (stdlib cannot DecodeConfig without x/image/webp).
	if bytes.Equal(b[0:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP")) {
		return true
	}
	return false
}

// filterValidPageImages drops undecodable blobs so rmc staging only sees real images.
func filterValidPageImages(in map[string][]byte) map[string][]byte {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string][]byte, len(in))
	for name, b := range in {
		if validRasterImage(b) {
			out[name] = b
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
