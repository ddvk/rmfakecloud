package app

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/fs"
	"github.com/ddvk/rmfakecloud/internal/ui"

	"github.com/gin-gonic/gin"
)

const (
	userIDKey   = "UserID"
	deviceIDKey = "DeviceID"
)

// App web app
type App struct {
	router        *gin.Engine
	cfg           *config.Config
	srv           *http.Server
	docStorer     storage.DocumentStorer
	userStorer    storage.UserStorer
	metaStorer    storage.MetadataStorer
	hub           *hub.Hub
	codeConnector common.CodeConnector
}

// Start starts the app
func (app *App) Start() {

	var tlsConfig *tls.Config
	if app.cfg.Certificate.Certificate != nil {
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{
				app.cfg.Certificate,
			},
		}
	}
	app.srv = &http.Server{
		Addr:      ":" + app.cfg.Port,
		Handler:   app.router,
		TLSConfig: tlsConfig,
	}

	if tlsConfig != nil {
		if err := app.srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	} else {
		if err := app.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}
}

// Stop the app
func (app *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// app.hub.Stop()
	if err := app.srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
}

// NewApp constructs an app
func NewApp(cfg *config.Config) App {
	debugMode := log.GetLevel() == log.DebugLevel
	if !debugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	fsStorage := fs.NewStorage(cfg)
	usrs, err := fsStorage.GetUsers()

	if err != nil {
		log.Warn(err)
	}

	if len(usrs) == 0 {
		log.Warn("No users found, the first login will create a user")
		//TODO: not thread safe
		cfg.CreateFirstUser = true
	}
	h := hub.NewHub()
	codeConnector := NewCodeConnector()
	router := gin.Default()

	// corsConfig := cors.DefaultConfig()

	// // corsConfig.AllowOrigins = []string{"*"}

	// // To be able to send tokens to the server.
	// // corsConfig.AllowCredentials = true

	// // OPTIONS method for ReactJS
	// corsConfig.AddAllowMethods("OPTIONS")
	// corsConfig.AddAllowHeaders("Authorization")

	// Register the middleware
	// router.Use(cors.New(corsConfig))

	if debugMode {
		router.Use(requestLoggerMiddleware())
	}

	app := App{
		router:        router,
		cfg:           cfg,
		docStorer:     fsStorage,
		userStorer:    fsStorage,
		metaStorer:    fsStorage,
		hub:           h,
		codeConnector: codeConnector,
	}
	reactApp := ui.New(cfg, fsStorage, codeConnector, h, fsStorage)

	app.registerRoutes(router)
	fsStorage.RegisterRoutes(router)
	reactApp.RegisterRoutes(router)
	return app
}

func badReq(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": message})
}

func internalError(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": message})
}
