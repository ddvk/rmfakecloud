package app

import (
	"net/http"
	"time"

	"github.com/ddvk/rmfakecloud/internal/app/passcodestore"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const passcodeLog = "[passcode] "

func (app *App) passcodeReset(c *gin.Context) {
	requestID := c.Param("uuid")
	if requestID == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	claims, err := app.getUserClaims(c)
	if err != nil {
		log.Warn(passcodeLog, err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	deviceName := claims.DeviceDesc
	if deviceName == "" {
		deviceName = "reMarkable"
	}

	now := time.Now().UTC()
	req := messages.PasscodeReset{
		DeviceID:   claims.DeviceID,
		DeviceName: deviceName,
		RequestID:  requestID,
		Created:    now,
		Expires:    now.Add(passcodestore.ResetTTL),
		Approved:   false,
	}

	if err := app.passcodeStore.Create(userID(c), req); err != nil {
		log.Warn(passcodeLog, err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	log.Infof("%screated reset request %s for device %s", passcodeLog, requestID, claims.DeviceID)
	c.Status(http.StatusCreated)
}

func (app *App) getPasscodeReset(c *gin.Context) {
	requestID := c.Param("uuid")
	if requestID == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	uid := userID(c)
	item, err := app.passcodeStore.Get(requestID)
	if err != nil || item.UID != uid {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, item.PasscodeReset)
}
