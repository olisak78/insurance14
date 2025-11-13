package service

// Temporary compatibility types to satisfy generated mocks until mockgen can be run.
// These were previously used by Team link endpoints which have since been removed.

// AddLinkRequest represents a request to add a link to a team.
// Kept minimal for compatibility with existing mocks.
type AddLinkRequest struct {
	// Optional fields retained for compatibility; not used in current code paths.
	URL   string `json:"url,omitempty" validate:"required,url"`
	Title string `json:"title,omitempty" validate:"required,min=1"`
}

// UpdateLinksRequest represents a request to replace all links for a team.
// Kept minimal for compatibility with existing mocks.
type UpdateLinksRequest struct {
	Links []AddLinkRequest `json:"links,omitempty"`
}
