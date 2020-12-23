package ui

import (
	"net/http"
	"path"

	log "github.com/sirupsen/logrus"

	"github.com/ddvk/rmfakecloud/internal/db"
	"github.com/ddvk/rmfakecloud/internal/webassets"
	"github.com/gin-gonic/gin"
)

type ReactAppWrapper struct {
	fs     http.FileSystem
	prefix string
}

const indexReplacement = "/default"

func (w ReactAppWrapper) Open(filepath string) (http.File, error) {
	fullpath := filepath
	//index.html hack
	if filepath != indexReplacement {
		fullpath = path.Join(w.prefix, filepath)
	} else {
		fullpath = "/index.html"
	}
	f, err := w.fs.Open(fullpath)
	return f, err
}
func (w ReactAppWrapper) Register(router *gin.Engine) {
	router.StaticFS(w.prefix, w)

	//hack for index.html
	router.NoRoute(func(c *gin.Context) {
		c.FileFromFS(indexReplacement, w)
	})

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("/favicon.ico", webassets.Assets)
	})
}

type Document struct {
	ID   string `json:id`
	Name string `json:name`
}
type DocumentList struct {
	Documents []Document `json:documents`
}

func badReq(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
	c.Abort()
}

/// RegisterUI add the react ui
func RegisterUI(e *gin.Engine, metaStorer db.MetadataStorer, userStorer db.UserStorer) {
	staticWrapper := ReactAppWrapper{
		fs:     webassets.Assets,
		prefix: "/static",
	}
	staticWrapper.Register(e)

	r := e.Group("/ui/api")
	r.GET("list", func(c *gin.Context) {
		docs, err := metaStorer.GetAllMetadata(false)
		if err != nil {
			log.Error(err)
			badReq(c, err.Error())
			return
		}

		c.JSON(http.StatusOK, docs)
	})

}
