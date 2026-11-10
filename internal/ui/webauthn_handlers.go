package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type webAuthnCeremonyResponse struct {
	SessionID string      `json:"sessionId"`
	Options   interface{} `json:"publicKey"`
}

type webAuthnFinishRequest struct {
	SessionID    string          `json:"sessionId"`
	Credential   json.RawMessage `json:"credential"`
	Name         string          `json:"name,omitempty"`
}

type webAuthnCredentialView struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

func (app *ReactAppWrapper) webAuthnEnabled() bool {
	return app.cfg != nil && app.cfg.WebAuthn && app.webAuthn != nil && app.webAuthnSessions != nil
}

func (app *ReactAppWrapper) webAuthnStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"enabled": app.webAuthnEnabled()})
}

func (app *ReactAppWrapper) webAuthnRegisterBegin(c *gin.Context) {
	if !app.webAuthnEnabled() {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	uid := userID(c)
	user, err := app.userStorer.GetUser(uid)
	if err != nil || user == nil {
		log.Error(uiLogger, "webauthn register begin: ", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	waUser := model.WebAuthnUser{User: user}
	opts := []webauthn.RegistrationOption{
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementPreferred),
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			ResidentKey:      protocol.ResidentKeyRequirementPreferred,
			UserVerification: protocol.VerificationPreferred,
		}),
	}
	creation, session, err := app.webAuthn.BeginRegistration(waUser, opts...)
	if err != nil {
		log.Error(uiLogger, "webauthn BeginRegistration: ", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sid := app.webAuthnSessions.Put(session)
	c.JSON(http.StatusOK, webAuthnCeremonyResponse{
		SessionID: sid,
		Options:   creation.Response,
	})
}

func (app *ReactAppWrapper) webAuthnRegisterFinish(c *gin.Context) {
	if !app.webAuthnEnabled() {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	var req webAuthnFinishRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.SessionID == "" || len(req.Credential) == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	session, ok := app.webAuthnSessions.Take(req.SessionID)
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid or expired session"})
		return
	}
	uid := userID(c)
	user, err := app.userStorer.GetUser(uid)
	if err != nil || user == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	parsed, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(req.Credential))
	if err != nil {
		log.Error(uiLogger, "webauthn parse register: ", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid credential"})
		return
	}
	cred, err := app.webAuthn.CreateCredential(model.WebAuthnUser{User: user}, session, parsed)
	if err != nil {
		log.Error(uiLogger, "webauthn CreateCredential: ", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "Passkey"
	}
	user.WebAuthnCredentials = append(user.WebAuthnCredentials, model.FromLibraryCredential(cred, name))
	user.UpdatedAt = time.Now()
	if err := app.userStorer.UpdateUser(user); err != nil {
		log.Error(uiLogger, "webauthn save credential: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":   model.CredentialIDBase64(cred.ID),
		"name": name,
	})
}

func (app *ReactAppWrapper) webAuthnListCredentials(c *gin.Context) {
	if !app.webAuthnEnabled() {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	uid := userID(c)
	user, err := app.userStorer.GetUser(uid)
	if err != nil || user == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	out := make([]webAuthnCredentialView, 0, len(user.WebAuthnCredentials))
	for _, cred := range user.WebAuthnCredentials {
		out = append(out, webAuthnCredentialView{
			ID:        model.CredentialIDBase64(cred.ID),
			Name:      cred.Name,
			CreatedAt: cred.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, out)
}

func (app *ReactAppWrapper) webAuthnDeleteCredential(c *gin.Context) {
	if !app.webAuthnEnabled() {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	credID, err := model.ParseCredentialIDBase64(c.Param("id"))
	if err != nil || len(credID) == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	uid := userID(c)
	user, err := app.userStorer.GetUser(uid)
	if err != nil || user == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if !user.RemoveWebAuthnCredential(credID) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	user.UpdatedAt = time.Now()
	if err := app.userStorer.UpdateUser(user); err != nil {
		log.Error(uiLogger, "webauthn delete credential: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}

func (app *ReactAppWrapper) webAuthnLoginBegin(c *gin.Context) {
	if !app.webAuthnEnabled() {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	assertion, session, err := app.webAuthn.BeginDiscoverableLogin(
		webauthn.WithUserVerification(protocol.VerificationPreferred),
	)
	if err != nil {
		log.Error(uiLogger, "webauthn BeginDiscoverableLogin: ", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sid := app.webAuthnSessions.Put(session)
	c.JSON(http.StatusOK, webAuthnCeremonyResponse{
		SessionID: sid,
		Options:   assertion.Response,
	})
}

func (app *ReactAppWrapper) webAuthnLoginFinish(c *gin.Context) {
	if !app.webAuthnEnabled() {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	var req webAuthnFinishRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.SessionID == "" || len(req.Credential) == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	session, ok := app.webAuthnSessions.Take(req.SessionID)
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid or expired session"})
		return
	}
	parsed, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(req.Credential))
	if err != nil {
		log.Error(uiLogger, "webauthn parse login: ", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid credential"})
		return
	}
	handler := func(rawID, userHandle []byte) (webauthn.User, error) {
		users, err := app.userStorer.GetUsers()
		if err != nil {
			return nil, err
		}
		u := model.FindUserByWebAuthnHandle(users, userHandle)
		if u == nil {
			return nil, protocol.ErrBadRequest.WithDetails("user not found")
		}
		return model.WebAuthnUser{User: u}, nil
	}
	waUser, cred, err := app.webAuthn.ValidatePasskeyLogin(handler, session, parsed)
	if err != nil {
		log.Warn(uiLogger, "webauthn login failed: ", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	wrapped, ok := waUser.(model.WebAuthnUser)
	if !ok || wrapped.User == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	user := wrapped.User
	if user.UpdateWebAuthnCredentialSignCount(cred.ID, cred) {
		user.UpdatedAt = time.Now()
		if err := app.userStorer.UpdateUser(user); err != nil {
			log.Warn(uiLogger, "webauthn persist sign count: ", err)
		}
	}

	tokenString, expiresAfter, err := app.issueWebTokenForUser(user, uuid.NewString())
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(cookieName, tokenString, int(expiresAfter.Seconds()), "/", "", app.cfg.HTTPSCookie, true)
	c.String(http.StatusOK, tokenString)
}
