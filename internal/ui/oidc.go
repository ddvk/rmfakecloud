package ui

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const (
	oidcStateCookie    = "oidc_state"
	oidcNonceCookie    = "oidc_nonce"
	oidcVerifierCookie = "oidc_pkce_verifier"
	oidcCookieMaxAge   = 300
	// oidcSuccessPath is the frontend route the callback redirects to on success.
	// The frontend must register a matching <Route path="/oidc-success"> to handle it.
	oidcSuccessPath = "/oidc-success"
)

var (
	errNoUserID         = errors.New("no userid available: configured claim not found or empty")
	errEmailNotVerified = errors.New("email not verified")
)

type oidcUserIdentity struct {
	Value     string
	ClaimName string
}

func newOIDCUserIdentity(value, claimName string) oidcUserIdentity {
	if claimName == "email" {
		// Email identities must compare case-insensitively so repeated logins resolve to the same user.
		value = strings.ToLower(value)
	}
	return oidcUserIdentity{Value: value, ClaimName: claimName}
}

func (identity oidcUserIdentity) usesEmail() bool {
	return identity.ClaimName == "email"
}

// oidcClaims holds the standard OIDC claims read from the ID token.
// The configurable userid and admin claims may be arbitrary dotted paths;
// those still require the raw map passed to extractClaimPath.
type oidcClaims struct {
	Nonce             string `json:"nonce"`
	Email             string `json:"email"`
	EmailVerified     any    `json:"email_verified"` // bool or string "true"/"false" depending on provider
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	PreferredUsername string `json:"preferred_username"`
}

// randomURLSafeString generates a cryptographically random base64url-encoded string from n bytes.
func randomURLSafeString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// setOIDCCookie writes a short-lived SameSite=Lax cookie for OIDC flow state.
// Lax is required so the cookie survives the cross-site top-level redirect back from the IdP.
func (app *ReactAppWrapper) setOIDCCookie(c *gin.Context, name, value string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, value, oidcCookieMaxAge, "/", "", app.cfg.HTTPSCookie, true)
}

// clearOIDCCookie removes an OIDC flow cookie using the same attributes.
func (app *ReactAppWrapper) clearOIDCCookie(c *gin.Context, name string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, "", -1, "/", "", app.cfg.HTTPSCookie, true)
}

// extractClaimPath traverses a dotted path (e.g. "realm_access.roles") in a raw claims map.
func extractClaimPath(raw map[string]any, path string) (any, bool) {
	parts := strings.SplitN(path, ".", 2)
	val, ok := raw[parts[0]]
	if !ok {
		return nil, false
	}
	if len(parts) == 1 {
		return val, true
	}
	nested, ok := val.(map[string]any)
	if !ok {
		return nil, false
	}
	return extractClaimPath(nested, parts[1])
}

// claimHasValue checks if a claim value (string, []any, or []string) contains expected.
func claimHasValue(value any, expected string) bool {
	switch v := value.(type) {
	case string:
		return v == expected
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s == expected {
				return true
			}
		}
	case []string:
		for _, s := range v {
			if s == expected {
				return true
			}
		}
	}
	return false
}

// claimIsTrue interprets an OIDC boolean claim that may arrive as a bool or as a
// string ("true"/"false"), as allowed by different providers.
func claimIsTrue(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(strings.TrimSpace(v), "true")
	}
	return false
}

// oidcBegin starts the OIDC authorization code flow with PKCE, state, and nonce.
func (app *ReactAppWrapper) oidcBegin(c *gin.Context) {
	state, err := randomURLSafeString(32)
	if err != nil {
		log.Error("[oidc] failed to generate state: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	nonce, err := randomURLSafeString(32)
	if err != nil {
		log.Error("[oidc] failed to generate nonce: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	pkceVerifier, err := randomURLSafeString(32)
	if err != nil {
		log.Error("[oidc] failed to generate PKCE verifier: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	app.setOIDCCookie(c, oidcStateCookie, state)
	app.setOIDCCookie(c, oidcNonceCookie, nonce)
	app.setOIDCCookie(c, oidcVerifierCookie, pkceVerifier)

	authURL := app.oauth2Config.AuthCodeURL(
		state,
		gooidc.Nonce(nonce),
		oauth2.S256ChallengeOption(pkceVerifier),
	)
	c.Redirect(http.StatusFound, authURL)
}

// exchangeAndVerifyToken exchanges the authorization code for tokens and returns verified claims.
// Returns the raw claims map (for configurable dotted-path lookups), a typed oidcClaims
// struct (for standard fields), and true on success; false on error (caller already got an HTTP response).
func (app *ReactAppWrapper) exchangeAndVerifyToken(c *gin.Context, ctx context.Context, code, pkceVerifier string) (map[string]any, oidcClaims, bool) {
	// Exchange authorization code for tokens, presenting the PKCE verifier
	oauth2Token, err := app.oauth2Config.Exchange(ctx, code, oauth2.VerifierOption(pkceVerifier))
	if err != nil {
		log.Error("[oidc] token exchange failed: ", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token exchange failed"})
		return nil, oidcClaims{}, false
	}

	// Extract raw ID token string
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing id_token in provider response"})
		return nil, oidcClaims{}, false
	}

	// Verify ID token signature, expiry, issuer, and audience
	idTokenVerifier := app.oidcProvider.Verifier(&gooidc.Config{ClientID: app.cfg.OIDC.ClientID})
	idToken, err := idTokenVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		log.Warn("[oidc] ID token verification failed: ", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "ID token verification failed"})
		return nil, oidcClaims{}, false
	}

	// Deserialize into typed struct for standard fields
	var claims oidcClaims
	if err := idToken.Claims(&claims); err != nil {
		log.Error("[oidc] failed to extract typed claims: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return nil, oidcClaims{}, false
	}
	claims.Nonce = idToken.Nonce

	// Deserialize into raw map for configurable dotted-path claim lookups
	var rawClaims map[string]any
	if err := idToken.Claims(&rawClaims); err != nil {
		log.Error("[oidc] failed to extract raw claims: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return nil, oidcClaims{}, false
	}

	return rawClaims, claims, true
}

// resolveOIDCIdentity extracts and validates the user identity from OIDC claims.
// It tries the configured claim first (using the raw map for dotted-path support),
// falls back to the typed email claim if not found, then enforces email verification
// when the identity is email-based.
// cfg.OIDCUserIDClaim is always non-empty (FromEnv applies the default).
func (app *ReactAppWrapper) resolveOIDCIdentity(rawClaims map[string]any, claims oidcClaims) (oidcUserIdentity, error) {
	userIDClaimName := app.cfg.OIDC.UserIDClaim

	// Try to extract the configured claim (may be a dotted path like "realm_access.roles")
	if claimVal, ok := extractClaimPath(rawClaims, userIDClaimName); ok {
		if strVal, ok := claimVal.(string); ok {
			if userIDValue := strings.TrimSpace(strVal); userIDValue != "" {
				identity := newOIDCUserIdentity(userIDValue, userIDClaimName)
				return identity, app.validateEmailIdentity(identity, claims)
			}
		}
	}

	// If the primary claim does not yield a userid, fall back to the email claim.
	if userIDClaimName != "email" {
		if userIDValue := strings.TrimSpace(claims.Email); userIDValue != "" {
			identity := newOIDCUserIdentity(userIDValue, "email")
			return identity, app.validateEmailIdentity(identity, claims)
		}
	}

	return oidcUserIdentity{ClaimName: userIDClaimName}, errNoUserID
}

// validateEmailIdentity enforces email verification for email-based identities.
// Without this, a provider that lets users set an arbitrary unverified email
// could be used to provision or take over an account for another address.
func (app *ReactAppWrapper) validateEmailIdentity(identity oidcUserIdentity, claims oidcClaims) error {
	if identity.usesEmail() && !app.cfg.OIDC.AllowUnverifiedEmail && !claimIsTrue(claims.EmailVerified) {
		return errEmailNotVerified
	}
	return nil
}

// evaluateOIDCAdminStatus returns a *bool reflecting the result of the configured
// admin claim check. nil means the admin claim is not configured; a non-nil pointer
// holds the evaluated value. Callers must not touch a user's admin flag when nil.
func (app *ReactAppWrapper) evaluateOIDCAdminStatus(rawClaims map[string]any) *bool {
	if app.cfg.OIDC.AdminClaim == "" || app.cfg.OIDC.AdminClaimValue == "" {
		return nil
	}
	isAdmin := false
	if claimVal, ok := extractClaimPath(rawClaims, app.cfg.OIDC.AdminClaim); ok {
		isAdmin = claimHasValue(claimVal, app.cfg.OIDC.AdminClaimValue)
	}
	return &isAdmin
}

// provisionNewUser creates and registers a new OIDC-provisioned user.
func (app *ReactAppWrapper) provisionNewUser(userIDValue string, claims oidcClaims, isAdmin bool) (*model.User, error) {
	randomPassword, err := model.GenPassword()
	if err != nil {
		log.Error("[oidc] failed to generate password for provisioning: ", err)
		return nil, err
	}

	user, err := model.NewUser(userIDValue, randomPassword)
	if err != nil {
		log.Error("[oidc] failed to build user: ", err)
		return nil, err
	}
	if email := strings.TrimSpace(claims.Email); email != "" {
		user.Email = model.NormalizeUserID(email)
		user.EmailVerified = claimIsTrue(claims.EmailVerified)
	}

	// Populate user profile from OIDC claims
	if claims.Name != "" {
		user.Name = claims.Name
	}
	if claims.GivenName != "" {
		user.GivenName = claims.GivenName
	}
	if claims.FamilyName != "" {
		user.FamilyName = claims.FamilyName
	}
	if claims.PreferredUsername != "" {
		user.Nickname = claims.PreferredUsername
	}

	user.IsAdmin = isAdmin

	if err := app.userStorer.RegisterUser(user); err != nil {
		log.Error("[oidc] failed to register provisioned user: ", err)
		return nil, err
	}

	return user, nil
}

// getOrProvisionUser looks up or auto-provisions an OIDC user.
// adminStatus is nil when no admin claim is configured; in that case the existing
// user's admin flag is left untouched. A non-nil pointer carries the evaluated result.
func (app *ReactAppWrapper) getOrProvisionUser(userKey string, identity oidcUserIdentity, claims oidcClaims, adminStatus *bool) (*model.User, error) {
	isAdmin := adminStatus != nil && *adminStatus
	user, err := app.userStorer.GetUser(userKey)
	if err != nil {
		if !errors.Is(err, storage.ErrUserNotFound) {
			log.Error("[oidc] storage error looking up user: ", err)
			return nil, err
		}
		// User not found — provision new user
		var newUser *model.User
		newUser, err = app.provisionNewUser(identity.Value, claims, isAdmin)
		if err != nil {
			return nil, err
		}
		log.Info("[oidc] provisioned new user: ", userKey, " (claim=\"", identity.ClaimName, "\", value=\"", identity.Value, "\") admin=", isAdmin)
		return newUser, nil
	}

	// Existing user — only re-evaluate admin when an admin claim is configured,
	// otherwise an OIDC login would silently strip admin from an existing user.
	if adminStatus != nil && user.IsAdmin != *adminStatus {
		user.IsAdmin = *adminStatus
		if err := app.userStorer.UpdateUser(user); err != nil {
			log.Error("[oidc] failed to update user admin status: ", err)
			return nil, err
		}
		log.Info("[oidc] updated admin status for ", userKey, " to ", *adminStatus)
	}

	return user, nil
}

// completeOIDCLogin resolves the user identity from verified claims, gets or provisions
// the account, issues a session cookie, and redirects to the success page.
// It handles all identity/provisioning concerns after the protocol-level checks in oidcCallback.
func (app *ReactAppWrapper) completeOIDCLogin(c *gin.Context, rawClaims map[string]any, claims oidcClaims) {
	// Extract, validate and normalise the user identity from claims
	identity, err := app.resolveOIDCIdentity(rawClaims, claims)
	if err != nil {
		switch {
		case errors.Is(err, errNoUserID):
			log.Warn("[oidc] no userid available: configured claim '", identity.ClaimName, "' not found or empty")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "userid claim not found or empty"})
		case errors.Is(err, errEmailNotVerified):
			log.Warn("[oidc] rejected login: email not verified for ", identity.Value)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "email not verified"})
		default:
			log.Error("[oidc] resolveOIDCIdentity error: ", err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}

	// The stored user id is the sanitized userid; use the same key for lookup and
	// provisioning so subsequent logins resolve to the same account.
	userKey := model.NormalizeUserID(identity.Value)

	// Determine admin solely from the configured role claim; re-evaluated on every login
	adminStatus := app.evaluateOIDCAdminStatus(rawClaims)

	// Get or provision user (lookup existing, or create new)
	user, err := app.getOrProvisionUser(userKey, identity, claims, adminStatus)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Issue session and redirect to success page
	if _, err := app.issueWebSession(c, user); err != nil {
		log.Error("[oidc] failed to issue session: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Redirect(http.StatusFound, oidcSuccessPath)
}

// oidcCallback handles the redirect back from the OIDC provider.
// It validates the protocol-level cookies (state, nonce, PKCE), exchanges the
// authorization code, then delegates identity and provisioning to completeOIDCLogin.
func (app *ReactAppWrapper) oidcCallback(c *gin.Context) {
	ctx := c.Request.Context()

	// Provider-side error — sanitize before returning to client
	if errParam := c.Query("error"); errParam != "" {
		log.Warn("[oidc] provider error: ", errParam, " — ", c.Query("error_description"))
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication failed at identity provider"})
		return
	}

	// Validate state cookie
	stateCookie, err := c.Cookie(oidcStateCookie)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing state cookie"})
		return
	}
	if subtle.ConstantTimeCompare([]byte(stateCookie), []byte(c.Query("state"))) != 1 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "authentication failed"})
		return
	}

	// Read nonce and PKCE verifier before clearing any cookies
	nonceCookie, err := c.Cookie(oidcNonceCookie)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing nonce cookie"})
		return
	}
	pkceVerifier, err := c.Cookie(oidcVerifierCookie)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing pkce verifier cookie"})
		return
	}

	// Clear all OIDC flow cookies before any external calls
	app.clearOIDCCookie(c, oidcStateCookie)
	app.clearOIDCCookie(c, oidcNonceCookie)
	app.clearOIDCCookie(c, oidcVerifierCookie)

	// Exchange authorization code for tokens and verify claims
	rawClaims, claims, ok := app.exchangeAndVerifyToken(c, ctx, c.Query("code"), pkceVerifier)
	if !ok {
		return // Error already written to response
	}

	// Verify nonce with constant-time comparison
	if subtle.ConstantTimeCompare([]byte(claims.Nonce), []byte(nonceCookie)) != 1 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "authentication failed"})
		return
	}

	app.completeOIDCLogin(c, rawClaims, claims)
}

// meHandler returns the current authenticated user's profile as JSON.
// Used by the frontend after an OIDC redirect to hydrate localStorage.
func (app *ReactAppWrapper) meHandler(c *gin.Context) {
	uid := userID(c)
	user, err := app.userStorer.GetUser(uid)
	if err != nil {
		log.Error("[me] user not found: ", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	scopes := ""
	if user.Sync15 {
		scopes = isSync15Key
	}
	roles := []string{"User"}
	if user.IsAdmin {
		roles = []string{AdminRole}
	}
	c.JSON(http.StatusOK, gin.H{
		"UserID": user.ID,
		"Email":  user.Email,
		"Scopes": scopes,
		"Roles":  roles,
	})
}
