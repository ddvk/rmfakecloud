package ui

import (
	"embed"
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

//go:embed themes/*
var builtinThemesFS embed.FS

var themeIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

// ThemeMeta is list metadata for a shell theme.
type ThemeMeta struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Published bool   `json:"published"`
	Builtin   bool   `json:"builtin"`
}

type themeXMLRoot struct {
	XMLName   xml.Name `xml:"theme"`
	ID        string   `xml:"id,attr"`
	Name      string   `xml:"name,attr"`
	Published string   `xml:"published,attr"`
}

type themeStore struct {
	dir string
	mu  sync.RWMutex
}

func newThemeStore(dataDir string) *themeStore {
	dir := filepath.Join(dataDir, "themes")
	_ = os.MkdirAll(dir, 0o755)
	return &themeStore{dir: dir}
}

func (s *themeStore) list(includeUnpublished bool) ([]ThemeMeta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	byID := map[string]ThemeMeta{}

	for _, id := range s.builtinIDs() {
		meta, err := s.readBuiltinMeta(id)
		if err != nil {
			continue
		}
		byID[meta.ID] = meta
	}

	entries, err := os.ReadDir(s.dir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		if !themeIDPattern.MatchString(id) {
			continue
		}
		meta, err := s.readDiskMeta(id)
		if err != nil {
			continue
		}
		byID[id] = meta
	}

	_, hasDark := byID["dark"]
	out := make([]ThemeMeta, 0, len(byID))
	for _, m := range byID {
		if !includeUnpublished && !m.Published {
			continue
		}
		if m.ID == "default" && hasDark {
			continue
		}
		out = append(out, m)
	}
	sortThemes(out)
	return out, nil
}

func (s *themeStore) builtinIDs() []string {
	entries, err := fs.ReadDir(builtinThemesFS, "themes")
	if err != nil {
		return []string{"default"}
	}
	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".xml") {
			continue
		}
		id := strings.TrimSuffix(name, ".xml")
		if themeIDPattern.MatchString(id) {
			ids = append(ids, id)
		}
	}
	return ids
}

func sortThemes(out []ThemeMeta) {
	rank := map[string]int{"light": 0, "dark": 1, "system": 2, "hicontrast": 3}
	sort.Slice(out, func(i, j int) bool {
		ri, iok := rank[out[i].ID]
		rj, jok := rank[out[j].ID]
		if iok && jok {
			return ri < rj
		}
		if iok {
			return true
		}
		if jok {
			return false
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
}

func (s *themeStore) getXML(id string) ([]byte, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !themeIDPattern.MatchString(id) {
		return nil, false, errors.New("invalid theme id")
	}
	path := filepath.Join(s.dir, id, "theme.xml")
	if b, err := os.ReadFile(path); err == nil {
		return b, false, nil
	}
	b, err := builtinThemesFS.ReadFile("themes/" + id + ".xml")
	if err != nil {
		return nil, false, fs.ErrNotExist
	}
	return b, true, nil
}

func (s *themeStore) asset(name string) ([]byte, error) {
	name = filepath.Base(name)
	return builtinThemesFS.ReadFile("themes/" + name)
}

func (s *themeStore) save(id, name string, published bool, xmlBytes []byte) error {
	if !themeIDPattern.MatchString(id) {
		return errors.New("invalid theme id")
	}
	if err := validateThemeXML(xmlBytes); err != nil {
		return err
	}
	// Ensure id/name/published attrs match
	xmlBytes, err := ensureThemeAttrs(xmlBytes, id, name, published)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	dir := filepath.Join(s.dir, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "theme.xml"), xmlBytes, 0o644)
}

func (s *themeStore) setPublished(id string, published bool) error {
	b, builtin, err := s.getXML(id)
	if err != nil {
		return err
	}
	root, err := parseThemeRoot(b)
	if err != nil {
		return err
	}
	if builtin {
		// Materialize builtin to disk before changing publish state
		return s.save(id, root.Name, published, b)
	}
	return s.save(id, root.Name, published, b)
}

func (s *themeStore) delete(id string) error {
	if id == "default" || id == "light" || id == "dark" || id == "system" || id == "hicontrast" {
		return errors.New("cannot delete a built-in theme")
	}
	if !themeIDPattern.MatchString(id) {
		return errors.New("invalid theme id")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	path := filepath.Join(s.dir, id)
	if _, err := os.Stat(path); err != nil {
		return fs.ErrNotExist
	}
	return os.RemoveAll(path)
}

func (s *themeStore) readDiskMeta(id string) (ThemeMeta, error) {
	b, err := os.ReadFile(filepath.Join(s.dir, id, "theme.xml"))
	if err != nil {
		return ThemeMeta{}, err
	}
	root, err := parseThemeRoot(b)
	if err != nil {
		return ThemeMeta{}, err
	}
	return ThemeMeta{
		ID:        id,
		Name:      root.Name,
		Published: strings.EqualFold(root.Published, "true") || root.Published == "",
		Builtin:   false,
	}, nil
}

func (s *themeStore) readBuiltinMeta(id string) (ThemeMeta, error) {
	b, err := builtinThemesFS.ReadFile("themes/" + id + ".xml")
	if err != nil {
		return ThemeMeta{}, err
	}
	root, err := parseThemeRoot(b)
	if err != nil {
		return ThemeMeta{}, err
	}
	name := root.Name
	if name == "" {
		name = id
	}
	return ThemeMeta{
		ID:        id,
		Name:      name,
		Published: strings.EqualFold(root.Published, "true") || root.Published == "",
		Builtin:   true,
	}, nil
}

func parseThemeRoot(b []byte) (themeXMLRoot, error) {
	var root themeXMLRoot
	if err := xml.Unmarshal(b, &root); err != nil {
		return root, fmt.Errorf("invalid theme xml: %w", err)
	}
	if root.ID == "" {
		return root, errors.New("theme id attribute required")
	}
	return root, nil
}

func validateThemeXML(b []byte) error {
	root, err := parseThemeRoot(b)
	if err != nil {
		return err
	}
	if !themeIDPattern.MatchString(root.ID) {
		return errors.New("invalid theme id in xml")
	}
	// Must have colors and layout elements
	if !strings.Contains(string(b), "<colors") || !strings.Contains(string(b), "<layout") {
		return errors.New("theme must include colors and layout")
	}
	return nil
}

func ensureThemeAttrs(b []byte, id, name string, published bool) ([]byte, error) {
	root, err := parseThemeRoot(b)
	if err != nil {
		return nil, err
	}
	if name == "" {
		name = root.Name
	}
	if name == "" {
		name = id
	}
	pub := "false"
	if published {
		pub = "true"
	}
	s := string(b)
	// Replace opening theme tag attrs conservatively
	re := regexp.MustCompile(`(?s)<theme\b[^>]*>`)
	s = re.ReplaceAllString(s, fmt.Sprintf(`<theme id="%s" name="%s" published="%s">`, xmlEscapeAttr(id), xmlEscapeAttr(name), pub))
	return []byte(s), nil
}

func xmlEscapeAttr(v string) string {
	v = strings.ReplaceAll(v, `&`, `&amp;`)
	v = strings.ReplaceAll(v, `"`, `&quot;`)
	v = strings.ReplaceAll(v, `<`, `&lt;`)
	return v
}
