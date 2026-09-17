package config

import "testing"

func TestDeriveWebAuthnFromStorageURL(t *testing.T) {
	rpid, origins, ok := deriveWebAuthnFromStorageURL("https://www.example.com:3000")
	if !ok || rpid != "www.example.com" || len(origins) != 1 || origins[0] != "https://www.example.com:3000" {
		t.Fatalf("got rpid=%q origins=%v ok=%v", rpid, origins, ok)
	}
	if _, _, ok := deriveWebAuthnFromStorageURL("http://acorn:3000"); ok {
		t.Fatal("http should not derive")
	}
	if _, _, ok := deriveWebAuthnFromStorageURL("not-a-url"); ok {
		t.Fatal("invalid should not derive")
	}
}

func TestSplitCSV(t *testing.T) {
	got := splitCSV(" https://a.com ,https://b.com, ")
	if len(got) != 2 || got[0] != "https://a.com" || got[1] != "https://b.com" {
		t.Fatalf("%v", got)
	}
}
