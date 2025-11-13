package auth

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// AuthConfig holds all authentication configuration for the application
type AuthConfig struct {
	JWTSecret   string                    `yaml:"jwt_secret" json:"jwt_secret"`
	RedirectURL string                    `yaml:"redirect_url" json:"redirect_url"`
	Providers   map[string]ProviderConfig `yaml:"providers" json:"providers"`
}

// ProviderConfig holds configuration for a specific provider
type ProviderConfig struct {
	ClientID          string `yaml:"client_id" json:"client_id"`
	ClientSecret      string `yaml:"client_secret" json:"client_secret"`
	EnterpriseBaseURL string `yaml:"enterprise_base_url,omitempty" json:"enterprise_base_url,omitempty"`
}

// LoadAuthConfig loads and validates authentication configuration
func LoadAuthConfig(configPath string) (*AuthConfig, error) {
	// Create a new viper instance for auth config
	v := viper.New()

	// Set config file details
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("auth")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	// Set default values
	setAuthDefaults(v)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults and environment variables
		} else {
			return nil, fmt.Errorf("error reading auth config file: %w", err)
		}
	}

	// Override with environment variables
	v.AutomaticEnv()

	var config AuthConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling auth config: %w", err)
	}

	// Manual fix for enterprise_base_url mapping issue
	// Viper seems to have trouble mapping snake_case to PascalCase, so let's do it manually
	for providerName, provider := range config.Providers {
		if provider.EnterpriseBaseURL == "" {
			// Try to get the value directly from viper
			viperKey := fmt.Sprintf("providers.%s.enterprise_base_url", providerName)
			if enterpriseURL := v.GetString(viperKey); enterpriseURL != "" {
				provider.EnterpriseBaseURL = enterpriseURL
				config.Providers[providerName] = provider
			}
		}
	}

	// Override with environment variables for sensitive data
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.JWTSecret = jwtSecret
	}

	// Check for AUTH_REDIRECT_URL environment variable
	authRedirectURL := os.Getenv("AUTH_REDIRECT_URL")
	if authRedirectURL != "" {
		config.RedirectURL = authRedirectURL
	} else {
		// If config.RedirectURL is still empty after unmarshal, use the default
		if config.RedirectURL == "" {
			config.RedirectURL = v.GetString("redirect_url")
		}
	}


	// Override provider secrets from environment using your specific variable names
	config = overrideFromEnvironment(config)

	// Validate configuration
	if err := config.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("auth config validation failed: %w", err)
	}

	return &config, nil
}

// GetProvider returns the configuration for a specific provider
func (c *AuthConfig) GetProvider(provider string) (*ProviderConfig, error) {
	providerConfig, exists := c.Providers[provider]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", provider)
	}

	return &providerConfig, nil
}

// ValidateConfig validates the authentication configuration
func (c *AuthConfig) ValidateConfig() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if c.RedirectURL == "" {
		return fmt.Errorf("redirect URL is required")
	}

	if len(c.Providers) == 0 {
		return fmt.Errorf("at least one provider must be configured")
	}

	// Validate each provider
	for providerName, provider := range c.Providers {
		if provider.ClientID == "" {
			return fmt.Errorf("client_id is required for provider '%s'", providerName)
		}
		if provider.ClientSecret == "" {
			return fmt.Errorf("client_secret is required for provider '%s'", providerName)
		}
	}

	return nil
}

// setAuthDefaults sets default values for auth configuration
func setAuthDefaults(v *viper.Viper) {
	v.SetDefault("redirect_url", "http://localhost:3000")
	// No default JWT secret - must be provided via environment variable or generated

	// Default providers configuration - don't set enterprise_base_url defaults to let YAML values take precedence
	v.SetDefault("providers", map[string]interface{}{
		"githubtools": map[string]interface{}{
			"client_id":     "",
			"client_secret": "",
		},
		"githubwdf": map[string]interface{}{
			"client_id":     "",
			"client_secret": "",
		},
	})
}

// overrideFromEnvironment overrides config values with your specific environment variables
func overrideFromEnvironment(config AuthConfig) AuthConfig {
	// Helper function to safely update provider config
	updateProviderConfig := func(providerName, clientID, clientSecret string) {
		if provider, exists := config.Providers[providerName]; exists {
			// Create a copy of the provider config to modify
			newProvider := provider

			// Override client credentials if provided
			if clientID != "" {
				newProvider.ClientID = clientID
			}
			if clientSecret != "" {
				newProvider.ClientSecret = clientSecret
			}

			// Expand environment variables in existing values if they contain ${...}
			if newProvider.ClientID != "" && len(newProvider.ClientID) > 3 && newProvider.ClientID[:2] == "${" && newProvider.ClientID[len(newProvider.ClientID)-1:] == "}" {
				envVar := newProvider.ClientID[2 : len(newProvider.ClientID)-1]
				if envValue := os.Getenv(envVar); envValue != "" {
					newProvider.ClientID = envValue
				}
			}
			if newProvider.ClientSecret != "" && len(newProvider.ClientSecret) > 3 && newProvider.ClientSecret[:2] == "${" && newProvider.ClientSecret[len(newProvider.ClientSecret)-1:] == "}" {
				envVar := newProvider.ClientSecret[2 : len(newProvider.ClientSecret)-1]
				if envValue := os.Getenv(envVar); envValue != "" {
					newProvider.ClientSecret = envValue
				}
			}

			// EnterpriseBaseURL is preserved from the original config

			config.Providers[providerName] = newProvider
		}
	}

	// GitHub Tools
	updateProviderConfig("githubtools",
		os.Getenv("GITHUB_TOOLS_APP_CLIENT_ID"),
		os.Getenv("GITHUB_TOOLS_APP_CLIENT_SECRET"))

	// GitHub WDF
	updateProviderConfig("githubwdf",
		os.Getenv("GITHUB_WDF_APP_CLIENT_ID"),
		os.Getenv("GITHUB_WDF_APP_CLIENT_SECRET"))

	return config
}
