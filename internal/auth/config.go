package auth

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// AuthConfig holds all authentication configuration for the application
type AuthConfig struct {
	DefaultEnvironment string                    `yaml:"default_environment" json:"default_environment"`
	JWTSecret          string                    `yaml:"jwt_secret" json:"jwt_secret"`
	RedirectURL        string                    `yaml:"redirect_url" json:"redirect_url"`
	Providers          map[string]ProviderConfig `yaml:"providers" json:"providers"`
}

// ProviderConfig holds configuration for a specific provider
type ProviderConfig struct {
	Environments map[string]EnvironmentConfig `yaml:"environments" json:"environments"`
}

// EnvironmentConfig holds environment-specific configuration for a provider
type EnvironmentConfig struct {
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
			fmt.Printf("DEBUG: Auth config file not found, using defaults\n")
		} else {
			return nil, fmt.Errorf("error reading auth config file: %w", err)
		}
	} else {
		fmt.Printf("DEBUG: Successfully read auth config from: %s\n", v.ConfigFileUsed())

		// Debug the raw YAML values
		fmt.Printf("DEBUG: Raw YAML - providers.githubtools.environments.development.enterprise_base_url: '%s'\n",
			v.GetString("providers.githubtools.environments.development.enterprise_base_url"))
		fmt.Printf("DEBUG: Raw YAML - providers.githubtools.environments.development.client_id: '%s'\n",
			v.GetString("providers.githubtools.environments.development.client_id"))
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
		for envName, envConfig := range provider.Environments {
			if envConfig.EnterpriseBaseURL == "" {
				// Try to get the value directly from viper
				viperKey := fmt.Sprintf("providers.%s.environments.%s.enterprise_base_url", providerName, envName)
				if enterpriseURL := v.GetString(viperKey); enterpriseURL != "" {
					envConfig.EnterpriseBaseURL = enterpriseURL
					provider.Environments[envName] = envConfig
					config.Providers[providerName] = provider
				}
			}
		}
	}

	// Override with environment variables for sensitive data
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.JWTSecret = jwtSecret
	}

	// Check for AUTH_REDIRECT_URL environment variable
	authRedirectURL := os.Getenv("AUTH_REDIRECT_URL")
	fmt.Printf("DEBUG: AUTH_REDIRECT_URL env var: '%s'\n", authRedirectURL)
	if authRedirectURL != "" {
		fmt.Printf("DEBUG: Overriding redirect_url with env var\n")
		config.RedirectURL = authRedirectURL
	} else {
		fmt.Printf("DEBUG: No AUTH_REDIRECT_URL env var, keeping config value\n")
		// If config.RedirectURL is still empty after unmarshal, use the default
		if config.RedirectURL == "" {
			fmt.Printf("DEBUG: Config redirect_url is empty, setting from viper\n")
			config.RedirectURL = v.GetString("redirect_url")
		}
	}

	fmt.Printf("DEBUG: Final redirect_url before validation: %s\n", config.RedirectURL)

	// Override provider secrets from environment using your specific variable names
	config = overrideFromEnvironment(config)

	// Debug: Print the final config values
	if githubToolsConfig, err := config.GetProvider("githubtools", "development"); err == nil {
		fmt.Printf("DEBUG: Final githubtools development config - ClientID: '%s', EnterpriseBaseURL: '%s'\n",
			githubToolsConfig.ClientID, githubToolsConfig.EnterpriseBaseURL)
	}

	// Validate configuration
	if err := config.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("auth config validation failed: %w", err)
	}

	return &config, nil
}

// GetProvider returns the configuration for a specific provider and environment
func (c *AuthConfig) GetProvider(provider, env string) (*EnvironmentConfig, error) {
	// Use default environment if not specified
	if env == "" {
		env = c.DefaultEnvironment
	}

	providerConfig, exists := c.Providers[provider]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", provider)
	}

	envConfig, exists := providerConfig.Environments[env]
	if !exists {
		return nil, fmt.Errorf("environment '%s' not found for provider '%s'", env, provider)
	}

	return &envConfig, nil
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
		if len(provider.Environments) == 0 {
			return fmt.Errorf("provider '%s' must have at least one environment", providerName)
		}

		for envName, envConfig := range provider.Environments {
			if envConfig.ClientID == "" {
				return fmt.Errorf("client_id is required for provider '%s' environment '%s'", providerName, envName)
			}
			if envConfig.ClientSecret == "" {
				return fmt.Errorf("client_secret is required for provider '%s' environment '%s'", providerName, envName)
			}
		}
	}

	// Validate default environment exists
	if c.DefaultEnvironment != "" {
		found := false
		for _, provider := range c.Providers {
			if _, exists := provider.Environments[c.DefaultEnvironment]; exists {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("default environment '%s' not found in any provider", c.DefaultEnvironment)
		}
	}

	return nil
}

// setAuthDefaults sets default values for auth configuration
func setAuthDefaults(v *viper.Viper) {
	v.SetDefault("default_environment", "production")
	v.SetDefault("redirect_url", "http://localhost:3000")
	// No default JWT secret - must be provided via environment variable or generated

	// Default providers configuration - don't set enterprise_base_url defaults to let YAML values take precedence
	v.SetDefault("providers", map[string]interface{}{
		"githubtools": map[string]interface{}{
			"environments": map[string]interface{}{
				"development": map[string]interface{}{
					"client_id":     "",
					"client_secret": "",
				},
				"production": map[string]interface{}{
					"client_id":     "",
					"client_secret": "",
				},
			},
		},
		"githubwdf": map[string]interface{}{
			"environments": map[string]interface{}{
				"development": map[string]interface{}{
					"client_id":     "",
					"client_secret": "",
				},
				"production": map[string]interface{}{
					"client_id":     "",
					"client_secret": "",
				},
			},
		},
	})
}

// overrideFromEnvironment overrides config values with your specific environment variables
func overrideFromEnvironment(config AuthConfig) AuthConfig {
	// Helper function to safely update environment config
	updateEnvConfig := func(providerName, envName, clientID, clientSecret string) {
		if provider, exists := config.Providers[providerName]; exists {
			if envConfig, envExists := provider.Environments[envName]; envExists {
				// Create a copy of the environment config to modify
				newEnvConfig := envConfig

				// Override client credentials if provided
				if clientID != "" {
					newEnvConfig.ClientID = clientID
				}
				if clientSecret != "" {
					newEnvConfig.ClientSecret = clientSecret
				}

				// Expand environment variables in existing values if they contain ${...}
				if newEnvConfig.ClientID != "" && len(newEnvConfig.ClientID) > 3 && newEnvConfig.ClientID[:2] == "${" && newEnvConfig.ClientID[len(newEnvConfig.ClientID)-1:] == "}" {
					envVar := newEnvConfig.ClientID[2 : len(newEnvConfig.ClientID)-1]
					if envValue := os.Getenv(envVar); envValue != "" {
						newEnvConfig.ClientID = envValue
					}
				}
				if newEnvConfig.ClientSecret != "" && len(newEnvConfig.ClientSecret) > 3 && newEnvConfig.ClientSecret[:2] == "${" && newEnvConfig.ClientSecret[len(newEnvConfig.ClientSecret)-1:] == "}" {
					envVar := newEnvConfig.ClientSecret[2 : len(newEnvConfig.ClientSecret)-1]
					if envValue := os.Getenv(envVar); envValue != "" {
						newEnvConfig.ClientSecret = envValue
					}
				}

				// EnterpriseBaseURL is preserved from the original config

				provider.Environments[envName] = newEnvConfig
				config.Providers[providerName] = provider
			}
		}
	}

	// GitHub Tools - Development
	updateEnvConfig("githubtools", "development",
		os.Getenv("GITHUB_TOOLS_APP_CLIENT_ID_LOCAL"),
		os.Getenv("GITHUB_TOOLS_APP_CLIENT_SECRET_LOCAL"))

	// GitHub Tools - Production
	updateEnvConfig("githubtools", "production",
		os.Getenv("GITHUB_TOOLS_APP_CLIENT_ID"),
		os.Getenv("GITHUB_TOOLS_APP_CLIENT_SECRET"))

	// GitHub WDF - Development
	updateEnvConfig("githubwdf", "development",
		os.Getenv("GITHUB_WDF_APP_CLIENT_ID_LOCAL"),
		os.Getenv("GITHUB_WDF_APP_CLIENT_SECRET_LOCAL"))

	// GitHub WDF - Production
	updateEnvConfig("githubwdf", "production",
		os.Getenv("GITHUB_WDF_APP_CLIENT_ID"),
		os.Getenv("GITHUB_WDF_APP_CLIENT_SECRET"))

	return config
}
