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
	ImageUrl string `json:imageUrl`
	ParentId string `json:parentId`
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
					ID:   "001",
					Name: "The Adventures of Huckleberry Finn by Mark Twain",
					ImageUrl: "https://picsum.photos/100/150",
					ParentId: "root",
				},
				Document{
					ID:   "002",
					Name: "The Great Gatsby by F. Scott Fizgerald",
					ImageUrl: "https://images-na.ssl-images-amazon.com/images/I/41iers%2BHLSL._SL160_.jpg",
					ParentId: "root",
				},
				Document{
					ID:   "003",
					Name: "The Stories of Anton Chekhov by Anton Checkhov",
					ImageUrl: "https://picsum.photos/100/150",
					ParentId: "root",
				},
				Document{
					ID:   "004",
					Name: "War and Peace by Leo Tolstoy",
					ImageUrl: "https://picsum.photos/100/150",
					ParentId: "root",
				},

				Document{
					ID:   "005",
					Name: " Madame Bovary by Gustav Flaubert",
					ImageUrl: "https://picsum.photos/100/150",
					ParentId: "root",
				},

				Document{
					ID:   "006",
					Name: "The Adventures of Huckleberry Finn by Mark Twain",
					ImageUrl: "https://picsum.photos/100/150",
					ParentId: "root",
				},

				Document{
					ID:   "007",
					Name: " The Brothers Karamazov by Fyodor Dostoyevsky",
					ImageUrl: "https://picsum.photos/100/150",
					ParentId: "root",
				},

				Document{
					ID:   "008",
					Name: "Don Quixote by Miguel de Cervantes",
					ImageUrl: "https://m.media-amazon.com/images/I/51nBHIQv6zL._SL160_.jpg",
					ParentId: "root",
				},

				Document{
					ID:   "009",
					Name: "Ulysses by James Joyce",
					ImageUrl: "https://picsum.photos/100/150",
					ParentId: "root",
				},
				Document{
					ID:   "010",
					Name: "Crime and Punishment by Fyodor Dostoyevsky",
					ImageUrl: "https://picsum.photos/100/150",
					ParentId: "root",
				},


			},
		}
		c.JSON(http.StatusOK, documentList.Documents)
	})

}
