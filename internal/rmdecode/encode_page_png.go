package rmdecode

import "fmt"

// EncodeRmPageToPNG converts raw .rm page bytes to PNG (1404×1872)
// using the Go stroke renderer (v3/v5).
func EncodeRmPageToPNG(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty .rm data")
	}
	ver, err := ParseVersion(data)
	if err != nil {
		return nil, err
	}
	switch ver {
	case 3, 5:
		page, err := DecodeLegacy(data)
		if err != nil {
			return nil, err
		}
		return RenderWritingsPNG(page)
	default:
		return nil, fmt.Errorf("unsupported .rm version %d", ver)
	}
}
