package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/hwr"
	"github.com/ddvk/rmfakecloud/internal/mqtt"
	"github.com/ddvk/rmfakecloud/internal/storage"

	"github.com/ddvk/rmfakecloud/internal/storage/fs"
	"github.com/ddvk/rmfakecloud/internal/ui"

	"github.com/gin-gonic/gin"
)

const (
	userIDKey      = "UserID"
	deviceIDKey    = "DeviceID"
	syncVersionKey = "SyncVersion"
)

// App web app
type App struct {
	router        *gin.Engine
	cfg           *config.Config
	srv           *http.Server
	docStorer     storage.DocumentStorer
	userStorer    storage.UserStorer
	metaStorer    storage.MetadataStorer
	blobStorer    storage.BlobStorage
	hub           *hub.Hub
	codeConnector CodeConnector
	hwrClient     *hwr.HWRClient
	mqttBroker    *mqtt.Broker
}

// Start starts the app
func (app *App) Start() {
	// configs
	if app.cfg.CloudHost != config.DefaultHost {
		log.Info(config.EnvStorageURL, " (Cloud URL): ", app.cfg.StorageURL)
		log.Info("(Cloud HOST): ", app.cfg.CloudHost)
	}
	log.Info("Data: ", app.cfg.DataDir)
	log.Info("HTTP listening on port: ", app.cfg.Port)

	if !app.cfg.TrustProxy {
		app.router.SetTrustedProxies(nil)
	}

	var tlsConfig *tls.Config
	if app.cfg.Certificate.Certificate != nil {
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{
				app.cfg.Certificate,
			},
		}
		log.Info("TLS certificates configured")
	}

	if app.mqttBroker != nil {
		log.Infof("MQTT listening on port: %s", app.cfg.MQTTPort)
		app.mqttBroker = mqtt.NewBroker(app.cfg.MQTTPort, tlsConfig, app.validateMQTTToken, app.cfg.ICEServers)
		if err := app.mqttBroker.Start(); err != nil {
			log.Errorf("Failed to start MQTT broker: %v", err)
		}
	}

	app.srv = &http.Server{
		Addr:      ":" + app.cfg.Port,
		Handler:   app.router,
		TLSConfig: tlsConfig,
	}

	if tlsConfig != nil {
		log.Info("Starting HTTPS server")
		if err := app.srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	} else {
		log.Info("Starting HTTP server (plain)")
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
	if app.mqttBroker != nil {
		if err := app.mqttBroker.Stop(); err != nil {
			log.Errorf("Error stopping MQTT broker: %v", err)
		}
	}
	if err := app.srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
}


// NewApp constructs an app
func NewApp(cfg *config.Config) App {
	debugMode := log.GetLevel() >= log.DebugLevel
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
	ntfHub := hub.NewHub()
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
		blobStorer:    fsStorage,
		hub:           ntfHub,
		codeConnector: codeConnector,
		hwrClient: &hwr.HWRClient{
			Cfg: cfg,
		},
	}

	app.mqttBroker = mqtt.NewBroker(cfg.MQTTPort, nil, app.validateMQTTToken, cfg.ICEServers)

	app.registerRoutes(router)

	uiApp := ui.New(cfg, fsStorage, codeConnector, ntfHub, fsStorage, fsStorage)
	uiApp.RegisterRoutes(router)

	storageapp := storage.NewApp(cfg, fsStorage, fsStorage, blobStorage)
	storageapp.RegisterRoutes(router)

	return app
}

func badReq(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": message})
}

func internalError(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": message})
}

func (app *App) validateMQTTToken(token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("empty token")
	}

	claims := &UserClaims{}
	err := common.ClaimsFromToken(claims, token, app.cfg.JWTSecretKey)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	if claims.Profile.UserID == "" {
		return "", fmt.Errorf("missing user ID in token")
	}

	if claims.Version != tokenVersion {
		return "", fmt.Errorf("invalid token version")
	}

	userID := common.SanitizeUid(strings.TrimPrefix(claims.Profile.UserID, "auth0|"))
	return userID, nil
}
