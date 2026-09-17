package ui

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/applog"
	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/ddvk/rmfakecloud/internal/storage/epub"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	uimethods "github.com/ddvk/rmfakecloud/internal/ui/methods"
	uitemplates "github.com/ddvk/rmfakecloud/internal/ui/templates"
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
	type nav struct{ href, label, note string }
	type section struct {
		id, title, intro string
		paras            []string
		steps            []string
		bullets          []string
		nav              []nav
	}
	sections := []section{
		{
			id:    "overview",
			title: "What this cloud is",
			intro: "rmfakecloud is a self-hosted stand-in for reMarkable’s cloud. Your tablet syncs notebooks, PDFs, and EPUBs here instead of remarkable.com.",
			paras: []string{
				"This website lets you browse My Files, open documents in the browser, upload and download files, pair a tablet, manage integrations, and (if enabled) share the tablet screen live.",
				"Supported tablets include reMarkable 1, 2, Paper Pro, Paper Pro Move, and Paper Pure. Sync works with tablet software through about 3.27.x; newer tablet builds may still work but are not guaranteed.",
			},
			nav: []nav{
				{"/connect", "Connect", "Pair a tablet with a one-time code"},
				{"/documents", "My Files", "Browse and open documents"},
				{"/profile", "Profile", "Password, passkeys, devices, theme"},
			},
		},
		{
			id:    "connect",
			title: "Connect a tablet",
			intro: "Pairing uses a short code from this site, entered on the tablet under Account → Connect.",
			steps: []string{
				"Open Connect in the navigation (or go to the Connect page).",
				"Wait for a one-time code to appear. It refreshes on a timer; use a fresh code if it expires.",
				"On the tablet: open the menu → Account (or Settings → Account) → Connect / pair with cloud.",
				"Enter the code exactly as shown. The tablet should show that it is connected to your cloud.",
				"After pairing, leave the tablet on Wi‑Fi so it can finish the first full sync. The Connect page detects a successful pair and issues a fresh code automatically.",
			},
			paras: []string{
				"If pairing fails, confirm the tablet can reach this server (same Wi‑Fi or VPN as needed), that the code has not expired, and that registration/pairing is allowed on this instance. After an official tablet software update, the on-device proxy or hosts changes may be wiped—reinstall the tablet-side proxy, then reconnect.",
			},
			nav: []nav{{"/connect", "Open Connect", "Generate a pairing code"}},
		},
		{
			id:    "my-files",
			title: "My Files",
			intro: "My Files is the browser file browser for everything synced to this cloud.",
			bullets: []string{
				"Folders appear first; documents are shown as framed page cards (notebook paper, PDF, or EPUB cover).",
				"Corner badges show format and encoding when known (for example RM v6, PDF 1.7, or PDF 1.7 · v6 when a PDF has ink overlays).",
				"Subtitles show page progress (Page N of M) when page counts are available.",
				"Use Search, +, and Select from the dock at the bottom of My Files. The + menu offers Add folder and Upload.",
				"Select mode lets you rename one item, move or delete one or many, and toggle favorites (★).",
				"Favorites (stars) sort ahead of other items in the same folder.",
			},
			paras: []string{
				"Opening a document uses the in-browser viewer for that type. Downloading a PDF or EPUB uses the Download link on the viewer (or export from the API). Notebooks are shown as vector pages in the browser; PDF export of a notebook is only via Download PDF, not as the on-screen viewer.",
			},
			nav: []nav{{"/documents", "Open My Files", "Browse folders and documents"}},
		},
		{
			id:    "notebooks",
			title: "Notebooks",
			intro: "Notebooks are reMarkable handwritten documents (lines / .rm pages).",
			paras: []string{
				"In the browser, notebooks open as a paginated page viewer. The first page shown is the last-opened page from the tablet when that metadata is present; otherwise page 1. Use Previous / Next or the arrow keys to move between pages.",
				"Display uses SVG (or a high-fidelity ink renderer when available). It does not convert the notebook to PDF just to show it. Use Download PDF only when you want a PDF file.",
				"Encoding badges: RM v3 / v5 / v6 refer to the lines file version on the tablet. Older notebooks are often v3 or v5; current tablet software typically writes v6.",
				"If a page cannot be decoded (missing data or tooling unavailable), you may see ruled-paper placeholder art instead of ink. Sync the tablet again or ask an admin to check rendering tools on the server.",
			},
			bullets: []string{
				"Keyboard: ← / → or Page Up / Page Down to change page; Home / End for first / last.",
				"Thumbnails in My Files use the last-opened page when possible.",
			},
		},
		{
			id:    "templates",
			title: "Templates",
			intro: "Templates are page backgrounds and structured layouts you apply on the tablet when creating or editing notebook pages.",
			paras: []string{
				"On the tablet, choose a template (blank, lined, grid, dotted, and others) when you add a page or change the page template. The ink you write sits on top of that background. Synced notebooks keep which template each page used; this cloud shows the handwriting in the notebook viewer and may show a simple ruled placeholder when the exact template art is not available.",
				"reMarkable Methods (sometimes labeled rm Methods on the device) are structured note layouts—for example Cornell notes, outline, mind map, flowchart, or checklist. They sync like other library items and are separate from ordinary notebooks and PDFs.",
				"Admins manage the synced template library from Templates in the web UI: upload .template / .rmdoc files, rename, download, or delete. Tablets pick up changes on the next sync. Non-admins cannot change templates from My Files.",
			},
			bullets: []string{
				"Built-in style templates: blank, lined, grid, dotted, and device-specific variants (previewed on the admin Templates page).",
				"Methods: structured layouts (Cornell, outline, mind map, and similar) for organized note-taking.",
				"Changing a page’s template on the tablet updates after the next sync; older pages keep the template they had when written.",
				"If a page looks blank or only ruled paper in the browser, the ink layer may still be syncing or the server may not have that template’s artwork—resync and open again.",
			},
			nav: []nav{{"/admin/templates", "Templates", "Admin template library"}},
		},
		{
			id:    "pdfs-epubs",
			title: "PDFs and EPUBs",
			intro: "PDFs and EPUBs sync like notebooks and open in dedicated viewers.",
			bullets: []string{
				"PDF: opens in the PDF viewer. Download PDF saves the file. Annotated PDFs may show a badge like PDF 1.7 · v6 when ink overlays exist.",
				"EPUB: opens as an unpacked website (HTML/CSS) in an iframe with a contents list when available. Download EPUB saves the original file.",
				"EPUB covers and page thumbs in My Files come from the book’s cover image or a placeholder if none is found.",
			},
		},
		{
			id:    "upload-download",
			title: "Upload and download",
			intro: "You can add files from the browser and export synced documents.",
			steps: []string{
				"In My Files, open the dock + menu → Upload and choose a PDF, EPUB, or supported document.",
				"Optional: set the parent folder so the file lands in the folder you are viewing.",
				"After upload, wait for sync; the tablet picks up new files on the next sync.",
				"To export: open the document and use Download PDF / Download EPUB, or use the documents API export for advanced clients.",
			},
			paras: []string{
				"Deleting a document in the web UI removes it from this cloud; the tablet will remove it on the next sync. Do not manually delete files from the server data directory unless you understand that the tablet will treat them as deleted.",
			},
		},
		{
			id:    "integrations",
			title: "Integrations",
			intro: "Integrations pull files from external storage into the tablet’s Integrations area.",
			bullets: []string{
				"WebDAV (Nextcloud, ownCloud, and similar)",
				"FTP",
				"Local folders on the server (when configured)",
				"Messaging webhooks and calendar (ICS) when enabled by the administrator",
			},
			paras: []string{
				"Create and remove integrations from the Integrations page. Paths and credentials depend on how your admin configured the server. Dropbox and Google Drive may appear experimental depending on this build.",
			},
			nav: []nav{{"/integrations", "Integrations", "Add or remove storage backends"}},
		},
		{
			id:    "screenshare",
			title: "Screen share",
			intro: "Live screen share shows the tablet display in the browser when MQTT/WebRTC is configured on this instance.",
			paras: []string{
				"Open Screen share from the navigation while the tablet is sharing. If the page stays empty, sharing may be off on the tablet, or this server may not have screen-share services enabled.",
			},
			nav: []nav{{"/screenshare", "Screen share", "Live tablet view"}},
		},
		{
			id:    "profile",
			title: "Profile, themes, and passkeys",
			intro: "Your profile controls sign-in and how the website looks.",
			bullets: []string{
				"Change password from Profile.",
				"Passkeys (when enabled): register a device authenticator while logged in; then you can sign in without typing a password.",
				"Registered devices: tablets that paired with a Connect code are listed on Profile; you can re-issue a device token without a new code.",
				"Themes: pick a shell style (for example reMarkable, Google Docs–like, iCloud–like, or desktop OS) and deploy a theme if you have permission.",
			},
			nav: []nav{{"/profile", "Profile", "Password, passkeys, devices, theme"}},
		},
		{
			id:    "admin",
			title: "Administration",
			intro: "Administrators manage users, themes, and the synced template library.",
			bullets: []string{
				"Create and delete users from Admin.",
				"Theme Studio edits shared themes and chrome layouts.",
				"Templates: upload, rename, download, or delete synced templates and methods; preview built-in SVGs.",
				"Server logs on the Admin page show recent in-memory log lines from this process (Refresh or wait for auto-update).",
				"Server environment (data directory, TLS, SMTP, registration open/closed, rendering tools) is configured on the host—not from this Help page.",
			},
			nav: []nav{
				{"/admin", "Admin", "Users (admins only)"},
				{"/admin/templates", "Templates", "Template library (admins only)"},
			},
		},
		{
			id:    "troubleshooting",
			title: "Troubleshooting",
			intro: "Common fixes without leaving this cloud.",
			bullets: []string{
				"Cannot pair: get a new code on Connect; confirm tablet Wi‑Fi; after a tablet OS update, reinstall the on-device proxy and try again.",
				"Documents missing: wait for sync; check that you are in the correct folder; look in Trash if your tablet moved items there.",
				"Notebook pages blank or ruled paper only: the page may still be syncing, or v6 rendering tools may be unavailable on the server—resync and retry.",
				"Login fails: reset password via an admin if needed; try a registered passkey if you use one.",
				"Upload too large: this instance may enforce a maximum upload size set by the administrator.",
			},
			paras: []string{
				"If the whole site is unreachable, the problem is network or the server process—not something you can fix from Help. Contact whoever hosts this instance.",
			},
		},
	}

	for _, s := range sections {
		fmt.Fprintf(b, `<section id="%s" title="%s">`, xmlAttr(s.id), xmlAttr(s.title))
		if s.intro != "" {
			fmt.Fprintf(b, `<intro>%s</intro>`, esc(s.intro))
		}
		for _, p := range s.paras {
			fmt.Fprintf(b, `<p>%s</p>`, esc(p))
		}
		if len(s.steps) > 0 {
			b.WriteString(`<steps>`)
			for _, st := range s.steps {
				fmt.Fprintf(b, `<step>%s</step>`, esc(st))
			}
			b.WriteString(`</steps>`)
		}
		if len(s.bullets) > 0 {
			b.WriteString(`<bullets>`)
			for _, bu := range s.bullets {
				fmt.Fprintf(b, `<item>%s</item>`, esc(bu))
			}
			b.WriteString(`</bullets>`)
		}
		for _, n := range s.nav {
			fmt.Fprintf(b, `<link href="%s" internal="true" note="%s">%s</link>`,
				xmlAttr(n.href), xmlAttr(n.note), esc(n.label))
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

	devices := []model.RegisteredDevice{}
	if user != nil {
		devices = append(devices, user.RegisteredDevices...)
		sort.Slice(devices, func(i, j int) bool {
			return devices[i].LastSeen.After(devices[j].LastSeen)
		})
	}
	b.WriteString(`<devices>`)
	for _, d := range devices {
		modelLabel := d.Model
		if modelLabel == "" {
			modelLabel = d.DeviceDesc
		}
		if modelLabel == "" {
			modelLabel = "Unknown"
		}
		lastSeen := ""
		if !d.LastSeen.IsZero() {
			lastSeen = d.LastSeen.Local().Format(time.RFC3339)
		}
		registered := ""
		if !d.RegisteredAt.IsZero() {
			registered = d.RegisteredAt.Local().Format(time.RFC3339)
		}
		fmt.Fprintf(&b, `<device id="%s" model="%s" desc="%s" last-seen="%s" registered="%s"/>`,
			xmlAttr(d.DeviceID), xmlAttr(modelLabel), xmlAttr(d.DeviceDesc), xmlAttr(lastSeen), xmlAttr(registered))
	}
	b.WriteString(`</devices>`)

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
		empty := "true"
		if len(d.Entries) > 0 {
			empty = "false"
		}
		fmt.Fprintf(&b, `<folder id="%s" name="%s" pinned="%s" empty="%s" modified="%s"/>`,
			xmlAttr(d.ID), xmlAttr(d.Name), pin, empty, xmlAttr(d.LastModified.Format(time.RFC3339)))
	}
	b.WriteString(`</folders><files>`)
	for _, d := range files {
		pin := "false"
		if d.Pinned {
			pin = "true"
		}
		kind := normalizeDocType(d.DocumentType)
		fmt.Fprintf(&b, `<doc id="%s" name="%s" type="%s" label="%s" writings="%s" pages="%d" page="%d" thumb-page="%d" pinned="%s" modified="%s"/>`,
			xmlAttr(d.ID), xmlAttr(d.Name), xmlAttr(kind), xmlAttr(docFormatLabel(d.FormatLabel, kind)),
			writingsAttr(d.HasWritings),
			d.PageCount, d.CurrentPage, models.ThumbPage1(d.CurrentPage, d.PageCount), pin, xmlAttr(d.LastModified.Format(time.RFC3339)))
	}
	b.WriteString(`</files></documents></body>`)
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func (app *ReactAppWrapper) pageNotebookThumb(c *gin.Context) {
	u := app.requireThumbUser(c)
	if u == nil {
		return
	}
	app.serveNotebookThumb(c, u.ID, common.ParamS("docid", c))
}

func (app *ReactAppWrapper) pageNotebookPageSVG(c *gin.Context) {
	u := app.requireThumbUser(c)
	if u == nil {
		return
	}
	pagenum, err := strconv.Atoi(c.Param("pagenum"))
	if err != nil || pagenum < 1 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	app.serveNotebookSVG(c, u.ID, common.ParamS("docid", c), pagenum)
}

func (app *ReactAppWrapper) pageAnnotatedPageThumb(c *gin.Context) {
	u := app.requireThumbUser(c)
	if u == nil {
		return
	}
	pagenum, err := strconv.Atoi(c.Param("pagenum"))
	if err != nil || pagenum < 1 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	app.serveAnnotatedPageThumb(c, u.ID, common.ParamS("docid", c), pagenum)
}

func (app *ReactAppWrapper) pageNotebook(c *gin.Context, u *pageUser, css, chrome, ft, fm, docID, name string, opened0, pages int, encoding string) {
	if bh := app.blobStorage(); bh != nil {
		if n := bh.NotebookPageCount(u.ID, docID); n > pages {
			pages = n
		}
	}
	if pages < 1 {
		pages = 1
	}
	start := models.ThumbPage1(opened0, pages)
	if q := strings.TrimSpace(c.Query("page")); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			start = n
		}
	}
	if start < 1 {
		start = 1
	}
	if start > pages {
		start = pages
	}
	svgHref := fmt.Sprintf("/documents/%s/page/%d/svg", docID, start)
	pdfHref := "/ui/api/documents/" + docID + "?type=pdf"
	var b bytes.Buffer
	writePageOpen(&b, "notebook", name+" — rmfakecloud", "/documents/"+docID, chrome, css, u, ft, fm, defaultNav("/documents", u.Admin))
	fmt.Fprintf(&b, `<body><notebook doc-id="%s" name="%s" encoding="%s" page="%d" pages="%d" svg-href="%s" download-href="%s"/></body>`,
		xmlAttr(docID), xmlAttr(name), xmlAttr(docFormatLabel(encoding, "notebook")), start, pages, xmlAttr(svgHref), xmlAttr(pdfHref))
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func (app *ReactAppWrapper) pageEpubThumb(c *gin.Context) {
	u := app.requireThumbUser(c)
	if u == nil {
		return
	}
	app.serveEpubThumb(c, u.ID, common.ParamS("docid", c))
}

func docFormatLabel(label, kind string) string {
	label = strings.TrimSpace(label)
	if label != "" {
		return label
	}
	switch normalizeDocType(kind) {
	case "pdf":
		return "PDF"
	case "epub":
		return "EPUB"
	default:
		return "RM"
	}
}

func writingsAttr(has bool) string {
	if has {
		return "true"
	}
	return "false"
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
	encoding := ""
	hasWritings := false
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
			encoding = d.FormatLabel
			hasWritings = d.HasWritings
		}
	}
	encoding = docFormatLabel(encoding, kind)
	if kind == "epub" {
		if hasWritings {
			app.pageAnnotated(c, u, css, chrome, ft, fm, docID, name, page, pages, encoding, "epub")
			return
		}
		app.pageEpub(c, u, css, chrome, ft, fm, docID, name, page, pages)
		return
	}
	if kind == "notebook" {
		app.pageNotebook(c, u, css, chrome, ft, fm, docID, name, page, pages, encoding)
		return
	}
	if hasWritings {
		app.pageAnnotated(c, u, css, chrome, ft, fm, docID, name, page, pages, encoding, "pdf")
		return
	}
	var b bytes.Buffer
	writePageOpen(&b, "pdf", name+" — rmfakecloud", "/documents/"+docID, chrome, css, u, ft, fm, defaultNav("/documents", u.Admin))
	fmt.Fprintf(&b, `<body><pdf doc-id="%s" name="%s" encoding="%s" url="/ui/api/documents/%s?type=pdf"/></body>`,
		xmlAttr(docID), xmlAttr(name), xmlAttr(encoding), xmlAttr(docID))
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

// pageAnnotated shows PDF/EPUB pages as PNG composites (payload/cover × .rm ink).
func (app *ReactAppWrapper) pageAnnotated(c *gin.Context, u *pageUser, css, chrome, ft, fm, docID, name string, opened0, pages int, encoding, kind string) {
	if pages < 1 {
		pages = 1
	}
	start := models.ThumbPage1(opened0, pages)
	if q := strings.TrimSpace(c.Query("page")); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			start = n
		}
	}
	if start < 1 {
		start = 1
	}
	if start > pages {
		start = pages
	}
	pngHref := fmt.Sprintf("/ui/api/documents/%s/page/%d", docID, start)
	dlType := "pdf"
	dlLabel := "Download PDF"
	if kind == "epub" {
		dlType = "epub"
		dlLabel = "Download EPUB"
	}
	dlHref := "/ui/api/documents/" + docID + "?type=" + dlType
	var b bytes.Buffer
	writePageOpen(&b, "annotated", name+" — rmfakecloud", "/documents/"+docID, chrome, css, u, ft, fm, defaultNav("/documents", u.Admin))
	fmt.Fprintf(&b, `<body><annotated doc-id="%s" name="%s" encoding="%s" kind="%s" page="%d" pages="%d" png-href="%s" download-href="%s" download-label="%s"/></body>`,
		xmlAttr(docID), xmlAttr(name), xmlAttr(encoding), xmlAttr(kind), start, pages, xmlAttr(pngHref), xmlAttr(dlHref), xmlAttr(dlLabel))
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
	b.WriteString(`<logs>`)
	for _, line := range applog.Default().Texts(200) {
		fmt.Fprintf(&b, `<line>%s</line>`, esc(line))
	}
	b.WriteString(`</logs></admin></body>`)
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func (app *ReactAppWrapper) pageTemplates(c *gin.Context) {
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
	backend := app.getBackend(c)
	tree, err := backend.GetDocumentTree(u.ID)
	if err != nil {
		log.Error(err)
		redirectFlash(c, "/admin", "error", "Unable to load templates")
		return
	}

	var b bytes.Buffer
	writePageOpen(&b, "templates", "Templates — rmfakecloud", "/admin/templates", chrome, css, u, ft, fm, defaultNav("/admin/templates", u.Admin))
	b.WriteString(`<body><templates-admin>`)
	b.WriteString(`<templates>`)
	writeTemplateAdminItems(&b, tree.Templates, "template")
	b.WriteString(`</templates>`)
	b.WriteString(`<methods>`)
	writeTemplateAdminItems(&b, tree.Methods, "method")
	b.WriteString(`</methods>`)
	b.WriteString(`<builtin-templates>`)
	for _, t := range uitemplates.ListBuiltins() {
		fmt.Fprintf(&b, `<item id="%s" name="%s" kind="template" builtin="true"/>`,
			xmlAttr(t.ID), xmlAttr(t.Name))
	}
	b.WriteString(`</builtin-templates>`)
	b.WriteString(`<builtin-methods>`)
	for _, m := range uimethods.ListBuiltins() {
		fmt.Fprintf(&b, `<item id="%s" name="%s" kind="method" builtin="true"/>`,
			xmlAttr(m.ID), xmlAttr(m.Name))
	}
	b.WriteString(`</builtin-methods></templates-admin></body>`)
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func writeTemplateAdminItems(b *bytes.Buffer, entries []viewmodel.Entry, kind string) {
	for _, e := range entries {
		d, ok := e.(*viewmodel.Document)
		if !ok || d == nil {
			continue
		}
		fmt.Fprintf(b, `<item id="%s" name="%s" kind="%s" builtin="false" modified="%s"/>`,
			xmlAttr(d.ID), xmlAttr(d.Name), xmlAttr(kind), xmlAttr(d.LastModified.Format(time.RFC3339)))
	}
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
