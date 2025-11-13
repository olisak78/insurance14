package service

import (
	"encoding/json"
	"errors"
	"fmt"

	apperrors "developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ComponentService handles business logic for components
type ComponentService struct {
	repo             *repository.ComponentRepository
	organizationRepo *repository.OrganizationRepository
	projectRepo      *repository.ProjectRepository
	validator        *validator.Validate
}

// NewComponentService creates a new component service
func NewComponentService(repo *repository.ComponentRepository, orgRepo *repository.OrganizationRepository, projRepo *repository.ProjectRepository, validator *validator.Validate) *ComponentService {
	return &ComponentService{
		repo:             repo,
		organizationRepo: orgRepo,
		projectRepo:      projRepo,
		validator:        validator,
	}
}

// ComponentProjectView is a minimal view for /components?project-name=<name>
type ComponentProjectView struct {
	ID          uuid.UUID       `json:"id"`
	OwnerID     uuid.UUID       `json:"owner_id"`
	Name        string          `json:"name"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	QOS         string          `json:"qos,omitempty"`
	Sonar       string          `json:"sonar,omitempty"`
	GitHub      string          `json:"github,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
}

// GetByProjectNameAllView returns ALL components for a project (unpaginated) with a minimal view:
// - Omits project_id, created_at, updated_at, metadata
// - Adds fields: qos (metadata.ci.qos), sonar (metadata.sonar.project_id), github (metadata.github.url)
func (s *ComponentService) GetByProjectNameAllView(projectName string) ([]ComponentProjectView, error) {
	if projectName == "" {
		return []ComponentProjectView{}, nil
	}
	project, err := s.projectRepo.GetByName(projectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to resolve project by name: %w", err)
	}
	if project == nil {
		return nil, apperrors.ErrProjectNotFound
	}

	components, _, err := s.repo.GetComponentsByProjectID(project.ID, 1000000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get components by project: %w", err)
	}

	views := make([]ComponentProjectView, len(components))
	for i, c := range components {
		views[i] = ComponentProjectView{
			ID:          c.ID,
			OwnerID:     c.OwnerID,
			Name:        c.Name,
			Title:       c.Title,
			Description: c.Description,
			Metadata:    c.Metadata, // Include full metadata for filtering
		}

		// Extract qos, sonar, github from metadata if present
		if len(c.Metadata) > 0 {
			var meta map[string]interface{}
			if err := json.Unmarshal(c.Metadata, &meta); err == nil {
				// qos from metadata.ci.qos
				if ciRaw, ok := meta["ci"]; ok {
					if ciMap, ok := ciRaw.(map[string]interface{}); ok {
						if qosRaw, ok := ciMap["qos"]; ok {
							if qosStr, ok := qosRaw.(string); ok {
								views[i].QOS = qosStr
							}
						}
					}
				}
				// sonar from metadata.sonar.project_id
				if sonarRaw, ok := meta["sonar"]; ok {
					if sonarMap, ok := sonarRaw.(map[string]interface{}); ok {
						if pidRaw, ok := sonarMap["project_id"]; ok {
							if pidStr, ok := pidRaw.(string); ok {
								views[i].Sonar = "https://sonar.tools.sap/dashboard?id=" + pidStr
							}
						}
					}
				}
				// github from metadata.github.url
				if ghRaw, ok := meta["github"]; ok {
					if ghMap, ok := ghRaw.(map[string]interface{}); ok {
						if urlRaw, ok := ghMap["url"]; ok {
							if urlStr, ok := urlRaw.(string); ok {
								views[i].GitHub = urlStr
							}
						}
					}
				}
			}
		}
	}

	return views, nil
}

// GetProjectTitleByID returns the project's title by ID
func (s *ComponentService) GetProjectTitleByID(id uuid.UUID) (string, error) {
	project, err := s.projectRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", apperrors.ErrProjectNotFound
		}
		return "", fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return "", apperrors.ErrProjectNotFound
	}
	return project.Title, nil
}
