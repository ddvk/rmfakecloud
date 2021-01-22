package ui

import (
	"net/http"
	"path"

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

/// RegisterUI add the react ui
func RegisterUI(e *gin.Engine) {
	staticWrapper := ReactAppWrapper{
		fs:     webassets.Assets,
		prefix: "/static",
	}
	staticWrapper.Register(e)

	r := e.Group("/ui/api")
	r.GET("list", func(c *gin.Context) {
		documentList := DocumentList{
			Documents: []Document{
				Document{
					ID:   "doc1",
					Name: "test",
				},
				Document{
					ID:   "doc2",
					Name: "test",
				},
				Document{
					ID:   "doc3",
					Name: "test",
				},
			},
		}
		c.JSON(http.StatusOK, documentList.Documents)
	})

}
