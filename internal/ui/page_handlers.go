package ui

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/ddvk/rmfakecloud/internal/storage/epub"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (app *ReactAppWrapper) pageHome(c *gin.Context) {
	u := app.optionalUser(c)
	if u == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	_, css, chrome, _ := app.loadUserTheme(c, u)
	ft, fm := app.flashFromQuery(c)
	var b bytes.Buffer
	writePageOpen(&b, "home", "rmfakecloud", "/", chrome, css, u, ft, fm, defaultNav("/", u.Admin))
	b.WriteString(`<body><home>`)
	b.WriteString(`<p>Self-hosted reMarkable cloud. Use Documents to browse notebooks, Connect to pair a tablet, and Profile to deploy a theme.</p>`)
	b.WriteString(`<p><a href="/help">Help</a> · <a href="/documents">Documents</a> · <a href="/connect">Connect</a></p>`)
	b.WriteString(`</home></body>`)
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func (app *ReactAppWrapper) pageHelp(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	_, css, chrome, _ := app.loadUserTheme(c, u)
	ft, fm := app.flashFromQuery(c)
	var b bytes.Buffer
	writePageOpen(&b, "help", "Help — rmfakecloud", "/help", chrome, css, u, ft, fm, defaultNav("/help", u.Admin))
	b.WriteString(`<body><help>`)
	writeHelpSections(&b)
	b.WriteString(`</help></body>`)
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func writeHelpSections(b *bytes.Buffer) {
	type link struct {
		href, label, note string
		internal          bool
	}
	type section struct {
		id, title, intro string
		links            []link
	}
	docs := "https://ddvk.github.io/rmfakecloud"
	hc := "https://support.remarkable.com/hc/en-us"
	sections := []section{
		{id: "this-instance", title: "This cloud (quick start)", intro: "Day-to-day pages on this instance.", links: []link{
			{"/connect", "Connect / pair a tablet", "One-time code for Account → Connect", true},
			{"/documents", "Documents", "Browse, upload, download", true},
			{"/integrations", "Integrations", "WebDAV, FTP, local folders", true},
			{"/screenshare", "Screen share", "Live view when MQTT/WebRTC is configured", true},
			{"/profile", "Profile", "Password, passkeys, deploy a theme", true},
		}},
		{id: "official", title: "Official reMarkable help", intro: "From reMarkable Support.", links: []link{
			{"https://support.remarkable.com/s/", "reMarkable Support home", "Search official articles", false},
			{hc + "/categories/360000160977-Getting-Started", "Getting started", "Setup and first steps", false},
			{hc + "/categories/360000161018-My-reMarkable", "My reMarkable", "Device settings and care", false},
			{hc + "/categories/360000160997-Software", "Software", "Updates and apps", false},
		}},
		{id: "rmfakecloud", title: "rmfakecloud documentation", intro: "Project docs for self-hosted sync.", links: []link{
			{docs + "/", "Documentation home", "Overview", false},
			{docs + "/remarkable/setup/", "Tablet setup", "Proxy, hosts, pairing", false},
			{docs + "/install/configuration/", "Server configuration", "Environment variables", false},
			{docs + "/usage/integrations/", "Integrations usage", "WebDAV, FTP, webhooks", false},
			{"https://github.com/ddvk/rmfakecloud", "GitHub repository", "Source and issues", false},
		}},
	}
	for _, s := range sections {
		fmt.Fprintf(b, `<section id="%s" title="%s"><intro>%s</intro>`, xmlAttr(s.id), xmlAttr(s.title), esc(s.intro))
		for _, l := range s.links {
			internal := "false"
			if l.internal {
				internal = "true"
			}
			fmt.Fprintf(b, `<link href="%s" internal="%s" note="%s">%s</link>`,
				xmlAttr(l.href), internal, xmlAttr(l.note), esc(l.label))
		}
		b.WriteString(`</section>`)
	}
}

func (app *ReactAppWrapper) pageLogin(c *gin.Context) {
	if u := app.optionalUser(c); u != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}
	_, css, chrome, _ := app.loadUserTheme(c, nil)
	errMsg := c.Query("error")
	email := c.Query("email")
	passkey := "false"
	if app.webAuthnEnabled() {
		passkey = "true"
	}
	reg := "false"
	if app.cfg.RegistrationOpen {
		reg = "true"
	}
	var b bytes.Buffer
	writePageOpen(&b, "login", "Login — rmfakecloud", "/login", chrome, css, nil, "", "", nil)
	fmt.Fprintf(&b, `<body><login show-brand="true" passkey="%s" registration="%s" email="%s" error="%s"/></body>`,
		passkey, reg, xmlAttr(email), xmlAttr(errMsg))
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func (app *ReactAppWrapper) pageConnect(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	themeXML, css, chrome, ff := app.loadUserTheme(c, u)
	gauge, ploc, pstyle, prompt := connectAttrsFromTheme(themeXML, ff)
	ft, fm := app.flashFromQuery(c)
	var b bytes.Buffer
	writePageOpen(&b, "connect", "Connect — rmfakecloud", "/connect", chrome, css, u, ft, fm, defaultNav("/connect", u.Admin))
	fmt.Fprintf(&b, `<body><connect code="" gauge-type="%s" prompt-location="%s" prompt-style="%s" prompt="%s"/></body>`,
		xmlAttr(gauge), xmlAttr(ploc), xmlAttr(pstyle), xmlAttr(prompt))
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func (app *ReactAppWrapper) pageProfile(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	user := app.getModelUser(u.ID)
	_, css, chrome, _ := app.loadUserTheme(c, u)
	ft, fm := app.flashFromQuery(c)
	themes, _ := app.themes.list(u.Admin)
	selected := "default"
	if user != nil && user.ThemeID != "" {
		selected = user.ThemeID
	}
	if selected == "default" {
		selected = "dark"
	}
	var b bytes.Buffer
	writePageOpen(&b, "profile", "Profile — rmfakecloud", "/profile", chrome, css, u, ft, fm, defaultNav("/profile", u.Admin))
	b.WriteString(`<body><profile>`)
	b.WriteString(`<themes>`)
	for _, t := range themes {
		if !t.Published && !u.Admin {
			continue
		}
		sel := ""
		if t.ID == selected {
			sel = ` selected="true"`
		}
		fmt.Fprintf(&b, `<theme id="%s" name="%s"%s/>`, xmlAttr(t.ID), xmlAttr(t.Name), sel)
	}
	b.WriteString(`</themes>`)
	b.WriteString(`<overrides>`)
	keys := []string{"background1", "background2", "foreground1", "foreground2", "action"}
	overrides := map[string]string{}
	if user != nil {
		overrides = user.ThemeColorOverrides
	}
	for _, k := range keys {
		v := "#212529"
		if overrides != nil && overrides[k] != "" {
			v = overrides[k]
		} else {
			switch k {
			case "background2":
				v = "#0f0f0f"
			case "foreground1":
				v = "#f8f7f6"
			case "foreground2":
				v = "#e8e3d9"
			case "action":
				v = "#EE7B30"
			}
		}
		fmt.Fprintf(&b, `<color key="%s" value="%s"/>`, xmlAttr(k), xmlAttr(v))
	}
	b.WriteString(`</overrides>`)
	enabled := "false"
	if app.webAuthnEnabled() {
		enabled = "true"
	}
	fmt.Fprintf(&b, `<passkeys enabled="%s">`, enabled)
	if user != nil {
		for _, cred := range user.WebAuthnCredentials {
			name := cred.Name
			cid := model.CredentialIDBase64(cred.ID)
			if name == "" {
				name = cid
			}
			created := ""
			if !cred.CreatedAt.IsZero() {
				created = cred.CreatedAt.Format(time.RFC3339)
			}
			fmt.Fprintf(&b, `<cred id="%s" name="%s" created="%s"/>`, xmlAttr(cid), xmlAttr(name), xmlAttr(created))
		}
	}
	b.WriteString(`</passkeys>`)
	b.WriteString(`</profile></body>`)
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func (app *ReactAppWrapper) pageDocuments(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	_, css, chrome, _ := app.loadUserTheme(c, u)
	ft, fm := app.flashFromQuery(c)
	folderID := c.Query("folder")
	backend := app.getBackend(c)
	tree, err := backend.GetDocumentTree(u.ID)
	if err != nil {
		log.Error(err)
		redirectFlash(c, "/documents", "error", "Unable to load documents")
		return
	}
	folderName := "My Files"
	parentID := ""
	entries := tree.Entries
	if folderID != "" {
		if found := findFolderEntries(tree.Entries, folderID); found != nil {
			entries = found
			folderName = findFolderName(tree.Entries, folderID)
			if folderName == "" {
				folderName = "My Files"
			}
			parentID = findFolderParent(tree.Entries, folderID)
		} else {
			entries = nil
		}
	}
	folders, files := splitDocEntries(entries)
	sort.SliceStable(folders, func(i, j int) bool {
		if folders[i].Pinned != folders[j].Pinned {
			return folders[i].Pinned
		}
		return folders[i].Name < folders[j].Name
	})
	sort.SliceStable(files, func(i, j int) bool {
		if files[i].Pinned != files[j].Pinned {
			return files[i].Pinned
		}
		return files[i].LastModified.After(files[j].LastModified)
	})

	var b bytes.Buffer
	writePageOpen(&b, "documents", folderName+" — rmfakecloud", "/documents", chrome, css, u, ft, fm, defaultNav("/documents", u.Admin))
	fmt.Fprintf(&b, `<body><documents folder-id="%s" folder-name="%s" parent-id="%s">`,
		xmlAttr(folderID), xmlAttr(folderName), xmlAttr(parentID))
	b.WriteString(`<folders>`)
	for _, d := range folders {
		pin := "false"
		if d.Pinned {
			pin = "true"
		}
		fmt.Fprintf(&b, `<folder id="%s" name="%s" pinned="%s" modified="%s"/>`,
			xmlAttr(d.ID), xmlAttr(d.Name), pin, xmlAttr(d.LastModified.Format(time.RFC3339)))
	}
	b.WriteString(`</folders><files>`)
	for _, d := range files {
		pin := "false"
		if d.Pinned {
			pin = "true"
		}
		fmt.Fprintf(&b, `<doc id="%s" name="%s" type="%s" pages="%d" page="%d" thumb-page="%d" pinned="%s" modified="%s"/>`,
			xmlAttr(d.ID), xmlAttr(d.Name), xmlAttr(normalizeDocType(d.DocumentType)),
			d.PageCount, d.CurrentPage, models.ThumbPage1(d.CurrentPage, d.PageCount), pin, xmlAttr(d.LastModified.Format(time.RFC3339)))
	}
	b.WriteString(`</files></documents></body>`)
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func normalizeDocType(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	t = strings.TrimPrefix(t, ".")
	switch t {
	case "pdf", "application/pdf":
		return "pdf"
	case "epub", "application/epub+zip":
		return "epub"
	case "folder", "collection":
		return "folder"
	default:
		return "notebook"
	}
}

func splitDocEntries(entries []viewmodel.Entry) (folders []viewmodel.Directory, files []viewmodel.Document) {
	for _, e := range entries {
		switch d := e.(type) {
		case *viewmodel.Directory:
			folders = append(folders, *d)
		case *viewmodel.Document:
			files = append(files, *d)
		}
	}
	return
}

func findFolderParent(entries []viewmodel.Entry, id string) string {
	var walk func(list []viewmodel.Entry, parent string) (string, bool)
	walk = func(list []viewmodel.Entry, parent string) (string, bool) {
		for _, e := range list {
			d, ok := e.(*viewmodel.Directory)
			if !ok {
				continue
			}
			if d.ID == id {
				return parent, true
			}
			if p, found := walk(d.Entries, d.ID); found {
				return p, true
			}
		}
		return "", false
	}
	p, _ := walk(entries, "")
	return p
}

func findFolderEntries(entries []viewmodel.Entry, id string) []viewmodel.Entry {
	for _, e := range entries {
		d, ok := e.(*viewmodel.Directory)
		if !ok {
			continue
		}
		if d.ID == id {
			return d.Entries
		}
		if nested := findFolderEntries(d.Entries, id); nested != nil {
			return nested
		}
	}
	return nil
}

func findFolderName(entries []viewmodel.Entry, id string) string {
	for _, e := range entries {
		d, ok := e.(*viewmodel.Directory)
		if !ok {
			continue
		}
		if d.ID == id {
			return d.Name
		}
		if n := findFolderName(d.Entries, id); n != "" {
			return n
		}
	}
	return ""
}

func findDocument(entries []viewmodel.Entry, id string) *viewmodel.Document {
	for _, e := range entries {
		switch d := e.(type) {
		case *viewmodel.Document:
			if d.ID == id {
				return d
			}
		case *viewmodel.Directory:
			if found := findDocument(d.Entries, id); found != nil {
				return found
			}
		}
	}
	return nil
}

func (app *ReactAppWrapper) pagePDF(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	docID := c.Param("docid")
	_, css, chrome, _ := app.loadUserTheme(c, u)
	ft, fm := app.flashFromQuery(c)
	name := docID
	kind := "pdf"
	pages := 0
	page := 0
	backend := app.getBackend(c)
	if tree, err := backend.GetDocumentTree(u.ID); err == nil && tree != nil {
		d := findDocument(tree.Entries, docID)
		if d == nil {
			d = findDocument(tree.Trash, docID)
		}
		if d != nil {
			name = d.Name
			kind = normalizeDocType(d.DocumentType)
			pages = d.PageCount
			page = d.CurrentPage
		}
	}
	if kind == "epub" {
		app.pageEpub(c, u, css, chrome, ft, fm, docID, name, page, pages)
		return
	}
	var b bytes.Buffer
	writePageOpen(&b, "pdf", name+" — rmfakecloud", "/documents/"+docID, chrome, css, u, ft, fm, defaultNav("/documents", u.Admin))
	fmt.Fprintf(&b, `<body><pdf doc-id="%s" name="%s" url="/ui/api/documents/%s?type=pdf"/></body>`,
		xmlAttr(docID), xmlAttr(name), xmlAttr(docID))
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func (app *ReactAppWrapper) pageEpub(c *gin.Context, u *pageUser, css, chrome, ft, fm, docID, name string, page, pages int) {
	type epubManifestBackend interface {
		GetEpubManifest(uid, docid string) (*epub.Manifest, error)
	}
	var spine []string
	backend := app.getBackend(c)
	if eb, ok := backend.(epubManifestBackend); ok {
		if man, err := eb.GetEpubManifest(u.ID, docID); err == nil && man != nil {
			spine = man.Spine
		}
	}
	if pages == 0 {
		pages = len(spine)
	}
	start := models.ThumbPage1(page, pages)
	if start < 1 {
		start = 1
	}
	startPath := ""
	if len(spine) > 0 {
		idx := start - 1
		if idx < 0 {
			idx = 0
		}
		if idx >= len(spine) {
			idx = len(spine) - 1
		}
		startPath = spine[idx]
	}

	startHref := ""
	if startPath != "" {
		startHref = epubAssetURL(docID, startPath)
	}
	var b bytes.Buffer
	writePageOpen(&b, "epub", name+" — rmfakecloud", "/documents/"+docID, chrome, css, u, ft, fm, defaultNav("/documents", u.Admin))
	fmt.Fprintf(&b, `<body><epub doc-id="%s" name="%s" start="%s" start-href="%s" download-href="%s" page="%d" pages="%d">`,
		xmlAttr(docID), xmlAttr(name), xmlAttr(startPath), xmlAttr(startHref),
		xmlAttr("/ui/api/documents/"+docID+"?type=epub"), page, pages)
	b.WriteString(`<spine>`)
	for i, p := range spine {
		label := epubSpineLabel(p, i)
		fmt.Fprintf(&b, `<item path="%s" label="%s" href="%s"/>`, xmlAttr(p), xmlAttr(label), xmlAttr(epubAssetURL(docID, p)))
	}
	b.WriteString(`</spine></epub></body>`)
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func (app *ReactAppWrapper) pageIntegrations(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	user := app.getModelUser(u.ID)
	_, css, chrome, _ := app.loadUserTheme(c, u)
	ft, fm := app.flashFromQuery(c)
	var b bytes.Buffer
	writePageOpen(&b, "integrations", "Integrations — rmfakecloud", "/integrations", chrome, css, u, ft, fm, defaultNav("/integrations", u.Admin))
	b.WriteString(`<body><integrations>`)
	if user != nil {
		for _, i := range user.Integrations {
			fmt.Fprintf(&b, `<integration id="%s" name="%s" provider="%s"/>`,
				xmlAttr(i.ID), xmlAttr(i.Name), xmlAttr(i.Provider))
		}
	}
	b.WriteString(`</integrations></body>`)
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func (app *ReactAppWrapper) pageAdmin(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	if !u.Admin {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	_, css, chrome, _ := app.loadUserTheme(c, u)
	ft, fm := app.flashFromQuery(c)
	users, err := app.userStorer.GetUsers()
	if err != nil {
		redirectFlash(c, "/", "error", "Unable to load users")
		return
	}
	var b bytes.Buffer
	writePageOpen(&b, "admin", "Admin — rmfakecloud", "/admin", chrome, css, u, ft, fm, defaultNav("/admin", u.Admin))
	b.WriteString(`<body><admin>`)
	for _, usr := range users {
		admin := "false"
		if usr.IsAdmin {
			admin = "true"
		}
		fmt.Fprintf(&b, `<user id="%s" email="%s" name="%s" admin="%s"/>`,
			xmlAttr(usr.ID), xmlAttr(usr.Email), xmlAttr(usr.Name), admin)
	}
	b.WriteString(`</admin></body>`)
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func (app *ReactAppWrapper) pageThemeStudio(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	if !u.Admin {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	_, css, chrome, _ := app.loadUserTheme(c, u)
	ft, fm := app.flashFromQuery(c)
	themes, _ := app.themes.list(true)
	editID := c.Query("id")
	if editID == "" && len(themes) > 0 {
		editID = themes[0].ID
	}
	var editXML []byte
	var editName string
	var published bool
	if editID != "" {
		editXML, _, _ = app.themes.getXML(editID)
		root, err := parseThemeRoot(editXML)
		if err == nil {
			editName = root.Name
			published = strings.EqualFold(root.Published, "true") || root.Published == ""
		}
	}
	var b bytes.Buffer
	writePageOpen(&b, "themes", "Theme studio — rmfakecloud", "/admin/themes", chrome, css, u, ft, fm, defaultNav("/admin", u.Admin))
	b.WriteString(`<body><themes-studio>`)
	for _, t := range themes {
		pub := "false"
		if t.Published {
			pub = "true"
		}
		builtin := "false"
		if t.Builtin {
			builtin = "true"
		}
		fmt.Fprintf(&b, `<theme id="%s" name="%s" published="%s" builtin="%s"/>`,
			xmlAttr(t.ID), xmlAttr(t.Name), pub, builtin)
	}
	pub := "false"
	if published {
		pub = "true"
	}
	fmt.Fprintf(&b, `<editor id="%s" name="%s" published="%s">%s</editor>`,
		xmlAttr(editID), xmlAttr(editName), pub, xmlCDATA(string(editXML)))
	b.WriteString(`</themes-studio></body>`)
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func (app *ReactAppWrapper) pageScreenShare(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	_, css, chrome, _ := app.loadUserTheme(c, u)
	ft, fm := app.flashFromQuery(c)
	var b bytes.Buffer
	writePageOpen(&b, "screenshare", "Screen share — rmfakecloud", "/screenshare", chrome, css, u, ft, fm, defaultNav("/screenshare", u.Admin))
	b.WriteString(`<body><screenshare/></body>`)
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func (app *ReactAppWrapper) page404(c *gin.Context) {
	u := app.optionalUser(c)
	_, css, chrome, _ := app.loadUserTheme(c, u)
	admin := u != nil && u.Admin
	var b bytes.Buffer
	writePageOpen(&b, "error", "Not found — rmfakecloud", c.Request.URL.Path, chrome, css, u, "", "", defaultNav("", admin))
	b.WriteString(`<body><error code="404" message="Page not found"/></body>`)
	writePageClose(&b)
	setFormFactorHeaders(c)
	html, ok, err := app.transformPageXML(b.Bytes())
	if err != nil {
		log.Warn("server xslt failed, using browser bootstrap: ", err)
	}
	payload := bootstrapXSLTHTML(b.Bytes())
	if ok && len(html) > 0 {
		payload = html
	}
	c.Data(http.StatusNotFound, "text/html; charset=utf-8", payload)
}
