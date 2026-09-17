package epub

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestFindCoverImagePath(t *testing.T) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	w, err := zw.Create("OEBPS/cover.xhtml")
	must(err)
	_, err = w.Write([]byte(`<?xml version="1.0"?><html xmlns="http://www.w3.org/1999/xhtml"><body><img src="images/c.jpg"/></body></html>`))
	must(err)
	w, err = zw.Create("OEBPS/images/c.jpg")
	must(err)
	_, err = w.Write([]byte{0xff, 0xd8, 0xff, 0xe0}) // fake JPEG header
	must(err)
	must(zw.Close())

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	must(err)
	p, err := FindCoverImagePath(zr)
	must(err)
	if p != "OEBPS/images/c.jpg" {
		t.Fatalf("got %q", p)
	}
}

func TestFindFirstImageInFile(t *testing.T) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	w, err := zw.Create("OEBPS/ch1.xhtml")
	must(err)
	_, err = w.Write([]byte(`<html><body><p>x</p><img src="images/p.png"/></body></html>`))
	must(err)
	w, err = zw.Create("OEBPS/images/p.png")
	must(err)
	_, err = w.Write([]byte{0x89, 0x50, 0x4e, 0x47})
	must(err)
	must(zw.Close())

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	must(err)
	p, err := FindFirstImageInFile(zr, "OEBPS/ch1.xhtml")
	must(err)
	if p != "OEBPS/images/p.png" {
		t.Fatalf("got %q", p)
	}
}

func TestFindThumbImagePathNamedCover(t *testing.T) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	w, err := zw.Create("OEBPS/images/cover.jpg")
	must(err)
	_, err = w.Write(bytes.Repeat([]byte{0xff, 0xd8, 0xff, 0xe0}, 16))
	must(err)
	must(zw.Close())

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	must(err)
	p, err := FindThumbImagePath(zr)
	must(err)
	if p != "OEBPS/images/cover.jpg" {
		t.Fatalf("got %q", p)
	}
}

func TestFindThumbImagePathOPFProperties(t *testing.T) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	w, err := zw.Create("META-INF/container.xml")
	must(err)
	_, err = w.Write([]byte(`<?xml version="1.0"?><container><rootfiles><rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/></rootfiles></container>`))
	must(err)
	w, err = zw.Create("OEBPS/content.opf")
	must(err)
	_, err = w.Write([]byte(`<?xml version="1.0"?><package xmlns="http://www.idpf.org/2007/opf"><manifest><item id="ch" href="ch.xhtml" media-type="application/xhtml+xml"/><item id="cov" href="images/art.png" media-type="image/png" properties="cover-image"/></manifest><spine><itemref idref="ch"/></spine></package>`))
	must(err)
	w, err = zw.Create("OEBPS/ch.xhtml")
	must(err)
	_, err = w.Write([]byte(`<html><body><p>hi</p></body></html>`))
	must(err)
	w, err = zw.Create("OEBPS/images/art.png")
	must(err)
	_, err = w.Write(bytes.Repeat([]byte{0x89, 0x50, 0x4e, 0x47}, 20))
	must(err)
	must(zw.Close())

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	must(err)
	p, err := FindThumbImagePath(zr)
	must(err)
	if p != "OEBPS/images/art.png" {
		t.Fatalf("got %q", p)
	}
}
