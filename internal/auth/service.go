package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// RefreshTokenData stores information about a refresh token
type RefreshTokenData struct {
	UserID      int64     `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	MemberID    *string   `json:"member_id,omitempty"` // ID of member with matching email
	Provider    string    `json:"provider"`
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// MemberRepository defines the interface for member operations needed by auth service
type UserRepository interface {
	GetByEmail(email string) (interface{}, error)
}

// AuthService provides authentication functionality
type AuthService struct {
	config        *AuthConfig
	githubClients map[string]*GitHubClient
	refreshTokens map[string]*RefreshTokenData // In-memory store for refresh tokens
	tokenMutex    sync.RWMutex                 // Protect the refresh token store
	userRepo      UserRepository               // Repository for member lookup
}

// AuthClaims represents JWT token claims
type AuthClaims struct {
	UserID   int64  `json:"user_id" example:"12345"`
	Username string `json:"username" example:"johndoe"`
	Email    string `json:"email" example:"john.doe@example.com"`
	Provider string `json:"provider" example:"githubtools"`
	// Standard JWT fields
	Issuer               string `json:"iss,omitempty" example:"developer-portal-backend"`
	Subject              string `json:"sub,omitempty" example:"12345"`
	Audience             string `json:"aud,omitempty" example:"developer-portal"`
	ExpiresAt            int64  `json:"exp,omitempty" example:"1672531200"`
	IssuedAt             int64  `json:"iat,omitempty" example:"1672527600"`
	jwt.RegisteredClaims `swaggerignore:"true"`
}

// AuthStartResponse represents the response for auth start endpoint
type AuthStartResponse struct {
	URL string `json:"url"`
}

// AuthHandlerResponse represents the response for auth handler endpoint
type AuthHandlerResponse struct {
	AccessToken  string      `json:"accessToken"`
	TokenType    string      `json:"tokenType"`
	ExpiresIn    int64       `json:"expiresIn"`
	RefreshToken string      `json:"refreshToken,omitempty"`
	Profile      UserProfile `json:"profile"`
}

// RefreshTokenRequest represents the request for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// AuthRefreshResponse represents the response from the refresh endpoint
type AuthRefreshResponse struct {
	AccessToken      string      `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType        string      `json:"tokenType" example:"bearer"`
	ExpiresInSeconds int64       `json:"expiresInSeconds" example:"3600"`
	Scope            string      `json:"scope" example:"user:email read:user"`
	Profile          UserProfile `json:"profile"`
	Valid            bool        `json:"valid,omitempty" example:"true"`
}

// AuthLogoutResponse represents the response from the logout endpoint
type AuthLogoutResponse struct {
	Message string `json:"message" example:"Logged out successfully"`
}

// AuthValidateResponse represents the response from the token validation endpoint
type AuthValidateResponse struct {
	Valid  bool        `json:"valid" example:"true"`
	Claims *AuthClaims `json:"claims"`
}

// NewAuthService creates a new authentication service
func NewAuthService(config *AuthConfig, userRepo UserRepository) (*AuthService, error) {
	if err := config.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid auth config: %w", err)
	}

	// Initialize GitHub clients for each provider
	githubClients := make(map[string]*GitHubClient)
	for providerName, providerConfig := range config.Providers {
		githubClients[providerName] = NewGitHubClient(&providerConfig)
	}

	return &AuthService{
		config:        config,
		githubClients: githubClients,
		refreshTokens: make(map[string]*RefreshTokenData),
		tokenMutex:    sync.RWMutex{},
		userRepo:      userRepo,
	}, nil
}

// getMemberIDByEmail looks up a member by email and returns their ID as a string pointer
// Returns nil if member is not found or an error occurs
func (s *AuthService) getMemberIDByEmail(email string) *string {
	if s.userRepo == nil || email == "" {
		return nil
	}

	member, err := s.userRepo.GetByEmail(email)
	if err != nil || member == nil {
		return nil
	}

	// Use reflection to access the ID field from the member struct
	// This works with models.Member which has an ID field of type uuid.UUID
	val := reflect.ValueOf(member)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		idField := val.FieldByName("ID")
		if idField.IsValid() {
			// Convert UUID to string
			idStr := fmt.Sprintf("%v", idField.Interface())
			return &idStr
		}
	}

	return nil
}

// GetMemberIDByEmail is a public wrapper for getMemberIDByEmail
// This allows handlers to look up member ID when needed
func (s *AuthService) GetMemberIDByEmail(email string) *string {
	return s.getMemberIDByEmail(email)
}

// GetAuthURL generates OAuth2 authorization URL
func (s *AuthService) GetAuthURL(provider, state string) (string, error) {
	_, err := s.config.GetProvider(provider)
	if err != nil {
		return "", err
	}

	githubClient, exists := s.githubClients[provider]
	if !exists {
		return "", fmt.Errorf("GitHub client not found for provider %s", provider)
	}

	// Generate callback URL
	callbackURL := fmt.Sprintf("%s/api/auth/%s/handler/frame", s.config.RedirectURL, provider)

	oauth2Config := githubClient.GetOAuth2Config(callbackURL)
	authURL := oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	return authURL, nil
}

// HandleCallback processes OAuth2 callback and returns user information
func (s *AuthService) HandleCallback(ctx context.Context, provider, code, state string) (*AuthHandlerResponse, error) {
	_, err := s.config.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	githubClient, exists := s.githubClients[provider]
	if !exists {
		return nil, fmt.Errorf("GitHub client not found for provider %s", provider)
	}

	// Generate callback URL
	callbackURL := fmt.Sprintf("%s/api/auth/%s/handler/frame", s.config.RedirectURL, provider)

	oauth2Config := githubClient.GetOAuth2Config(callbackURL)

	// Exchange authorization code for access token
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user profile from GitHub
	profile, err := githubClient.GetUserProfile(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Look up member by email and populate MemberID if found
	profile.MemberID = s.getMemberIDByEmail(profile.Email)

	// Generate JWT token
	jwtToken, err := s.GenerateJWT(profile, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token data
	s.tokenMutex.Lock()
	s.refreshTokens[refreshToken] = &RefreshTokenData{
		UserID:      profile.ID,
		Username:    profile.Username,
		Email:       profile.Email,
		MemberID:    profile.MemberID,
		Provider:    provider,
		AccessToken: token.AccessToken,                   // Store the original OAuth access token
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour), // 30 days
		CreatedAt:   time.Now(),
	}
	s.tokenMutex.Unlock()

	response := &AuthHandlerResponse{
		AccessToken:  jwtToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		RefreshToken: refreshToken,
		Profile:      *profile,
	}

	return response, nil
}

// RefreshToken generates a new JWT token from a refresh token
func (s *AuthService) RefreshToken(refreshToken string) (*AuthHandlerResponse, error) {
	s.tokenMutex.RLock()
	tokenData, exists := s.refreshTokens[refreshToken]
	s.tokenMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if refresh token has expired (valid for 30 days)
	if time.Now().After(tokenData.ExpiresAt) {
		// Clean up expired token
		s.tokenMutex.Lock()
		delete(s.refreshTokens, refreshToken)
		s.tokenMutex.Unlock()
		return nil, fmt.Errorf("refresh token has expired")
	}

	// Create user profile from stored data
	profile := &UserProfile{
		ID:       tokenData.UserID,
		Username: tokenData.Username,
		Email:    tokenData.Email,
		MemberID: tokenData.MemberID,
	}

	// Generate new JWT token
	jwtToken, err := s.GenerateJWT(profile, tokenData.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new JWT: %w", err)
	}

	// Generate new refresh token
	newRefreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	// Store new refresh token and remove old one
	s.tokenMutex.Lock()
	delete(s.refreshTokens, refreshToken)
	s.refreshTokens[newRefreshToken] = &RefreshTokenData{
		UserID:      tokenData.UserID,
		Username:    tokenData.Username,
		Email:       tokenData.Email,
		MemberID:    tokenData.MemberID,
		Provider:    tokenData.Provider,
		AccessToken: tokenData.AccessToken,               // Keep the original OAuth access token
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour), // 30 days
		CreatedAt:   time.Now(),
	}
	s.tokenMutex.Unlock()

	response := &AuthHandlerResponse{
		AccessToken:  jwtToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		RefreshToken: newRefreshToken,
		Profile:      *profile,
	}

	return response, nil
}

// GenerateJWT creates a JWT token for the user
func (s *AuthService) GenerateJWT(userProfile *UserProfile, provider string) (string, error) {
	now := time.Now()
	claims := &AuthClaims{
		UserID:   userProfile.ID,
		Username: userProfile.Username,
		Email:    userProfile.Email,
		Provider: provider,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "developer-portal-backend",
			Subject:   fmt.Sprintf("%d", userProfile.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

// ValidateJWT validates and parses a JWT token
func (s *AuthService) ValidateJWT(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GenerateState generates a random state parameter for OAuth2
func (s *AuthService) GenerateState() (string, error) {
	return s.generateRandomString(32)
}

// generateRefreshToken generates a random refresh token
func (s *AuthService) generateRefreshToken() (string, error) {
	return s.generateRandomString(64)
}

// generateRandomString generates a random base64 encoded string
func (s *AuthService) generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GetGitHubAccessTokenFromClaims retrieves the GitHub OAuth access token for an authenticated user
// using validated JWT claims. This is used to make GitHub API calls on behalf of the user.
func (s *AuthService) GetGitHubAccessTokenFromClaims(claims *AuthClaims) (string, error) {
	if s == nil {
		return "", fmt.Errorf("auth service is not initialized")
	}

	if claims == nil {
		return "", fmt.Errorf("claims cannot be nil")
	}

	s.tokenMutex.RLock()
	defer s.tokenMutex.RUnlock()

	// Find a valid refresh token for this user/provider
	for _, tokenData := range s.refreshTokens {
		if tokenData.UserID == claims.UserID &&
			tokenData.Provider == claims.Provider &&
			time.Now().Before(tokenData.ExpiresAt) {

			return tokenData.AccessToken, nil
		}
	}

	return "", fmt.Errorf("no valid GitHub session found for user %d with provider %s", claims.UserID, claims.Provider)
}

// GetGitHubClient retrieves the GitHub client for a specific provider
func (s *AuthService) GetGitHubClient(provider string) (*GitHubClient, error) {
	if s == nil {
		return nil, fmt.Errorf("auth service is not initialized")
	}

	client, exists := s.githubClients[provider]
	if !exists {
		return nil, fmt.Errorf("GitHub client not found for provider %s", provider)
	}
	return client, nil
}

// Logout handles user logout (stateless JWT tokens don't require server-side logout)
func (s *AuthService) Logout() error {
	// For JWT tokens, logout is typically handled client-side by removing the token
	// In a production system, you might maintain a blacklist of invalidated tokens
	return nil
}
