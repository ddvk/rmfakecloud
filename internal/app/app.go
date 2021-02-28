package app

import (
	"context"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/db"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/ui"

	"github.com/gin-contrib/cors"
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
	userStorer    db.UserStorer
	metaStorer    db.MetadataStorer
	hub           *Hub
	ui            *ui.ReactAppWrapper
	codeConnector common.CodeConnector
}

// Start starts the app
func (app *App) Start() {
	app.srv = &http.Server{
		Addr:    ":" + app.cfg.Port,
		Handler: app.router,
	}

	if err := app.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}

// Stop the app
func (app *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
}

// NewApp constructs an app
func NewApp(cfg *config.Config, metaStorer db.MetadataStorer, docStorer storage.DocumentStorer, userStorer db.UserStorer) App {
	debugMode := log.GetLevel() == log.DebugLevel
	if !debugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	hub := NewHub()
	codeConnector := NewCodeConnector()
	gin.ForceConsoleColor()
	router := gin.Default()

	corsConfig := cors.DefaultConfig()

	// corsConfig.AllowOrigins = []string{"*"}

	// To be able to send tokens to the server.
	// corsConfig.AllowCredentials = true

	// OPTIONS method for ReactJS
	corsConfig.AddAllowMethods("OPTIONS")
	corsConfig.AddAllowHeaders("Authorization")

	// Register the middleware
	// router.Use(cors.New(corsConfig))

	// router.Use(ginlogrus.Logger(std.Out), gin.Recovery())

	if debugMode {
		router.Use(requestLoggerMiddleware())
	}

	reactApp := ui.New(cfg, userStorer, codeConnector)

	app := App{
		router:        router,
		cfg:           cfg,
		docStorer:     docStorer,
		userStorer:    userStorer,
		metaStorer:    metaStorer,
		hub:           hub,
		ui:            reactApp,
		codeConnector: codeConnector,
	}

	router.Use(requestLoggerMiddleware())
	app.registerRoutes(router)

	return app
}

func accessDenied(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": message})
}

func badReq(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": message})
}

func internalError(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": message})
}

/// remove remarkable ads
func stripAds(msg string) string {
	br := "<br>--<br>"
	i := strings.Index(msg, br)
	if i > 0 {
		return msg[:i]
	}
	return msg
}

// router.Use(ginlogrus.Logger(std.Out), gin.Recovery())
