package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const passcodeLog = "[ui-passcode] "

func (app *ReactAppWrapper) listPasscodeResets(c *gin.Context) {
	uid := userID(c)
	list := app.passcodeStore.ListForUser(uid)
	c.JSON(http.StatusOK, list)
}

func (app *ReactAppWrapper) dismissPasscodeReset(c *gin.Context) {
	uid := userID(c)
	requestID := c.Param("uuid")
	if requestID == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if err := app.passcodeStore.Delete(uid, requestID); err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	log.Infof("%sdismissed reset request %s", passcodeLog, requestID)
	c.Status(http.StatusOK)
}

func (app *ReactAppWrapper) approvePasscodeReset(c *gin.Context) {
	uid := userID(c)
	requestID := c.Param("uuid")
	if requestID == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	reset, err := app.passcodeStore.Approve(uid, requestID)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	app.h.NotifyPasscodeReset(uid, reset.DeviceID, reset.DeviceName, reset.RequestID)
	log.Infof("%sapproved reset request %s for device %s", passcodeLog, reset.RequestID, reset.DeviceID)
	c.Status(http.StatusOK)
}
