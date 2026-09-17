package rmdecode

import "fmt"

// EncodeRmPageToSVG converts raw .rm page bytes to a standalone SVG (1404×1872).
// v3/v5: Go stroke renderer. v6: rmc (+ vendored rmscene via PYTHONPATH).
func EncodeRmPageToSVG(data []byte) ([]byte, error) {
	return EncodeRmPageToSVGWithImages(data, nil)
}

// EncodeRmPageToSVGWithImages is EncodeRmPageToSVG with inserted page image blobs.
func EncodeRmPageToSVGWithImages(data []byte, images map[string][]byte) ([]byte, error) {
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
		s, err := RenderWritingsSVG(page)
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	case 6:
		s, err := RenderV6SVGWithRMC(data, images)
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	default:
		return nil, fmt.Errorf("unsupported .rm version %d", ver)
	}
}
