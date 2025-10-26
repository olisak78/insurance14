package service

import (
	"fmt"

	"developer-portal-backend/internal/auth"
)

//go:generate mockgen -source=github_auth_interface.go -destination=../mocks/github_auth_mocks.go -package=mocks

// GitHubAuthService defines the interface for auth service methods needed by GitHub service
type GitHubAuthService interface {
	GetGitHubAccessTokenFromClaims(claims *auth.AuthClaims) (string, error)
	GetGitHubClient(provider, environment string) (*auth.GitHubClient, error)
}

// authServiceAdapter adapts auth.AuthService to implement GitHubAuthService interface
type authServiceAdapter struct {
	authService *auth.AuthService
}

// NewAuthServiceAdapter creates an adapter for auth.AuthService
func NewAuthServiceAdapter(authService *auth.AuthService) GitHubAuthService {
	if authService == nil {
		return &authServiceAdapter{authService: nil}
	}
	return &authServiceAdapter{authService: authService}
}

func (a *authServiceAdapter) GetGitHubAccessTokenFromClaims(claims *auth.AuthClaims) (string, error) {
	if a.authService == nil {
		return "", fmt.Errorf("auth service is not initialized")
	}
	return a.authService.GetGitHubAccessTokenFromClaims(claims)
}

func (a *authServiceAdapter) GetGitHubClient(provider, environment string) (*auth.GitHubClient, error) {
	if a.authService == nil {
		return nil, fmt.Errorf("auth service is not initialized")
	}
	return a.authService.GetGitHubClient(provider, environment)
}
