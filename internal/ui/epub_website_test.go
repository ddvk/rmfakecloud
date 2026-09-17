package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestInjectHTMLBase(t *testing.T) {
	in := []byte(`<!DOCTYPE html><html lang="en"><head><title>Ch</title></head><body><p>Hi</p></body></html>`)
	out := injectHTMLBase(in, "/ui/api/documents/d/epub/OEBPS/")
	if !bytes.Contains(out, []byte(`<base href="/ui/api/documents/d/epub/OEBPS/">`)) {
		t.Fatalf("missing base: %s", out)
	}
	if bytes.Count(bytes.ToLower(out), []byte("<base")) != 1 {
		t.Fatal("expected a single base tag")
	}
	again := injectHTMLBase(out, "/other/")
	if !bytes.Equal(out, again) {
		t.Fatal("should not insert a second base")
	}
}

func TestInjectHTMLBaseCreatesHead(t *testing.T) {
	in := []byte(`<html><body>x</body></html>`)
	out := injectHTMLBase(in, "/epub/")
	if !bytes.Contains(out, []byte(`<head><base href="/epub/"></head>`)) {
		t.Fatalf("expected wrapped head: %s", out)
	}
}

func TestRewriteRootRelative(t *testing.T) {
	in := []byte(`<img src="/images/c.jpg"/><a href="/ch2.xhtml">n</a><a href="https://ex">e</a>`)
	out := rewriteRootRelative(in, "/ui/api/documents/d/epub")
	s := string(out)
	if !strings.Contains(s, `src="/ui/api/documents/d/epub/images/c.jpg"`) {
		t.Fatalf("img: %s", s)
	}
	if !strings.Contains(s, `href="/ui/api/documents/d/epub/ch2.xhtml"`) {
		t.Fatalf("a: %s", s)
	}
	if !strings.Contains(s, `href="https://ex"`) {
		t.Fatal("should leave absolute URLs")
	}
	proto := rewriteRootRelative([]byte(`<script src="//cdn.example/x.js"></script>`), "/epub")
	if !strings.Contains(string(proto), `src="//cdn.example/x.js"`) {
		t.Fatalf("protocol-relative: %s", proto)
	}
}

func TestEpubAssetURL(t *testing.T) {
	got := epubAssetURL("doc id", "OEBPS/My File.xhtml")
	want := "/ui/api/documents/doc%20id/epub/OEBPS/My%20File.xhtml"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestEpubSpineLabel(t *testing.T) {
	if g := epubSpineLabel("OEBPS/part0001.xhtml", 0); g != "part0001" {
		t.Fatalf("got %q", g)
	}
}
