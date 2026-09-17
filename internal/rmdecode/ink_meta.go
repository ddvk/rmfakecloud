package rmdecode

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
)

// InkDeviceMeta describes where the reMarkable device viewport sits inside an
// ink PNG that may be larger than 1404×1872 (uncropped strokes / nudges).
type InkDeviceMeta struct {
	OriginX int
	OriginY int
	ViewW   int
	ViewH   int
}

// DefaultInkDeviceMeta is the nominal device canvas with origin at top-left.
func DefaultInkDeviceMeta() InkDeviceMeta {
	return InkDeviceMeta{OriginX: 0, OriginY: 0, ViewW: PageWidthPt, ViewH: PageHeightPt}
}

// ReadInkDeviceMeta reads rmDeviceOrigin / rmDeviceSize PNG tEXt chunks.
// Missing metadata yields DefaultInkDeviceMeta (or origin 0,0 with size from fallback).
func ReadInkDeviceMeta(pngBytes []byte) InkDeviceMeta {
	m := DefaultInkDeviceMeta()
	texts := readPNGTextChunks(pngBytes)
	if v, ok := texts["rmDeviceSize"]; ok {
		if w, h, err := parseTwoInts(v); err == nil && w > 0 && h > 0 {
			m.ViewW, m.ViewH = w, h
		}
	}
	if v, ok := texts["rmDeviceOrigin"]; ok {
		if x, y, err := parseTwoInts(v); err == nil {
			m.OriginX, m.OriginY = x, y
		}
	}
	return m
}

// WriteInkDeviceMeta returns pngBytes with rmDeviceOrigin / rmDeviceSize tEXt set.
func WriteInkDeviceMeta(pngBytes []byte, m InkDeviceMeta) ([]byte, error) {
	if len(pngBytes) < 8 || !bytes.Equal(pngBytes[:8], []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
		return nil, fmt.Errorf("not a png")
	}
	out := make([]byte, 0, len(pngBytes)+128)
	out = append(out, pngBytes[:8]...)
	rest := pngBytes[8:]
	inserted := false
	for len(rest) >= 12 {
		length := int(binary.BigEndian.Uint32(rest[0:4]))
		if length < 0 || 8+length+4 > len(rest) {
			return nil, fmt.Errorf("png chunk truncated")
		}
		ctype := string(rest[4:8])
		chunk := rest[:8+length+4]
		rest = rest[8+length+4:]
		if ctype == "IHDR" {
			out = append(out, chunk...)
			out = append(out, pngTextChunk("rmDeviceOrigin", fmt.Sprintf("%d,%d", m.OriginX, m.OriginY))...)
			out = append(out, pngTextChunk("rmDeviceSize", fmt.Sprintf("%d,%d", m.ViewW, m.ViewH))...)
			inserted = true
			continue
		}
		if ctype == "tEXt" {
			key, _, _ := bytes.Cut(chunk[8:8+length], []byte{0})
			ks := string(key)
			if ks == "rmDeviceOrigin" || ks == "rmDeviceSize" {
				continue // replace below via inserted headers
			}
		}
		out = append(out, chunk...)
		if ctype == "IEND" {
			break
		}
	}
	if !inserted {
		return nil, fmt.Errorf("png missing IHDR")
	}
	return out, nil
}

func parseTwoInts(s string) (int, int, error) {
	parts := strings.Split(strings.TrimSpace(s), ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("want a,b")
	}
	a, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, err
	}
	b, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, err
	}
	return a, b, nil
}

func readPNGTextChunks(pngBytes []byte) map[string]string {
	out := map[string]string{}
	if len(pngBytes) < 8 || !bytes.Equal(pngBytes[:8], []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
		return out
	}
	rest := pngBytes[8:]
	for len(rest) >= 12 {
		length := int(binary.BigEndian.Uint32(rest[0:4]))
		if length < 0 || 8+length+4 > len(rest) {
			break
		}
		ctype := string(rest[4:8])
		data := rest[8 : 8+length]
		rest = rest[8+length+4:]
		if ctype == "tEXt" {
			key, val, ok := bytes.Cut(data, []byte{0})
			if ok {
				out[string(key)] = string(val)
			}
		}
		if ctype == "IEND" {
			break
		}
	}
	return out
}

func pngTextChunk(key, val string) []byte {
	payload := append(append([]byte(key), 0), []byte(val)...)
	length := len(payload)
	chunk := make([]byte, 0, 12+length)
	var hdr [8]byte
	binary.BigEndian.PutUint32(hdr[0:4], uint32(length))
	copy(hdr[4:8], []byte("tEXt"))
	chunk = append(chunk, hdr[:]...)
	chunk = append(chunk, payload...)
	crc := crc32.NewIEEE()
	_, _ = crc.Write(hdr[4:8])
	_, _ = crc.Write(payload)
	var cbuf [4]byte
	binary.BigEndian.PutUint32(cbuf[:], crc.Sum32())
	chunk = append(chunk, cbuf[:]...)
	return chunk
}
