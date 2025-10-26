# Developer Portal Backend

A robust Go backend server for a developer portal, built with modern best practices and clean architecture.

## 🚀 TLDR - Quick Start

```bash
# 1. Clone and setup
git clone <repository-url>
cd developer-portal-backend
make setup

# 2. Setup environment configuration
cp .env.example .env
# Edit .env to configure port and other settings

# 3. Start development server (auto-generates JWT secret)
make run-dev

# 4. Server runs on http://localhost:7008 (default) or PORT from .env
```

**That's it!** The server will be running with a test database and sample data loaded.

---

## 🏗️ Architecture

This project follows **Clean Architecture** principles with clear separation of concerns:

```
developer-portal-backend/
├── cmd/
│   └── server/                 # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/          # HTTP handlers (controllers)
│   │   ├── middleware/        # Custom middleware
│   │   └── routes/            # Route definitions
│   ├── auth/                  # Authentication (GitHub OAuth)
│   ├── config/                # Configuration management
│   ├── database/
│   │   ├── migrations/        # SQL migration files
│   │   └── models/            # Database models/entities
│   ├── mocks/                 # Test mocks
│   ├── repository/            # Data access layer
│   ├── service/               # Business logic layer
│   ├── testutils/            # Testing utilities
│   └── utils/                 # Utility functions
├── config/                    # Configuration files
├── docker/                    # Docker configuration
├── pkg/                       # Public packages
├── scripts/                   # Development scripts
├── .env.example              # Environment variables template
├── Makefile                  # Development tasks
└── README.md
```

## 🚀 Setup & Running

### Option 1: Quick Development Setup (Recommended)

```bash
# Complete setup with one command
make setup

# Start development server (auto-generates JWT secret and loads sample data)
make run-dev
```

### Option 2: Manual Setup

```bash
# 1. Install dependencies
make deps

# 2. Start database
make db-up

# 3. Setup environment
cp .env.example .env
# Edit .env as needed

# 4. Run migrations (if any)
make migrate-up

# 5. Load initial data
make load-initial-data

# 6. Start server
make run-dev
```

### Option 3: Production Mode

```bash
# Set JWT secret
export JWT_SECRET=$(openssl rand -base64 32)

# Build and run
make run
```

The server will start on `http://localhost:7008` (configurable via PORT in .env)

## 🐳 Docker Development

### Start All Services

```bash
# Start PostgreSQL, pgAdmin, and Redis
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

### Database Management

```bash
# Start only PostgreSQL
make db-up

# Reset database (WARNING: deletes all data)
make db-reset

# Load initial sample data
make load-initial-data
```

## 🗄️ Database Access

- **PostgreSQL**: `localhost:5432`
  - Database: `developer_portal`
  - Username: `postgres`
  - Password: `postgres`

- **pgAdmin**: `http://localhost:5050`
  - Email: `admin@developer-portal.com`
  - Password: `admin`

## 🔐 Authentication

The application supports **GitHub OAuth authentication** with multi-provider configuration:

### Supported Providers
- **githubtools**: GitHub Enterprise (https://github.tools.sap)
- **githubwdf**: GitHub Enterprise (https://github.wdf.sap.corp)

### Auth Configuration
Edit `config/auth.yaml` and set environment variables:
```bash
# For GitHub Tools
GITHUB_TOOLS_APP_CLIENT_ID=your_client_id
GITHUB_TOOLS_APP_CLIENT_SECRET=your_client_secret

# For GitHub WDF
GITHUB_WDF_APP_CLIENT_ID=your_client_id
GITHUB_WDF_APP_CLIENT_SECRET=your_client_secret
```

### Auth Flow (Backstage Compatible)
1. `GET /api/auth/{provider}/start` - Initiate OAuth flow
2. `GET /api/auth/{provider}/handler/frame` - Handle OAuth callback
3. `GET /api/auth/{provider}/refresh` - Refresh tokens
4. `POST /api/auth/{provider}/logout` - Logout

## 📡 API Endpoints

### Health Checks
- `GET /health` - Application health status
- `GET /health/ready` - Readiness check
- `GET /health/live` - Liveness check

### Organizations API (v1) - All endpoints require authentication
- `GET /api/v1/organizations` - List organizations
- `POST /api/v1/organizations` - Create organization
- `GET /api/v1/organizations/:id` - Get organization
- `PUT /api/v1/organizations/:id` - Update organization
- `DELETE /api/v1/organizations/:id` - Delete organization
- `GET /api/v1/organizations/:id/members` - Get org members
- `GET /api/v1/organizations/:id/teams` - Get org teams
- `GET /api/v1/organizations/:id/projects` - Get org projects
- `GET /api/v1/organizations/:id/components` - Get org components
- `GET /api/v1/organizations/:id/landscapes` - Get org landscapes

### Teams API (v1)
- `GET /api/v1/teams` - Get all teams (optional organization_id param)
- `POST /api/v1/teams` - Create team
- `GET /api/v1/teams/:id` - Get team
- `PUT /api/v1/teams/:id` - Update team
- `DELETE /api/v1/teams/:id` - Delete team
- `GET /api/v1/teams/:id/members` - Get team with members
- `GET /api/v1/teams/:id/details` - Get team details
- `POST /api/v1/teams/:id/links` - Add a link to a team
- `DELETE /api/v1/teams/:id/links` - Remove a link from a team (requires url query param)
- `GET /api/v1/teams/by-name/:name` - Get team by name (requires organization_id param)
- `GET /api/v1/teams/by-name/:name/members` - Get team members by name
- `GET /api/v1/teams/by-name/:name/components` - Get team components by name

### Members API (v1)
- `GET /api/v1/members` - Get members (requires organization_id param)
- `POST /api/v1/members` - Create member
- `GET /api/v1/members/:id` - Get member
- `PUT /api/v1/members/:id` - Update member
- `POST /api/v1/members/:id/quick-links` - Add a quick link to a member
- `DELETE /api/v1/members/:id/quick-links` - Remove a quick link from a member (requires url query param)
- `DELETE /api/v1/members/:id` - Delete member

### Projects API (v1)
- `GET /api/v1/projects` - Get projects (requires organization_id param)
- `POST /api/v1/projects` - Create project
- `GET /api/v1/projects/:id` - Get project
- `PUT /api/v1/projects/:id` - Update project
- `DELETE /api/v1/projects/:id` - Delete project
- `GET /api/v1/projects/:id/organization` - Get project with organization
- `GET /api/v1/projects/:id/components` - Get project components
- `GET /api/v1/projects/:id/landscapes` - Get project landscapes
- `GET /api/v1/projects/status/:status` - Get projects by status

### Components API (v1)
- `GET /api/v1/components` - List components
- `POST /api/v1/components` - Create component
- `GET /api/v1/components/:id` - Get component
- `GET /api/v1/components/by-name/:name` - Get component by name
- `PUT /api/v1/components/:id` - Update component
- `DELETE /api/v1/components/:id` - Delete component
- `GET /api/v1/components/:id/projects` - Get component projects
- `GET /api/v1/components/:id/deployments` - Get component deployments
- `GET /api/v1/components/:id/ownerships` - Get component ownerships
- `GET /api/v1/components/:id/details` - Get component full details

### Landscapes API (v1)
- `GET /api/v1/landscapes` - List landscapes
- `POST /api/v1/landscapes` - Create landscape
- `GET /api/v1/landscapes/:id` - Get landscape
- `PUT /api/v1/landscapes/:id` - Update landscape
- `DELETE /api/v1/landscapes/:id` - Delete landscape
- `GET /api/v1/landscapes/:id/projects` - Get landscape projects
- `GET /api/v1/landscapes/:id/deployments` - Get landscape deployments
- `GET /api/v1/landscapes/:id/details` - Get landscape details
- `GET /api/v1/landscapes/environment/:environment` - Get landscapes by environment

### Component Deployments API (v1)
- `GET /api/v1/component-deployments` - List component deployments
- `POST /api/v1/component-deployments` - Create component deployment
- `GET /api/v1/component-deployments/:id` - Get component deployment
- `PUT /api/v1/component-deployments/:id` - Update component deployment
- `DELETE /api/v1/component-deployments/:id` - Delete component deployment
- `GET /api/v1/component-deployments/:id/details` - Get deployment details

### Example API Usage

```bash
# Health check
curl http://localhost:7008/health

# List organizations (requires authentication)
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:7008/api/v1/organizations

# Get teams for organization
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  "http://localhost:7008/api/v1/teams?organization_id=1"

# Create a new team
curl -X POST http://localhost:7008/api/v1/teams \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Platform Team",
    "description": "Core platform development team",
    "organization_id": 1
  }'
```

## 🧪 Testing

### Quick Testing
```bash
# Run unit tests only (fast)
make test-unit

# Pre-commit checks (format, lint, unit tests)
make test-precommit
```

### Comprehensive Testing
```bash
# Setup test environment
make test-setup

# Run all tests (unit + integration)
make test-all

# Run tests with coverage
make test-coverage-full

# Run tests with race detection
make test-race

# Cleanup test environment
make test-teardown
```

### Docker Testing
```bash
# Run tests in Docker container
make test-docker

# Start Docker test environment
make test-docker-up
```

### Generate Test Mocks
```bash
# Generate mocks for testing
make mocks
```

## 🔧 Development Commands

```bash
# Show all available commands
make help

# Development workflow
make format          # Format code
make lint           # Run linter
make build          # Build application
make clean          # Clean build artifacts

# Hot reload development (requires air)
make dev

# Database migrations
make migrate-create NAME=migration_name
make migrate-up
make migrate-down
```

## 📚 API Documentation

### Swagger/OpenAPI Documentation

The API is fully documented using Swagger/OpenAPI specifications. 

**Access Swagger UI:**
- Start the server: `make run-dev` or `go run cmd/server/main.go`
- Open: `http://localhost:7007/swagger/index.html`

**Regenerate Documentation:**
After adding or modifying API endpoint annotations in handlers, regenerate the docs:

```bash
# Regenerate swagger documentation
/Users/i572719/go/bin/swag init -g cmd/server/main.go
```

*Note: This updates `docs/docs.go`, `docs/swagger.json`, and `docs/swagger.yaml` files.*

### Authentication Endpoints

The following authentication endpoints are documented in Swagger:

- `GET /api/auth/{provider}/start` - Start OAuth authentication flow
- `GET /api/auth/{provider}/handler/frame` - Handle OAuth callback  
- `GET /api/auth/{provider}/refresh` - Refresh authentication token
- `POST /api/auth/{provider}/logout` - Logout user
- `POST /api/auth/validate` - Validate JWT token

All authenticated API endpoints (under `/api/v1/`) require a valid JWT token.

## 📝 Configuration

Configuration is managed through environment variables and YAML files.

### Environment Variables (.env)

Copy the example file and customize as needed:
```bash
cp .env.example .env
```

Available environment variables:
```bash
# Server Configuration (defaults shown)
ENVIRONMENT=development
PORT=7008
LOG_LEVEL=info

# Database Configuration
DATABASE_URL=postgres://postgres:postgres@localhost:5432/developer_portal?sslmode=disable
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=developer_portal

# JWT Configuration
JWT_SECRET=your-secret-key-change-in-production

# CORS Configuration
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080

# GitHub OAuth (set these for authentication)
GITHUB_TOOLS_APP_CLIENT_ID=
GITHUB_TOOLS_APP_CLIENT_SECRET=
GITHUB_WDF_APP_CLIENT_ID=
GITHUB_WDF_APP_CLIENT_SECRET=
```

### Auth Configuration (config/auth.yaml)
```yaml
default_environment: "production"
redirect_url: "http://localhost:7008"  # Should match your PORT setting

providers:
  githubtools:
    environments:
      development:
        client_id: "${GITHUB_TOOLS_APP_CLIENT_ID_LOCAL}"
        client_secret: "${GITHUB_TOOLS_APP_CLIENT_SECRET_LOCAL}"
        enterprise_base_url: "https://github.tools.sap"
      production:
        client_id: "${GITHUB_TOOLS_APP_CLIENT_ID}"
        client_secret: "${GITHUB_TOOLS_APP_CLIENT_SECRET}"
        enterprise_base_url: "https://github.tools.sap"
```

## 🏗️ Project Structure Explained

This project follows **Clean Architecture** principles with clear separation of concerns:

### Core Layers

#### 📁 `cmd/server/` - Application Entry Point
- Contains `main.go` - bootstraps the application
- Loads configuration, establishes connections, starts server
- **Rule**: Only bootstrap code, no business logic

#### 📁 `internal/api/handlers/` - HTTP Controllers
- Handle HTTP requests/responses, input validation
- Delegate business logic to services
- **Rule**: Thin layer - only HTTP concerns

#### 📁 `internal/service/` - Business Logic Layer
- Core business rules, validation, orchestration
- Independent of HTTP and database details
- **Rule**: Contains the application's business logic

#### 📁 `internal/repository/` - Data Access Layer
- Database operations, CRUD, queries
- Maps between database and Go structs
- **Rule**: Only data access, no business rules

#### 📁 `internal/database/models/` - Database Entities
- Database schema, relationships, validation
- GORM models with proper associations
- **Rule**: Represents database structure

### Supporting Components

#### 📁 `internal/auth/` - Authentication
- GitHub OAuth integration
- JWT token management
- Backstage-compatible auth flow

#### 📁 `internal/config/` - Configuration
- Environment variable loading
- Configuration validation
- Centralized config access

#### 📁 `internal/testutils/` - Testing Utilities
- Test helpers, factories, HTTP utilities
- Database test setup and teardown

#### 📁 `internal/mocks/` - Test Mocks
- Generated mocks for testing
- Service and repository mocks

## 🔄 Data Flow Architecture

```
HTTP Request → Auth Middleware → Handler → Service → Repository → Database
                     ↓             ↓         ↓          ↓
HTTP Response ← Auth Middleware ← Handler ← Service ← Repository ← Database
```

### Request Flow:
1. **HTTP Request** arrives
2. **Auth Middleware** validates JWT token
3. **Handler** parses request, validates input
4. **Service** applies business logic
5. **Repository** performs database operations
6. **Database** stores/retrieves data
7. Response flows back through same layers

## 🔒 Security Considerations

- **JWT Authentication**: All API endpoints require valid JWT tokens
- **GitHub OAuth**: Secure authentication flow
- **Environment Variables**: Secrets managed through environment variables
- **CORS**: Configurable allowed origins
- **Input Validation**: Comprehensive request validation
- **SQL Injection Protection**: GORM provides protection

### Production Security Checklist
- [ ] Set secure `JWT_SECRET` (32+ random bytes)
- [ ] Configure production GitHub OAuth apps
- [ ] Restrict `ALLOWED_ORIGINS` to production domains
- [ ] Use HTTPS in production
- [ ] Secure database credentials
- [ ] Enable request logging and monitoring

## 🚀 Production Deployment

### Build for Production
```bash
# Build optimized binary
make build-prod

# Or build Docker image
make docker-build
```

### Environment Setup
```bash
# Generate secure JWT secret
export JWT_SECRET=$(openssl rand -base64 32)

# Set production environment
export ENVIRONMENT=production

# Configure GitHub OAuth
export GITHUB_TOOLS_APP_CLIENT_ID=your_production_client_id
export GITHUB_TOOLS_APP_CLIENT_SECRET=your_production_client_secret

# Set production database URL
export DATABASE_URL=your_production_database_url

# Restrict CORS origins
export ALLOWED_ORIGINS=https://your-frontend-domain.com
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linting (`make test-precommit`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Development Guidelines
- Follow Go best practices and conventions
- Write tests for new features (aim for >80% coverage)
- Update documentation as needed
- Use meaningful commit messages
- Ensure code passes all checks (`make test-precommit`)

## 🐛 Troubleshooting

### Common Issues

**1. Database Connection Failed**
```bash
# Check if database is running
make db-up

# Verify connection
docker-compose -f docker/docker-compose.yml logs postgres
```

**2. JWT Secret Missing**
```bash
# For development
make run-dev  # Auto-generates secret

# For production
export JWT_SECRET=$(openssl rand -base64 32)
make run
```

**3. Authentication Issues**
- Check GitHub OAuth app configuration
- Verify client ID/secret environment variables
- Ensure redirect URLs match configuration

**4. Port Already in Use**
```bash
# Change port in .env
PORT=8080

# Or kill process using port
lsof -ti:7008 | xargs kill
```

**5. Migration Errors**
```bash
# Check migration files
ls internal/database/migrations/

# Force migration version if needed
make migrate-force VERSION=1
```

## 📚 Additional Resources

- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [Backstage Authentication](https://backstage.io/docs/auth/)
- [Docker Compose](https://docs.docker.com/compose/)

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙋‍♂️ Support

For support and questions:
- Create an issue in the repository
- Check existing documentation
- Review troubleshooting section

---

**Happy Coding! 🚀**
