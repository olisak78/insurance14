package auth

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthConfig(t *testing.T) {
	t.Run("valid config structure", func(t *testing.T) {
		// Test creating a valid config directly
		config := &AuthConfig{
			DefaultEnvironment: "development",
			JWTSecret:          "test-signing-key",
			RedirectURL:        "http://localhost:3000",
			Providers: map[string]ProviderConfig{
				"githubtools": {
					Environments: map[string]EnvironmentConfig{
						"development": {
							ClientID:          "dev-client-id",
							ClientSecret:      "dev-client-secret",
							EnterpriseBaseURL: "https://github.tools.sap",
						},
						"production": {
							ClientID:          "prod-client-id",
							ClientSecret:      "prod-client-secret",
							EnterpriseBaseURL: "https://github.tools.sap",
						},
					},
				},
				"githubwdf": {
					Environments: map[string]EnvironmentConfig{
						"development": {
							ClientID:          "wdf-dev-client-id",
							ClientSecret:      "wdf-dev-client-secret",
							EnterpriseBaseURL: "https://github.wdf.sap.corp",
						},
						"production": {
							ClientID:          "wdf-prod-client-id",
							ClientSecret:      "wdf-prod-client-secret",
							EnterpriseBaseURL: "https://github.wdf.sap.corp",
						},
					},
				},
			},
		}

		// Test validation
		err := config.ValidateConfig()
		assert.NoError(t, err)
		assert.NotEmpty(t, config.JWTSecret)
		assert.NotEmpty(t, config.RedirectURL)
	})

	t.Run("missing jwt secret", func(t *testing.T) {
		config := &AuthConfig{
			RedirectURL: "http://localhost:3000",
			Providers: map[string]ProviderConfig{
				"githubtools": {
					Environments: map[string]EnvironmentConfig{
						"development": {
							ClientID:     "dev-client-id",
							ClientSecret: "dev-client-secret",
						},
					},
				},
			},
		}

		err := config.ValidateConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT secret is required")
	})

	t.Run("missing redirect url", func(t *testing.T) {
		config := &AuthConfig{
			JWTSecret: "test-secret",
			Providers: map[string]ProviderConfig{
				"githubtools": {
					Environments: map[string]EnvironmentConfig{
						"development": {
							ClientID:     "dev-client-id",
							ClientSecret: "dev-client-secret",
						},
					},
				},
			},
		}

		err := config.ValidateConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redirect URL is required")
	})

	t.Run("missing client credentials", func(t *testing.T) {
		config := &AuthConfig{
			JWTSecret:   "test-secret",
			RedirectURL: "http://localhost:3000",
			Providers: map[string]ProviderConfig{
				"githubtools": {
					Environments: map[string]EnvironmentConfig{
						"development": {
							// Missing ClientID and ClientSecret
						},
					},
				},
			},
		}

		err := config.ValidateConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client_id is required")
	})
}

func TestGitHubClientConfig(t *testing.T) {
	config := &EnvironmentConfig{
		ClientID:          "test-client-id",
		ClientSecret:      "test-client-secret",
		EnterpriseBaseURL: "https://github.example.com",
	}

	client := NewGitHubClient(config)
	assert.NotNil(t, client)

	oauthConfig := client.GetOAuth2Config("http://localhost:8080/callback")
	assert.Equal(t, "test-client-id", oauthConfig.ClientID)
	assert.Equal(t, "test-client-secret", oauthConfig.ClientSecret)
	assert.Equal(t, "http://localhost:8080/callback", oauthConfig.RedirectURL)
	assert.Contains(t, oauthConfig.Scopes, "user:email")
}

func TestJWTOperations(t *testing.T) {
	config := &AuthConfig{
		JWTSecret:   "test-signing-key-for-jwt-operations",
		RedirectURL: "http://localhost:3000",
		Providers: map[string]ProviderConfig{
			"githubtools": {
				Environments: map[string]EnvironmentConfig{
					"development": {
						ClientID:          "test-client-id",
						ClientSecret:      "test-client-secret",
						EnterpriseBaseURL: "https://github.tools.sap",
					},
				},
			},
		},
		DefaultEnvironment: "development",
	}

	service, err := NewAuthService(config, nil)
	require.NoError(t, err)

	userProfile := &UserProfile{
		ID:        12345,
		Username:  "testuser",
		Email:     "test@example.com",
		Name:      "Test User",
		AvatarURL: "https://avatars.githubusercontent.com/u/12345",
	}

	// Test token generation
	token, err := service.GenerateJWT(userProfile, "githubtools", "development")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test token validation
	validatedClaims, err := service.ValidateJWT(token)
	assert.NoError(t, err)
	assert.Equal(t, userProfile.ID, validatedClaims.UserID)
	assert.Equal(t, userProfile.Username, validatedClaims.Username)
	assert.Equal(t, userProfile.Email, validatedClaims.Email)
	assert.Equal(t, "githubtools", validatedClaims.Provider)
	assert.Equal(t, "development", validatedClaims.Environment)

	// Test invalid token
	_, err = service.ValidateJWT("invalid-token")
	assert.Error(t, err)
}

func TestAuthHandlers(t *testing.T) {
	// Create test config
	config := &AuthConfig{
		Providers: map[string]ProviderConfig{
			"githubtools": {
				Environments: map[string]EnvironmentConfig{
					"development": {
						ClientID:          "test-client-id",
						ClientSecret:      "test-client-secret",
						EnterpriseBaseURL: "https://github.tools.sap",
					},
				},
			},
		},
		JWTSecret:          "test-signing-key",
		RedirectURL:        "http://localhost:3000",
		DefaultEnvironment: "development",
	}

	service, err := NewAuthService(config, nil)
	require.NoError(t, err)

	handler := NewAuthHandler(service)

	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)

	t.Run("Start endpoint", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/auth/githubtools/start?env=development", nil)
		c.Params = gin.Params{{Key: "provider", Value: "githubtools"}}

		handler.Start(c)

		assert.Equal(t, http.StatusFound, w.Code)
		location := w.Header().Get("Location")
		assert.Contains(t, location, "github.tools.sap")
		assert.Contains(t, location, "oauth/authorize")
	})

	t.Run("Logout endpoint", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"env": "development",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/auth/githubtools/logout", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "provider", Value: "githubtools"}}

		handler.Logout(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Logged out successfully", response["message"])
	})
}

func TestProviderValidation(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		environment string
		expectValid bool
	}{
		{"valid githubtools dev", "githubtools", "development", true},
		{"valid githubtools prod", "githubtools", "production", true},
		{"valid githubwdf dev", "githubwdf", "development", true},
		{"invalid provider", "invalid", "development", false},
		{"invalid environment", "githubtools", "invalid", false},
	}

	config := &AuthConfig{
		Providers: map[string]ProviderConfig{
			"githubtools": {
				Environments: map[string]EnvironmentConfig{
					"development": {ClientID: "dev-id", ClientSecret: "dev-secret"},
					"production":  {ClientID: "prod-id", ClientSecret: "prod-secret"},
				},
			},
			"githubwdf": {
				Environments: map[string]EnvironmentConfig{
					"development": {ClientID: "wdf-dev-id", ClientSecret: "wdf-dev-secret"},
					"production":  {ClientID: "wdf-prod-id", ClientSecret: "wdf-prod-secret"},
				},
			},
		},
		JWTSecret:          "test-key",
		RedirectURL:        "http://localhost:3000",
		DefaultEnvironment: "production",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerConfig, err := config.GetProvider(tt.provider, tt.environment)
			if tt.expectValid {
				assert.NoError(t, err)
				assert.NotNil(t, providerConfig)
				assert.NotEmpty(t, providerConfig.ClientID)
			} else {
				assert.Error(t, err)
				assert.Nil(t, providerConfig)
			}
		})
	}
}

func TestStateGeneration(t *testing.T) {
	config := &AuthConfig{
		JWTSecret:   "test-key",
		RedirectURL: "http://localhost:3000",
		Providers: map[string]ProviderConfig{
			"githubtools": {
				Environments: map[string]EnvironmentConfig{
					"development": {ClientID: "test-id", ClientSecret: "test-secret"},
				},
			},
		},
		DefaultEnvironment: "development",
	}

	service, err := NewAuthService(config, nil)
	require.NoError(t, err)

	// Test multiple state generations to ensure uniqueness
	states := make(map[string]bool)
	for i := 0; i < 100; i++ {
		state, err := service.GenerateState()
		assert.NoError(t, err)
		assert.NotEmpty(t, state)
		assert.False(t, states[state], "State should be unique")
		states[state] = true
	}
}

func TestEnvironmentDefault(t *testing.T) {
	config := &AuthConfig{
		Providers: map[string]ProviderConfig{
			"githubtools": {
				Environments: map[string]EnvironmentConfig{
					"production": {ClientID: "prod-id", ClientSecret: "prod-secret"},
				},
			},
		},
		JWTSecret:          "test-key",
		RedirectURL:        "http://localhost:3000",
		DefaultEnvironment: "production",
	}

	// Test that empty environment defaults to default environment
	providerConfig, err := config.GetProvider("githubtools", "")
	assert.NoError(t, err)
	assert.NotNil(t, providerConfig)
	assert.Equal(t, "prod-id", providerConfig.ClientID)
}

func TestAuthService_GetAuthURL(t *testing.T) {
	config := &AuthConfig{
		Providers: map[string]ProviderConfig{
			"githubtools": {
				Environments: map[string]EnvironmentConfig{
					"development": {
						ClientID:          "test-client-id",
						ClientSecret:      "test-client-secret",
						EnterpriseBaseURL: "https://github.tools.sap",
					},
				},
			},
		},
		JWTSecret:          "test-key",
		RedirectURL:        "http://localhost:3000",
		DefaultEnvironment: "development",
	}

	service, err := NewAuthService(config, nil)
	require.NoError(t, err)

	authURL, err := service.GetAuthURL("githubtools", "development", "test-state")
	assert.NoError(t, err)
	assert.Contains(t, authURL, "github.tools.sap")
	assert.Contains(t, authURL, "oauth/authorize")
	assert.Contains(t, authURL, "test-state")
}

func TestRefreshToken(t *testing.T) {
	config := &AuthConfig{
		Providers: map[string]ProviderConfig{
			"githubtools": {
				Environments: map[string]EnvironmentConfig{
					"development": {
						ClientID:          "test-client-id",
						ClientSecret:      "test-client-secret",
						EnterpriseBaseURL: "https://github.tools.sap",
					},
				},
			},
		},
		JWTSecret:          "test-key",
		RedirectURL:        "http://localhost:3000",
		DefaultEnvironment: "development",
	}

	service, err := NewAuthService(config, nil)
	require.NoError(t, err)

	// Test invalid refresh token
	_, err = service.RefreshToken("invalid-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid refresh token")
}

// Helper functions for testing

func writeTemporaryFile(t *testing.T, content string) string {
	tmpFile, err := ioutil.TempFile("", "auth-config-*.yaml")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	return tmpFile.Name()
}

func removeTemporaryFile(t *testing.T, filename string) {
	err := os.Remove(filename)
	require.NoError(t, err)
}

// Mock member repository for testing
type mockMemberRepo struct {
	members map[string]mockMember
}

type mockMember struct {
	ID    string
	Email string
}

func (m *mockMemberRepo) GetByEmail(email string) (interface{}, error) {
	if member, ok := m.members[email]; ok {
		return &member, nil
	}
	return nil, nil
}

// TestMemberIDLookup tests that member ID is properly looked up and populated
func TestMemberIDLookup(t *testing.T) {
	config := &AuthConfig{
		JWTSecret:   "test-key",
		RedirectURL: "http://localhost:3000",
		Providers: map[string]ProviderConfig{
			"githubtools": {
				Environments: map[string]EnvironmentConfig{
					"development": {
						ClientID:     "test-client-id",
						ClientSecret: "test-client-secret",
					},
				},
			},
		},
		DefaultEnvironment: "development",
	}

	// Create mock member repository
	mockRepo := &mockMemberRepo{
		members: map[string]mockMember{
			"test@example.com": {
				ID:    "member-uuid-123",
				Email: "test@example.com",
			},
		},
	}

	service, err := NewAuthService(config, mockRepo)
	require.NoError(t, err)

	// Test getMemberIDByEmail with existing member
	memberID := service.getMemberIDByEmail("test@example.com")
	assert.NotNil(t, memberID)
	assert.Equal(t, "member-uuid-123", *memberID)

	// Test getMemberIDByEmail with non-existing member
	memberID = service.getMemberIDByEmail("nonexistent@example.com")
	assert.Nil(t, memberID)

	// Test getMemberIDByEmail with empty email
	memberID = service.getMemberIDByEmail("")
	assert.Nil(t, memberID)
}

// TestUserProfileWithMemberID tests that UserProfile correctly includes MemberID
func TestUserProfileWithMemberID(t *testing.T) {
	t.Run("profile with member ID", func(t *testing.T) {
		memberID := "member-123"
		profile := &UserProfile{
			ID:        12345,
			Username:  "testuser",
			Email:     "test@example.com",
			Name:      "Test User",
			AvatarURL: "https://example.com/avatar.png",
			MemberID:  &memberID,
		}

		assert.NotNil(t, profile.MemberID)
		assert.Equal(t, "member-123", *profile.MemberID)
	})

	t.Run("profile without member ID", func(t *testing.T) {
		profile := &UserProfile{
			ID:        12345,
			Username:  "testuser",
			Email:     "test@example.com",
			Name:      "Test User",
			AvatarURL: "https://example.com/avatar.png",
			MemberID:  nil,
		}

		assert.Nil(t, profile.MemberID)
	})
}

// TestRefreshTokenDataWithMemberID tests that refresh token data includes MemberID
func TestRefreshTokenDataWithMemberID(t *testing.T) {
	memberID := "member-456"
	tokenData := &RefreshTokenData{
		UserID:      12345,
		Username:    "testuser",
		Email:       "test@example.com",
		MemberID:    &memberID,
		Provider:    "githubtools",
		Environment: "development",
		AccessToken: "access-token",
	}

	assert.NotNil(t, tokenData.MemberID)
	assert.Equal(t, "member-456", *tokenData.MemberID)
}

// TestRefreshTokenWithMemberID tests that refresh token flow preserves MemberID
func TestRefreshTokenWithMemberID(t *testing.T) {
	config := &AuthConfig{
		JWTSecret:   "test-key",
		RedirectURL: "http://localhost:3000",
		Providers: map[string]ProviderConfig{
			"githubtools": {
				Environments: map[string]EnvironmentConfig{
					"development": {
						ClientID:     "test-client-id",
						ClientSecret: "test-client-secret",
					},
				},
			},
		},
		DefaultEnvironment: "development",
	}

	// Create mock member repository
	mockRepo := &mockMemberRepo{
		members: map[string]mockMember{
			"test@example.com": {
				ID:    "member-uuid-789",
				Email: "test@example.com",
			},
		},
	}

	service, err := NewAuthService(config, mockRepo)
	require.NoError(t, err)

	// Manually create a refresh token with member ID
	memberID := "member-uuid-789"
	refreshToken := "test-refresh-token"
	service.tokenMutex.Lock()
	service.refreshTokens[refreshToken] = &RefreshTokenData{
		UserID:      12345,
		Username:    "testuser",
		Email:       "test@example.com",
		MemberID:    &memberID,
		Provider:    "githubtools",
		Environment: "development",
		AccessToken: "old-access-token",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
	}
	service.tokenMutex.Unlock()

	// Test refresh token
	refreshed, err := service.RefreshToken(refreshToken)
	assert.NoError(t, err)
	assert.NotNil(t, refreshed)
	assert.NotNil(t, refreshed.Profile.MemberID)
	assert.Equal(t, "member-uuid-789", *refreshed.Profile.MemberID)
}

// TestGetMemberIDByEmailPublicMethod tests the public wrapper method
func TestGetMemberIDByEmailPublicMethod(t *testing.T) {
	config := &AuthConfig{
		JWTSecret:   "test-key",
		RedirectURL: "http://localhost:3000",
		Providers: map[string]ProviderConfig{
			"githubtools": {
				Environments: map[string]EnvironmentConfig{
					"development": {
						ClientID:     "test-client-id",
						ClientSecret: "test-client-secret",
					},
				},
			},
		},
		DefaultEnvironment: "development",
	}

	// Create mock member repository
	mockRepo := &mockMemberRepo{
		members: map[string]mockMember{
			"public@example.com": {
				ID:    "public-member-id",
				Email: "public@example.com",
			},
		},
	}

	service, err := NewAuthService(config, mockRepo)
	require.NoError(t, err)

	// Test public method
	memberID := service.GetMemberIDByEmail("public@example.com")
	assert.NotNil(t, memberID)
	assert.Equal(t, "public-member-id", *memberID)
}

// TestAuthServiceWithNilMemberRepo tests that auth service works without member repo
func TestAuthServiceWithNilMemberRepo(t *testing.T) {
	config := &AuthConfig{
		JWTSecret:   "test-key",
		RedirectURL: "http://localhost:3000",
		Providers: map[string]ProviderConfig{
			"githubtools": {
				Environments: map[string]EnvironmentConfig{
					"development": {
						ClientID:     "test-client-id",
						ClientSecret: "test-client-secret",
					},
				},
			},
		},
		DefaultEnvironment: "development",
	}

	service, err := NewAuthService(config, nil)
	require.NoError(t, err)

	// Test that getMemberIDByEmail returns nil when repo is nil
	memberID := service.getMemberIDByEmail("any@example.com")
	assert.Nil(t, memberID)

	// Test JWT generation still works without member repo
	profile := &UserProfile{
		ID:       12345,
		Username: "testuser",
		Email:    "test@example.com",
	}

	token, err := service.GenerateJWT(profile, "githubtools", "development")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}
