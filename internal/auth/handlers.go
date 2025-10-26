package auth

import (
	"encoding/json"
	"errors"
	"html"
	"net/http"
	"strings"

	apperrors "developer-portal-backend/internal/errors"

	"github.com/gin-gonic/gin"
)

// formatResponseAsJSON converts the response to JSON string for embedding in HTML
func formatResponseAsJSON(response interface{}) string {
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return "{}"
	}
	return string(jsonBytes)
}

// escapeJSString safely escapes a Go string for embedding inside JS string literals.
func escapeJSString(s string) string {
	// basic HTML escape then replace newlines/quotes for safe inline JS
	e := html.EscapeString(s)
	e = strings.ReplaceAll(e, "\n", `\n`)
	e = strings.ReplaceAll(e, "\r", ``)
	return e
}

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	service *AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Start handles GET /api/auth/{provider}/start?env=development
// @Summary Start OAuth authentication
// @Description Initiate OAuth authentication flow with the specified provider
// @Tags authentication
// @Accept json
// @Produce json
// @Param provider path string true "OAuth provider (githubtools or githubwdf)"
// @Param env query string false "Environment (development, staging, production)"
// @Success 302 {string} string "Redirect to OAuth provider authorization URL"
// @Failure 400 {object} map[string]interface{} "Invalid provider or request parameters"
// @Failure 500 {object} map[string]interface{} "Failed to generate authorization URL"
// @Router /api/auth/{provider}/start [get]
func (h *AuthHandler) Start(c *gin.Context) {
	provider := c.Param("provider")
	env := c.DefaultQuery("env", "")

	// Validate provider
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider is required"})
		return
	}

	// Validate supported providers
	if provider != "githubtools" && provider != "githubwdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported provider"})
		return
	}

	// Generate state parameter for OAuth2 security
	state, err := h.service.GenerateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state parameter"})
		return
	}

	// Persist/encode env alongside state (so callback can recover env if needed)
	stateWithEnv := state
	if env != "" {
		stateWithEnv = state + ":" + env
	}

	// Get authorization URL
	authURL, err := h.service.GetAuthURL(provider, env, stateWithEnv)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authorization URL", "details": err.Error()})
		return
	}

	// Redirect to GHES OAuth authorization URL
	c.Redirect(http.StatusFound, authURL)
}

// HandlerFrame handles GET /api/auth/{provider}/handler/frame?env=development
// Regular-token mode: posts { type: 'authorization_response', response: { accessToken, tokenType, expiresInSeconds, scope, profile{...} } } to the opener and closes.
// @Summary Handle OAuth callback
// @Description Handle OAuth callback from provider and return authentication result in HTML frame
// @Tags authentication
// @Accept json
// @Produce text/html
// @Param provider path string true "OAuth provider (githubtools or githubwdf)"
// @Param code query string true "OAuth authorization code from provider"
// @Param state query string true "OAuth state parameter for security"
// @Param env query string false "Environment (development, staging, production)"
// @Param error query string false "OAuth error parameter from provider"
// @Param error_description query string false "OAuth error description from provider"
// @Success 200 {string} string "HTML page that posts authentication result to opener window"
// @Failure 400 {object} map[string]interface{} "Invalid request parameters"
// @Router /api/auth/{provider}/handler/frame [get]
func (h *AuthHandler) HandlerFrame(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")
	env := c.DefaultQuery("env", "")

	// OAuth errors from provider
	if errorParam != "" {
		errorDescription := c.Query("error_description")
		errorHTML := `<!doctype html><html><body><script>
(function(){
  var msg = { type: "authorization_response", error: { name: "OAuthError", message: "` + escapeJSString(errorParam) + `: ` + escapeJSString(errorDescription) + `" } };
  try { if (window.opener) window.opener.postMessage(msg, "*"); } finally { window.close(); }
})();
</script></body></html>`
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, errorHTML)
		return
	}

	// Validate params
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider is required"})
		return
	}
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is required"})
		return
	}
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "State parameter is required"})
		return
	}

	// Extract environment from state if present
	actualState := state
	stateEnv := ""
	if strings.Contains(state, ":") {
		parts := strings.SplitN(state, ":", 2)
		actualState = parts[0]
		stateEnv = parts[1]
	}
	if env == "" && stateEnv != "" {
		env = stateEnv
	}

	// Service callback – may return various shapes; we'll normalize in JS
	serviceResp, err := h.service.HandleCallback(c.Request.Context(), provider, env, code, actualState)
	if err != nil {
		errorHTML := `<!doctype html><html><body><script>
(function(){
  var msg = { type: "authorization_response", error: { name: "Error", message: "` + escapeJSString(err.Error()) + `" } };
  try { if (window.opener) window.opener.postMessage(msg, "*"); } finally { window.close(); }
})();
</script></body></html>`
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, errorHTML)
		return
	}

	// Set session cookies for later use by refresh endpoint
	c.SetCookie("auth_token", serviceResp.AccessToken, 3600, "/", "", false, true)           // httpOnly for security
	c.SetCookie("refresh_token", serviceResp.RefreshToken, 30*24*3600, "/", "", false, true) // 30 days, httpOnly

	// Store user profile in cookie (JSON encoded)
	profileJSON, _ := json.Marshal(serviceResp.Profile)
	c.SetCookie("user_profile", string(profileJSON), 3600, "/", "", false, false) // not httpOnly so JS can read

	// Embed the raw service response and normalize to the regular-token payload in the browser.
	raw := formatResponseAsJSON(serviceResp)

	successHTML := `<!doctype html><html><body><script>
(function(){
  var src = ` + raw + ` || {};
  // Normalize various possible shapes into:
  // { accessToken, tokenType, expiresInSeconds, scope, profile{ login,email,name,avatarUrl } }
  function toStr(v){ return (v==null)? "" : (Array.isArray(v)? v.join(" ") : String(v)); }
  var accessToken = src.accessToken || src.access_token || src.token || "";
  var tokenType   = src.tokenType || src.token_type || "bearer";
  var expires     = src.expiresInSeconds || src.expires_in || 0;
  var scopeStr    = src.scope || src.scopes || "";
  scopeStr = Array.isArray(scopeStr) ? scopeStr.join(" ") : toStr(scopeStr);

  // profile could be under src.profile or src.user
  var p = src.profile || src.user || {};
  var profile = {
    login:     p.login     || p.username || "",
    email:     p.email     || "",
    name:      p.name      || p.displayName || "",
    avatarUrl: p.avatarUrl || p.avatar_url || p.picture || ""
  };

  var resp = {
    accessToken: accessToken,
    tokenType: tokenType,
    expiresInSeconds: Number(expires) || 0,
    scope: scopeStr,
    profile: profile
  };

  var message = { type: "authorization_response", response: resp };
  try { if (window.opener) window.opener.postMessage(message, "*"); } finally { window.close(); }
})();
</script></body></html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, successHTML)
}

// Refresh handles GET /api/auth/{provider}/refresh?env=development&refresh_token=...
// Regular-token mode JSON response. If your service returns a non-standard shape,
// we normalize it into { accessToken, tokenType, expiresInSeconds, scope, profile{...} }.
// @Summary Refresh authentication token
// @Description Refresh or validate authentication token using refresh token, Authorization header, or session cookies
// @Tags authentication
// @Accept json
// @Produce json
// @Param provider path string true "OAuth provider (githubtools or githubwdf)"
// @Param env query string false "Environment (development, staging, production)"
// @Param refresh_token query string false "Refresh token to use for getting new access token"
// @Param Authorization header string false "Bearer token for validation" example("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
// @Success 200 {object} AuthRefreshResponse "Successfully refreshed token"
// @Failure 400 {object} map[string]interface{} "Invalid provider"
// @Failure 401 {object} map[string]interface{} "Authentication required or token invalid"
// @Failure 500 {object} map[string]interface{} "Token refresh failed"
// @Router /api/auth/{provider}/refresh [get]
func (h *AuthHandler) Refresh(c *gin.Context) {
	provider := c.Param("provider")
	_ = c.DefaultQuery("env", "development") // env parameter acknowledged but not used in this context
	refreshToken := c.Query("refresh_token")

	// Validate provider
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider is required"})
		return
	}
	if provider != "githubtools" && provider != "githubwdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported provider"})
		return
	}

	// If no refresh_token, check multiple sources for session information
	if strings.TrimSpace(refreshToken) == "" {
		// 1. Check Authorization header first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Extract JWT token from Bearer header
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString != authHeader {
				// Try to get current session info from JWT
				claims, err := h.service.ValidateJWT(tokenString)
				if err == nil {
					// Generate a new JWT token for the current session
					userProfile := &UserProfile{
						ID:       claims.UserID,
						Username: claims.Username,
						Email:    claims.Email,
						MemberID: h.service.GetMemberIDByEmail(claims.Email),
					}

					newJWT, err := h.service.GenerateJWT(userProfile, claims.Provider, claims.Environment)
					if err == nil {
						profileResponse := gin.H{
							"login":     claims.Username,
							"email":     claims.Email,
							"name":      claims.Username,
							"avatarUrl": "",
						}
						if userProfile.MemberID != nil {
							profileResponse["memberId"] = *userProfile.MemberID
						}
						c.JSON(http.StatusOK, gin.H{
							"accessToken":      newJWT,
							"tokenType":        "bearer",
							"expiresInSeconds": 3600,
							"scope":            "",
							"profile":          profileResponse,
							"valid":            true,
						})
						return
					}
				}
			}
		}

		// 2. Check session cookies
		authTokenCookie, err1 := c.Cookie("auth_token")
		refreshTokenCookie, err2 := c.Cookie("refresh_token")

		if err1 == nil && authTokenCookie != "" {
			// Validate the JWT token from cookie
			claims, err := h.service.ValidateJWT(authTokenCookie)
			if err == nil {
				// Generate a new JWT token for the current session
				userProfile := &UserProfile{
					ID:       claims.UserID,
					Username: claims.Username,
					Email:    claims.Email,
					MemberID: h.service.GetMemberIDByEmail(claims.Email),
				}

				newJWT, err := h.service.GenerateJWT(userProfile, claims.Provider, claims.Environment)
				if err == nil {
					profileResponse := gin.H{
						"login":     claims.Username,
						"email":     claims.Email,
						"name":      claims.Username,
						"avatarUrl": "",
					}
					if userProfile.MemberID != nil {
						profileResponse["memberId"] = *userProfile.MemberID
					}
					c.JSON(http.StatusOK, gin.H{
						"accessToken":      newJWT,
						"tokenType":        "bearer",
						"expiresInSeconds": 3600,
						"scope":            "",
						"profile":          profileResponse,
						"valid":            true,
					})
					return
				}
			}
		} else if err2 == nil && refreshTokenCookie != "" {
			// Try to use the refresh token from cookie
			refreshed, err := h.service.RefreshToken(refreshTokenCookie)
			if err == nil {
				// Return the refreshed token data
				profileResponse := gin.H{
					"login":     refreshed.Profile.Username,
					"email":     refreshed.Profile.Email,
					"name":      refreshed.Profile.Name,
					"avatarUrl": refreshed.Profile.AvatarURL,
				}
				if refreshed.Profile.MemberID != nil {
					profileResponse["memberId"] = *refreshed.Profile.MemberID
				}
				c.JSON(http.StatusOK, gin.H{
					"accessToken":      refreshed.AccessToken,
					"tokenType":        refreshed.TokenType,
					"expiresInSeconds": refreshed.ExpiresIn,
					"scope":            "",
					"profile":          profileResponse,
					"valid":            true,
				})
				return
			}
		}

		// No valid session found, return 401 Unauthorized
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"details": "No valid session found. Please authenticate first.",
		})
		return
	}

	// Use your existing service to refresh
	refreshed, err := h.service.RefreshToken(refreshToken)
	if err != nil {
		// Return 401 for invalid/expired tokens, 403 for authorization failures
		if errors.Is(err, apperrors.ErrInvalidRefreshToken) || errors.Is(err, apperrors.ErrRefreshTokenExpired) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token refresh failed", "details": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token refresh failed", "details": err.Error()})
		}
		return
	}

	// Normalize the service response into regular-token shape
	// We do this by round-tripping through a generic map.
	var m map[string]interface{}
	b, _ := json.Marshal(refreshed)
	_ = json.Unmarshal(b, &m)

	accessToken := firstString(m, "accessToken", "access_token", "token")
	tokenType := firstString(m, "tokenType", "token_type")
	if tokenType == "" {
		tokenType = "bearer"
	}
	expires := firstInt64(m, "expiresInSeconds", "expires_in")
	scope := firstString(m, "scope")
	if scope == "" {
		if scopes, ok := m["scopes"].([]interface{}); ok {
			scope = joinInterfaces(scopes, " ")
		}
	}
	// profile/user block
	profAny := firstAny(m, "profile", "user")
	var profile map[string]interface{}
	if p, ok := profAny.(map[string]interface{}); ok {
		profile = p
	} else {
		profile = map[string]interface{}{}
	}
	login := firstString(profile, "login", "username")
	email := firstString(profile, "email")
	name := firstString(profile, "name", "displayName")
	avatar := firstString(profile, "avatarUrl", "avatar_url", "picture")
	memberID := firstString(profile, "memberId", "member_id")

	profileResponse := gin.H{
		"login":     login,
		"email":     email,
		"name":      name,
		"avatarUrl": avatar,
	}
	if memberID != "" {
		profileResponse["memberId"] = memberID
	}

	out := gin.H{
		"accessToken":      accessToken,
		"tokenType":        tokenType,
		"expiresInSeconds": expires,
		"scope":            scope,
		"profile":          profileResponse,
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, out)
}

// Logout handles POST /api/auth/{provider}/logout?env=development
// @Summary Logout user
// @Description Logout user and invalidate authentication session
// @Tags authentication
// @Accept json
// @Produce json
// @Param provider path string true "OAuth provider (githubtools or githubwdf)"
// @Param env query string false "Environment (development, staging, production)"
// @Success 200 {object} AuthLogoutResponse "Successfully logged out"
// @Failure 400 {object} map[string]interface{} "Invalid provider"
// @Failure 500 {object} map[string]interface{} "Logout failed"
// @Router /api/auth/{provider}/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	provider := c.Param("provider")
	_ = c.DefaultQuery("env", "") // env parameter acknowledged but not used in current implementation

	// Validate provider
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider is required"})
		return
	}
	if provider != "githubtools" && provider != "githubwdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported provider"})
		return
	}

	if err := h.service.Logout(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Logout failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// ValidateToken is a helper endpoint to validate JWT tokens (not part of Backstage spec but useful for debugging)
// @Summary Validate JWT token
// @Description Validate JWT token and return token claims
// @Tags authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token to validate" example("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
// @Success 200 {object} AuthValidateResponse "Token is valid with claims"
// @Failure 401 {object} map[string]interface{} "Authorization header required or token invalid"
// @Router /api/auth/validate [post]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	// Extract token from Bearer header
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		return
	}

	// Validate token
	claims, err := h.service.ValidateJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true, "claims": claims})
}

// ---------- tiny helpers for Refresh normalization ----------

func firstString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func firstInt64(m map[string]interface{}, keys ...string) int64 {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case float64:
				return int64(t)
			case int64:
				return t
			case int:
				return int64(t)
			case json.Number:
				if iv, err := t.Int64(); err == nil {
					return iv
				}
			}
		}
	}
	return 0
}

func firstAny(m map[string]interface{}, keys ...string) interface{} {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return v
		}
	}
	return nil
}

func joinInterfaces(arr []interface{}, sep string) string {
	sb := strings.Builder{}
	for i, v := range arr {
		if i > 0 {
			sb.WriteString(sep)
		}
		switch s := v.(type) {
		case string:
			sb.WriteString(s)
		default:
			b, _ := json.Marshal(s)
			sb.Write(b)
		}
	}
	return sb.String()
}
