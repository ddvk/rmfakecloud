package ui

import (
	"io"
	"net/http"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/app/passcodestore"
	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/screenshare"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/epub"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	log "github.com/sirupsen/logrus"
)

type backend interface {
	GetDocumentTree(uid string) (tree *viewmodel.DocumentTree, err error)
	Export(uid, doc, exporttype string, opt storage.ExportOption) (stream io.ReadCloser, err error)
	CreateDocument(uid, name, parent string, stream io.Reader) (doc *storage.Document, err error)
	CreateFolder(uid, name, parent string) (doc *storage.Document, err error)
	UpdateDocument(uid, docID, name, parent string) (err error)
	DeleteDocument(uid, docID string) (err error)
	Sync(uid string)
}
type codeGenerator interface {
	NewCode(string) (string, error)
}

type documentHandler interface {
	CreateDocument(uid, name, parent string, stream io.Reader) (doc *storage.Document, err error)
	CreateFolder(uid, name, parent string) (doc *storage.Document, err error)
	GetAllMetadata(uid string) (documents []*messages.RawMetadata, err error)
	ExportDocument(uid, id, format string, exportOption storage.ExportOption) (stream io.ReadCloser, err error)
	GetMetadata(uid, id string) (*messages.RawMetadata, error)
	UpdateMetadata(uid string, r *messages.RawMetadata) error
	RemoveDocument(uid, docid string) error
}

type blobHandler interface {
	GetCachedTree(uid string) (tree *models.HashTree, err error)
	CreateBlobDocument(uid, name, parent string, reader io.Reader) (doc *storage.Document, err error)
	UpdateBlobDocument(uid, docID, name, parent string) (err error)
	DeleteBlobDocument(uid, docID string) (err error)
	CreateBlobFolder(uid, name, parent string) (doc *storage.Document, err error)
	Export(uid, docid string) (io.ReadCloser, error)
	ExportRmDoc(uid, docid string) (io.ReadCloser, error)
	ExportPagePNG(uid, docid string, pageNum int) (io.ReadCloser, error)
	GetEpub(uid, docid string) (io.ReadCloser, error)
	GetEpubManifest(uid, docid string) (*epub.Manifest, error)
	GetEpubFile(uid, docid, filePath string) (io.ReadCloser, string, error)
	GetEpubCoverThumb(uid, docid string) (io.ReadCloser, string, error)
	GetEpubPageThumb(uid, docid string, pageIndex0 int) (io.ReadCloser, string, error)
}

type notificationHub interface {
	Deleted(uid, docID string) error
	Added(uid, docID string) error
	Updated(uid, docID string) error
	Sync(uid string) error
}

type mqttBridge interface {
	PublishSignaling(userID, clientID string, payload []byte)
	HasConnectedClient(userID string) bool
}

// ReactAppWrapper encapsulates the web UI (XML/XSLT pages + JSON APIs).
type ReactAppWrapper struct {
	fs               http.FileSystem
	prefix           string
	cfg              *config.Config
	userStorer       storage.UserStorer
	codeConnector    codeGenerator
	h                *hub.Hub
	passcodeStore    passcodestore.Store
	backends         map[common.SyncVersion]backend
	roomManager      *screenshare.RoomManager
	mqtt             mqttBridge
	webAuthn         *webauthn.WebAuthn
	webAuthnSessions *webAuthnSessionStore
	themes           *themeStore
}

// New creates the web UI app.
func New(cfg *config.Config,
	userStorer storage.UserStorer,
	codeConnector codeGenerator,
	h *hub.Hub,
	pcStore passcodestore.Store,
	docHandler documentHandler,
	blobHandler blobHandler,
	roomManager *screenshare.RoomManager,
	mqttBroker mqttBridge) *ReactAppWrapper {

	backend15 := &backend15{
		blobHandler: blobHandler,
		h:           h,
	}
	backend10 := &backend10{
		documentHandler: docHandler,
		hub:             h,
	}
	staticWrapper := &ReactAppWrapper{
		prefix:        "/assets",
		cfg:           cfg,
		userStorer:    userStorer,
		codeConnector: codeConnector,
		h:             h,
		passcodeStore: pcStore,
		backends: map[common.SyncVersion]backend{
			common.Sync10: backend10,
			common.Sync15: backend15,
		},
		roomManager: roomManager,
		mqtt:        mqttBroker,
		themes:      newThemeStore(cfg.DataDir),
	}
	staticWrapper.initStatic()
	if cfg != nil && cfg.WebAuthn {
		wa, err := webauthn.New(&webauthn.Config{
			RPDisplayName: "rmfakecloud",
			RPID:          cfg.WebAuthnRPID,
			RPOrigins:     cfg.WebAuthnOrigins,
		})
		if err != nil {
			log.Errorf("webauthn init failed, passkeys disabled: %v", err)
		} else {
			staticWrapper.webAuthn = wa
			staticWrapper.webAuthnSessions = newWebAuthnSessionStore()
		}
	}
	return staticWrapper
}

func badReq(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, viewmodel.NewErrorResponse(message))
}
