# Backend Conventions (Go)

This document outlines the conventions and standards for backend Go services in Kloudlite v2.

## 📁 Directory Structure

### Service Layout
```
/api/apps/{service-name}/
├── Dockerfile              # Container definition
├── Taskfile.yml           # Task automation
├── README.md              # Service documentation
├── fx-app/
│   └── main.go            # Dependency injection setup
├── internal/              # Private packages
│   ├── app/               # Application layer
│   │   ├── app.go         # App initialization
│   │   ├── grpc/          # gRPC server implementation
│   │   │   ├── grpc-server.go
│   │   │   └── main.go
│   │   └── {feature}/     # Feature-specific logic
│   ├── domain/            # Business logic layer
│   │   ├── domain.go      # Domain interfaces
│   │   ├── impl.go        # Domain implementation
│   │   ├── ports.go       # External interfaces
│   │   └── {feature}.go   # Feature domains
│   ├── entities/          # Data models
│   │   ├── {entity}.go    # Entity definitions
│   │   └── field-constants/
│   │       └── generated_constants.go
│   └── env/               # Environment configuration
│       └── env.go
└── main.go                # Service entry point
```

## 📝 File Naming Conventions

### Go Files
- **Format**: lowercase with hyphens
- **Examples**: 
  - `team-approval.go`
  - `platform-settings.go`
  - `grpc-server.go`

### Proto Files
- **Format**: lowercase with dots for namespacing
- **Examples**:
  - `accounts.external.proto` (public API)
  - `accounts-internal.proto` (internal API)
  - `auth.external.proto`

### Test Files
- **Format**: `{filename}_test.go`
- **Location**: Same directory as source
- **Examples**:
  - `teams_test.go`
  - `domain_test.go`

## 🏗️ Architecture Patterns

### Clean Architecture Layers

#### 1. Domain Layer (`/internal/domain/`)
- **Purpose**: Core business logic
- **Rules**:
  - NO external dependencies
  - NO framework-specific code
  - Pure Go interfaces and structs
  - Business rules and validations

```go
// domain.go - Interface definitions
type TeamService interface {
    CreateTeam(ctx UserContext, team entities.Team) (*entities.Team, error)
    GetTeam(ctx UserContext, slug string) (*entities.Team, error)
    UpdateTeam(ctx UserContext, team entities.Team) (*entities.Team, error)
    DeleteTeam(ctx UserContext, slug string) error
}

// teams.go - Implementation
type teamService struct {
    repo      repos.DbRepo[*entities.Team]
    logger    *slog.Logger
}

func (s *teamService) CreateTeam(ctx UserContext, team entities.Team) (*entities.Team, error) {
    // Business logic here
}
```

#### 2. Application Layer (`/internal/app/`)
- **Purpose**: Use cases and orchestration
- **Responsibilities**:
  - gRPC server implementation
  - External service integration
  - Request/response mapping
  - Transaction coordination

```go
// grpc/grpc-server.go
type grpcServer struct {
    accountsv1.UnimplementedAccountsServer
    domain     domain.Domain
    logger     *slog.Logger
}

func (s *grpcServer) CreateTeam(ctx context.Context, req *accountsv1.CreateTeamRequest) (*accountsv1.CreateTeamResponse, error) {
    userCtx := getUserContext(ctx)
    team, err := s.domain.Teams().CreateTeam(userCtx, mapRequestToEntity(req))
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    return mapEntityToResponse(team), nil
}
```

#### 3. Entities Layer (`/internal/entities/`)
- **Purpose**: Data structures
- **Rules**:
  - Simple structs with tags
  - No business logic
  - Validation tags allowed

```go
type Team struct {
    repos.BaseEntity `json:",inline" bson:",inline"`
    
    Slug        string            `json:"slug" bson:"slug"`
    DisplayName string            `json:"displayName" bson:"displayName"`
    Description string            `json:"description,omitempty" bson:"description,omitempty"`
    Status      TeamStatus        `json:"status" bson:"status"`
    Region      string            `json:"region" bson:"region"`
    Metadata    map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
}
```

## 🔧 Coding Standards

### Package Organization
```go
package domain

import (
    // Standard library
    "context"
    "errors"
    "fmt"
    
    // External packages
    "github.com/kloudlite/api/common"
    "go.uber.org/fx"
    
    // Internal packages
    "github.com/kloudlite/api/apps/accounts/internal/entities"
)
```

### Error Handling
```go
// Define errors at package level
var (
    ErrTeamNotFound      = errors.New("team not found")
    ErrInvalidSlug       = errors.New("invalid team slug")
    ErrPermissionDenied  = errors.New("permission denied")
)

// Wrap errors with context
func (s *teamService) GetTeam(ctx UserContext, slug string) (*entities.Team, error) {
    team, err := s.repo.FindOne(ctx, repos.Filter{"slug": slug})
    if err != nil {
        if errors.Is(err, repos.ErrNotFound) {
            return nil, ErrTeamNotFound
        }
        return nil, fmt.Errorf("failed to fetch team: %w", err)
    }
    return team, nil
}
```

### Context Usage
```go
// Always pass context as first parameter
func (s *service) DoSomething(ctx context.Context, param string) error {
    // Use context for cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Do work
    }
}

// Custom context types
type UserContext interface {
    context.Context
    GetUserId() repos.ID
    GetEmail() string
}
```

### Dependency Injection
```go
// fx-app/main.go
var Module = fx.Module(
    "accounts",
    fx.Provide(
        domain.NewDomain,
        env.LoadEnv,
        app.NewApp,
    ),
)

// Constructor pattern
func NewTeamService(
    repo repos.DbRepo[*entities.Team],
    logger *slog.Logger,
) TeamService {
    return &teamService{
        repo:   repo,
        logger: logger,
    }
}
```

## 🔌 gRPC Conventions

### Service Definition
```protobuf
syntax = "proto3";

package kloudlite.accounts.v1;
option go_package = "kloudlite.io/rpc/accounts";

service Accounts {
    // Team management
    rpc CreateTeam(CreateTeamRequest) returns (CreateTeamResponse);
    rpc GetTeam(GetTeamRequest) returns (GetTeamResponse);
    rpc UpdateTeam(UpdateTeamRequest) returns (UpdateTeamResponse);
    rpc DeleteTeam(DeleteTeamRequest) returns (DeleteTeamResponse);
}
```

### Naming Patterns
- **Service**: PascalCase (e.g., `Accounts`, `Auth`)
- **Methods**: VerbNoun pattern (e.g., `CreateTeam`, `ListNotifications`)
- **Messages**: `{Method}Request` and `{Method}Response`
- **Fields**: camelCase with descriptive names

### Error Handling
```go
import "google.golang.org/grpc/codes"
import "google.golang.org/grpc/status"

// Return appropriate gRPC status codes
switch err {
case ErrNotFound:
    return status.Error(codes.NotFound, "resource not found")
case ErrPermissionDenied:
    return status.Error(codes.PermissionDenied, "access denied")
default:
    return status.Error(codes.Internal, "internal error")
}
```

## 📊 Database Patterns

### Repository Pattern
```go
type TeamRepository interface {
    Create(ctx context.Context, team *entities.Team) (*entities.Team, error)
    FindBySlug(ctx context.Context, slug string) (*entities.Team, error)
    Update(ctx context.Context, team *entities.Team) (*entities.Team, error)
    Delete(ctx context.Context, id repos.ID) error
}
```

### Index Management
```go
func GetIndexes() []repos.IndexSpec {
    return []repos.IndexSpec{
        {
            Name: "slug",
            Key:  repos.IndexKey{Key: "slug", Value: repos.IndexAsc},
            Unique: true,
        },
        {
            Name: "status_created",
            Key: []repos.IndexKey{
                {Key: "status", Value: repos.IndexAsc},
                {Key: "createdAt", Value: repos.IndexDesc},
            },
        },
    }
}
```

## 🧪 Testing Standards

### Unit Tests
```go
func TestTeamService_CreateTeam(t *testing.T) {
    // Arrange
    repo := mocks.NewMockTeamRepository()
    service := NewTeamService(repo, slog.Default())
    
    // Act
    team, err := service.CreateTeam(ctx, entities.Team{
        Slug: "test-team",
        DisplayName: "Test Team",
    })
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "test-team", team.Slug)
}
```

### Table-Driven Tests
```go
func TestValidateSlug(t *testing.T) {
    tests := []struct {
        name    string
        slug    string
        wantErr bool
    }{
        {"valid slug", "my-team", false},
        {"empty slug", "", true},
        {"invalid chars", "my team", true},
        {"too long", strings.Repeat("a", 64), true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateSlug(tt.slug)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateSlug() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## 📦 Module Guidelines

### Service Modules
- One module per service
- Clear boundaries between services
- Communicate via gRPC only

### Shared Code
- Place in `/api/pkg/` directory
- No business logic in shared packages
- Only utilities and common types

## 🔒 Security Practices

### Authentication
```go
// Extract user context from gRPC metadata
func getUserContext(ctx context.Context) UserContext {
    md, _ := metadata.FromIncomingContext(ctx)
    userId := md.Get("user-id")[0]
    email := md.Get("email")[0]
    return NewUserContext(ctx, userId, email)
}
```

### Authorization
```go
// Check permissions in domain layer
func (s *teamService) UpdateTeam(ctx UserContext, team entities.Team) error {
    if !s.canUpdateTeam(ctx, team) {
        return ErrPermissionDenied
    }
    // Update logic
}
```

## 📈 Performance Guidelines

### Query Optimization
- Use indexes for common queries
- Limit result sets with pagination
- Project only needed fields

### Concurrency
```go
// Use goroutines for parallel operations
g, ctx := errgroup.WithContext(ctx)

g.Go(func() error {
    return s.fetchUserData(ctx)
})

g.Go(func() error {
    return s.fetchTeamData(ctx)
})

if err := g.Wait(); err != nil {
    return err
}
```

## 🚀 Best Practices

1. **Always validate input** at service boundaries
2. **Log errors** with context for debugging
3. **Use transactions** for multi-step operations
4. **Handle graceful shutdown** properly
5. **Document complex logic** with comments
6. **Keep functions small** and focused
7. **Return early** to reduce nesting
8. **Use meaningful variable names**