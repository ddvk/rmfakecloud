package ui

import (
	"net/http"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type serverSettingsResponse struct {
	RmcSrc    string `json:"rmcSrc"`
	RmcSrcEnv string `json:"rmcSrcEnv"`
}

func (app *ReactAppWrapper) getServerSettings(c *gin.Context) {
	if app.cfg == nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	s := app.cfg.GetServerSettings()
	c.JSON(http.StatusOK, serverSettingsResponse{
		RmcSrc:    s.RmcSrc,
		RmcSrcEnv: config.EnvRMCSrc,
	})
}

func (app *ReactAppWrapper) updateServerSettings(c *gin.Context) {
	if app.cfg == nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	var req config.ServerSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		badReq(c, err.Error())
		return
	}
	if err := app.cfg.UpdateServerSettings(req); err != nil {
		log.Warn(uiLogger, "update server settings: ", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, viewmodel.NewErrorResponse(err.Error()))
		return
	}
	s := app.cfg.GetServerSettings()
	c.JSON(http.StatusOK, serverSettingsResponse{
		RmcSrc:    s.RmcSrc,
		RmcSrcEnv: config.EnvRMCSrc,
	})
}
