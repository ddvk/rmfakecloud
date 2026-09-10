package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateAndSaveTheme(t *testing.T) {
	dir := t.TempDir()
	s := newThemeStore(dir)
	xml := []byte(`<?xml version="1.0"?>
<theme id="ocean" name="Ocean" published="false">
  <colors background1="#001" background2="#002" foreground1="#fff" foreground2="#eee" foreground3="#ddd" action="#0af" accept="#0f0" reject="#f00"/>
  <icons brand="cloud" documents="folder" integrations="puzzle" connect="link" screenshare="display" admin="gear" profile="person"/>
  <layout>
    <nav brand-position="start" user-menu="end">
      <item id="documents" visible="true"/>
    </nav>
    <login primary-button="end" passkey-button="start" show-brand="true"/>
  </layout>
</theme>`)
	if err := s.save("ocean", "Ocean", false, xml); err != nil {
		t.Fatal(err)
	}
	list, err := s.list(false)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range list {
		if m.ID == "ocean" {
			t.Fatal("unpublished should be hidden from public list")
		}
	}
	list, err = s.list(true)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range list {
		if m.ID == "ocean" {
			found = true
		}
	}
	if !found {
		t.Fatal("admin list should include ocean")
	}
	if _, err := os.Stat(filepath.Join(dir, "themes", "ocean", "theme.xml")); err != nil {
		t.Fatal(err)
	}
}

func TestBuiltinDefault(t *testing.T) {
	s := newThemeStore(t.TempDir())
	b, builtin, err := s.getXML("default")
	if err != nil || !builtin || len(b) == 0 {
		t.Fatalf("builtin default: err=%v builtin=%v len=%d", err, builtin, len(b))
	}
}

func TestBuiltinPalettesListed(t *testing.T) {
	s := newThemeStore(t.TempDir())
	list, err := s.list(false)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"light", "dark", "system", "hicontrast"}
	got := map[string]bool{}
	order := make([]string, 0, len(list))
	for _, m := range list {
		got[m.ID] = true
		order = append(order, m.ID)
		if m.ID == "default" {
			t.Fatal("default should be hidden when dark exists")
		}
	}
	for _, id := range want {
		if !got[id] {
			t.Fatalf("missing builtin %s in %v", id, order)
		}
		xml, builtin, err := s.getXML(id)
		if err != nil || !builtin || len(xml) == 0 {
			t.Fatalf("%s: err=%v builtin=%v len=%d", id, err, builtin, len(xml))
		}
	}
	if len(order) < 4 || order[0] != "light" || order[1] != "dark" || order[2] != "system" || order[3] != "hicontrast" {
		t.Fatalf("unexpected order %v", order)
	}
}
