package ui

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/storage"
	uimethods "github.com/ddvk/rmfakecloud/internal/ui/methods"
	uitemplates "github.com/ddvk/rmfakecloud/internal/ui/templates"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (app *ReactAppWrapper) requireAdminPageUser(c *gin.Context) *pageUser {
	u := app.requirePageUser(c)
	if u == nil {
		return nil
	}
	if !u.Admin {
		c.AbortWithStatus(http.StatusForbidden)
		return nil
	}
	return u
}

func (app *ReactAppWrapper) lookupSyncedTemplate(c *gin.Context, uid, docid string) (name, kind string, ok bool) {
	backend := app.getBackend(c)
	tree, err := backend.GetDocumentTree(uid)
	if err != nil || tree == nil {
		return "", "", false
	}
	for _, e := range tree.Templates {
		if d, is := e.(*viewmodel.Document); is && d != nil && d.ID == docid {
			return d.Name, "template", true
		}
	}
	for _, e := range tree.Methods {
		if d, is := e.(*viewmodel.Document); is && d != nil && d.ID == docid {
			return d.Name, "method", true
		}
	}
	return "", "", false
}

func (app *ReactAppWrapper) isSyncedTemplateDoc(c *gin.Context, uid, docid string) bool {
	_, _, ok := app.lookupSyncedTemplate(c, uid, docid)
	return ok
}

func (app *ReactAppWrapper) formUploadTemplate(c *gin.Context) {
	u := app.requireAdminPageUser(c)
	if u == nil {
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		redirectFlash(c, "/admin/templates", "error", "No file uploaded")
		return
	}
	ext := strings.ToLower(path.Ext(file.Filename))
	if ext != storage.TemplateFileExt && ext != storage.RmDocFileExt {
		redirectFlash(c, "/admin/templates", "error", "Upload a .template or .rmdoc file")
		return
	}
	f, err := file.Open()
	if err != nil {
		redirectFlash(c, "/admin/templates", "error", "Cannot open upload")
		return
	}
	defer f.Close()
	backend := app.getBackend(c)
	if _, err := backend.CreateDocument(u.ID, file.Filename, "", f); err != nil {
		log.Error(err)
		redirectFlash(c, "/admin/templates", "error", err.Error())
		return
	}
	backend.Sync(u.ID)
	redirectFlash(c, "/admin/templates", "success", "Template uploaded")
}

func (app *ReactAppWrapper) formUpdateTemplate(c *gin.Context) {
	u := app.requireAdminPageUser(c)
	if u == nil {
		return
	}
	docid := c.Param("docid")
	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		redirectFlash(c, "/admin/templates", "error", "Name required")
		return
	}
	if !app.isSyncedTemplateDoc(c, u.ID, docid) {
		redirectFlash(c, "/admin/templates", "error", "Not a template")
		return
	}
	backend := app.getBackend(c)
	if err := backend.UpdateDocument(u.ID, docid, name, ""); err != nil {
		redirectFlash(c, "/admin/templates", "error", err.Error())
		return
	}
	backend.Sync(u.ID)
	redirectFlash(c, "/admin/templates", "success", "Renamed")
}

func (app *ReactAppWrapper) formDeleteTemplate(c *gin.Context) {
	u := app.requireAdminPageUser(c)
	if u == nil {
		return
	}
	docid := c.Param("docid")
	if !app.isSyncedTemplateDoc(c, u.ID, docid) {
		redirectFlash(c, "/admin/templates", "error", "Not a template")
		return
	}
	backend := app.getBackend(c)
	if err := backend.DeleteDocument(u.ID, docid); err != nil {
		redirectFlash(c, "/admin/templates", "error", err.Error())
		return
	}
	backend.Sync(u.ID)
	redirectFlash(c, "/admin/templates", "success", "Deleted")
}

func (app *ReactAppWrapper) downloadTemplate(c *gin.Context) {
	u := app.requireAdminPageUser(c)
	if u == nil {
		return
	}
	docid := c.Param("docid")
	name, _, ok := app.lookupSyncedTemplate(c, u.ID, docid)
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	backend := app.getBackend(c)
	reader, err := backend.GetTemplate(u.ID, docid)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	defer reader.Close()
	filename := name
	if filename == "" {
		filename = docid
	}
	if !strings.HasSuffix(strings.ToLower(filename), storage.TemplateFileExt) {
		filename += storage.TemplateFileExt
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Header("X-Content-Type-Options", "nosniff")
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

func (app *ReactAppWrapper) syncedTemplateThumb(c *gin.Context) {
	u := app.requireAdminPageUser(c)
	if u == nil {
		return
	}
	docid := c.Param("docid")
	if !app.isSyncedTemplateDoc(c, u.ID, docid) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	backend := app.getBackend(c)
	reader, err := backend.GetTemplate(u.ID, docid)
	if err != nil {
		log.Debug(err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	defer reader.Close()
	svg, err := templateIconSVG(reader)
	if err != nil || svg == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Header("Content-Type", "image/svg+xml; charset=utf-8")
	c.Header("Cache-Control", "private, max-age=300")
	c.Header("X-Content-Type-Options", "nosniff")
	c.String(http.StatusOK, svg)
}

type templateFileJSON struct {
	IconData string `json:"iconData"`
}

func templateIconSVG(r io.Reader) (string, error) {
	var meta templateFileJSON
	if err := json.NewDecoder(r).Decode(&meta); err != nil {
		return "", err
	}
	raw := strings.TrimSpace(meta.IconData)
	if raw == "" {
		return "", fmt.Errorf("no iconData")
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(raw)
		if err != nil {
			return "", err
		}
	}
	s := strings.TrimSpace(string(decoded))
	if !strings.Contains(strings.ToLower(s), "<svg") {
		return "", fmt.Errorf("iconData is not svg")
	}
	return s, nil
}

func (app *ReactAppWrapper) builtinTemplateSVG(c *gin.Context) {
	u := app.requireAdminPageUser(c)
	if u == nil {
		return
	}
	kind := c.Param("kind")
	id := c.Param("id")
	var svg string
	switch kind {
	case "template":
		svg = uitemplates.GetSVG(id)
	case "method":
		svg = uimethods.GetSVG(id)
	}
	if svg == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Header("Content-Type", "image/svg+xml; charset=utf-8")
	c.Header("X-Content-Type-Options", "nosniff")
	c.String(http.StatusOK, svg)
}
