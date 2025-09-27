# API Conventions

This document outlines the conventions for APIs in Kloudlite v2, including gRPC services, REST endpoints, and Next.js server actions.

## 🔌 gRPC Conventions

### Proto File Organization
```
/api/grpc-proto/
├── accounts.external.proto    # Public API
├── accounts-internal.proto    # Internal service communication
├── auth.external.proto
├── auth-internal.proto
└── common.proto              # Shared types
```

### Service Definition Standards
```protobuf
syntax = "proto3";

package kloudlite.accounts.v1;
option go_package = "kloudlite.io/rpc/accounts";

// Service documentation
service Accounts {
    // Team Management
    rpc CreateTeam(CreateTeamRequest) returns (CreateTeamResponse);
    rpc GetTeam(GetTeamRequest) returns (GetTeamResponse);
    rpc UpdateTeam(UpdateTeamRequest) returns (UpdateTeamResponse);
    rpc DeleteTeam(DeleteTeamRequest) returns (DeleteTeamResponse);
    rpc ListTeams(ListTeamsRequest) returns (ListTeamsResponse);
    
    // Batch operations
    rpc BatchGetTeams(BatchGetTeamsRequest) returns (BatchGetTeamsResponse);
}
```

### Message Naming Conventions
```protobuf
// Request messages - {Method}Request
message CreateTeamRequest {
    string slug = 1;
    string display_name = 2;
    string description = 3;
    string region = 4;
}

// Response messages - {Method}Response
message CreateTeamResponse {
    Team team = 1;
}

// Shared types - PascalCase
message Team {
    string id = 1;
    string slug = 2;
    string display_name = 3;
    TeamStatus status = 4;
    google.protobuf.Timestamp created_at = 5;
}

// Enums - PascalCase with UPPER_SNAKE values
enum TeamStatus {
    TEAM_STATUS_UNSPECIFIED = 0;
    TEAM_STATUS_ACTIVE = 1;
    TEAM_STATUS_INACTIVE = 2;
    TEAM_STATUS_PENDING = 3;
}
```

### Field Conventions
```protobuf
message User {
    // IDs - string type
    string id = 1;
    string team_id = 2;
    
    // Names - snake_case
    string display_name = 3;
    string email_address = 4;
    
    // Booleans - is/has prefix
    bool is_active = 5;
    bool has_verified_email = 6;
    
    // Timestamps - google.protobuf.Timestamp
    google.protobuf.Timestamp created_at = 7;
    google.protobuf.Timestamp updated_at = 8;
    
    // Lists - repeated
    repeated string roles = 9;
    
    // Maps - map type
    map<string, string> metadata = 10;
    
    // Optional fields - optional keyword
    optional string bio = 11;
}
```

### Error Handling
```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// Standard error responses
func (s *grpcServer) GetTeam(ctx context.Context, req *pb.GetTeamRequest) (*pb.GetTeamResponse, error) {
    // Validation errors
    if req.Slug == "" {
        return nil, status.Error(codes.InvalidArgument, "slug is required")
    }
    
    // Not found
    team, err := s.domain.GetTeam(ctx, req.Slug)
    if err == ErrNotFound {
        return nil, status.Error(codes.NotFound, "team not found")
    }
    
    // Permission denied
    if !s.canAccessTeam(ctx, team) {
        return nil, status.Error(codes.PermissionDenied, "access denied")
    }
    
    // Internal errors
    if err != nil {
        s.logger.Error("failed to get team", "error", err)
        return nil, status.Error(codes.Internal, "internal error")
    }
    
    return &pb.GetTeamResponse{Team: mapTeamToProto(team)}, nil
}
```

### Metadata Handling
```go
// Server-side metadata extraction
func getUserContext(ctx context.Context) UserContext {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return nil
    }
    
    userId := getFirstValue(md, "user-id")
    email := getFirstValue(md, "email")
    roles := md.Get("roles")
    
    return NewUserContext(userId, email, roles)
}

// Client-side metadata injection
func (c *client) callWithAuth(ctx context.Context, method string, req, resp interface{}) error {
    md := metadata.New(map[string]string{
        "user-id": session.UserId,
        "email": session.Email,
        "authorization": "Bearer " + session.Token,
    })
    ctx = metadata.NewOutgoingContext(ctx, md)
    return c.invoke(ctx, method, req, resp)
}
```

## 🚀 Server Actions (Next.js)

### File Organization
```
/app/actions/
├── auth.ts           # Authentication actions
├── teams.ts          # Team management
├── notifications.ts  # Notification handling
└── utils.ts         # Shared utilities
```

### Action Structure
```typescript
'use server'

import { getServerSession } from 'next-auth'
import { revalidatePath, revalidateTag } from 'next/cache'
import { z } from 'zod'

// Input validation schema
const createTeamSchema = z.object({
  slug: z.string().min(3).max(63).regex(/^[a-z0-9-]+$/),
  displayName: z.string().min(1).max(100),
  description: z.string().max(500).optional(),
  region: z.string(),
})

// Type-safe input
type CreateTeamInput = z.infer<typeof createTeamSchema>

// Server action
export async function createTeam(input: CreateTeamInput) {
  try {
    // 1. Validate session
    const session = await getServerSession()
    if (!session?.user) {
      return { success: false, error: 'Unauthorized' }
    }
    
    // 2. Validate input
    const validated = createTeamSchema.parse(input)
    
    // 3. Call backend service
    const client = getAccountsClient()
    const metadata = await getAuthMetadata()
    
    const response = await new Promise<CreateTeamResponse>((resolve, reject) => {
      client.createTeam(validated, metadata, (error, response) => {
        if (error) reject(error)
        else resolve(response)
      })
    })
    
    // 4. Revalidate caches
    revalidatePath('/teams')
    revalidatePath('/overview')
    revalidateTag('teams')
    
    // 5. Return typed response
    return {
      success: true,
      data: {
        teamId: response.team.id,
        slug: response.team.slug,
      }
    }
  } catch (error) {
    // 6. Error handling
    if (error instanceof z.ZodError) {
      return { success: false, error: 'Invalid input', details: error.errors }
    }
    
    return { 
      success: false, 
      error: error instanceof Error ? error.message : 'Failed to create team' 
    }
  }
}
```

### Response Pattern
```typescript
// Success/Error response type
type ActionResponse<T> = 
  | { success: true; data: T }
  | { success: false; error: string; details?: any }

// Usage in components
const result = await createTeam(formData)
if (result.success) {
  router.push(`/teams/${result.data.teamId}`)
} else {
  setError(result.error)
}
```

### Form Integration
```tsx
// With form action
export function CreateTeamForm() {
  return (
    <form action={createTeam}>
      <input name="slug" required />
      <input name="displayName" required />
      <button type="submit">Create Team</button>
    </form>
  )
}

// With client-side handling
export function CreateTeamForm() {
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  
  async function handleSubmit(formData: FormData) {
    setError(null)
    setLoading(true)
    
    try {
      const result = await createTeam({
        slug: formData.get('slug') as string,
        displayName: formData.get('displayName') as string,
      })
      
      if (!result.success) {
        setError(result.error)
      }
    } finally {
      setLoading(false)
    }
  }
  
  return <form action={handleSubmit}>...</form>
}
```

## 🌐 REST API Conventions (Minimal Use)

### Route Structure
```
/app/api/
├── auth/[...nextauth]/   # NextAuth endpoints
├── webhooks/             # External webhooks
│   └── stripe/
└── public/               # Public APIs
    └── health/
```

### HTTP Methods
```typescript
// GET - Retrieve resources
export async function GET(request: Request) {
  const { searchParams } = new URL(request.url)
  const page = searchParams.get('page') || '1'
  
  // Fetch data
  const data = await fetchData({ page })
  
  return Response.json({ data })
}

// POST - Create resources
export async function POST(request: Request) {
  const body = await request.json()
  
  // Validate
  if (!body.name) {
    return Response.json(
      { error: 'Name is required' },
      { status: 400 }
    )
  }
  
  // Create resource
  const created = await createResource(body)
  
  return Response.json(created, { status: 201 })
}
```

### Error Responses
```typescript
// Consistent error format
interface ApiError {
  error: string
  code?: string
  details?: any
}

// Error handler
function handleApiError(error: unknown): Response {
  if (error instanceof ValidationError) {
    return Response.json({
      error: 'Validation failed',
      code: 'VALIDATION_ERROR',
      details: error.errors
    }, { status: 400 })
  }
  
  if (error instanceof UnauthorizedError) {
    return Response.json({
      error: 'Unauthorized',
      code: 'UNAUTHORIZED'
    }, { status: 401 })
  }
  
  // Default error
  return Response.json({
    error: 'Internal server error',
    code: 'INTERNAL_ERROR'
  }, { status: 500 })
}
```

## 📊 Pagination Standards

### Request Parameters
```typescript
// gRPC
message ListTeamsRequest {
    int32 page = 1;      // 1-based page number
    int32 page_size = 2; // Items per page (default: 20, max: 100)
    string sort_by = 3;  // Field to sort by
    SortOrder sort_order = 4;
}

// Server Action
interface ListOptions {
  page?: number      // Default: 1
  pageSize?: number  // Default: 20
  sortBy?: string    // Default: 'createdAt'
  sortOrder?: 'asc' | 'desc' // Default: 'desc'
}
```

### Response Format
```typescript
// gRPC
message ListTeamsResponse {
    repeated Team teams = 1;
    PageInfo page_info = 2;
}

message PageInfo {
    int32 total_items = 1;
    int32 total_pages = 2;
    int32 current_page = 3;
    int32 page_size = 4;
    bool has_next = 5;
    bool has_previous = 6;
}

// TypeScript
interface PaginatedResponse<T> {
  items: T[]
  pageInfo: {
    totalItems: number
    totalPages: number
    currentPage: number
    pageSize: number
    hasNext: boolean
    hasPrevious: boolean
  }
}
```

## 🔍 Filtering & Search

### Filter Syntax
```typescript
// Complex filters
interface TeamFilter {
  status?: TeamStatus | TeamStatus[]
  region?: string
  createdAfter?: Date
  createdBefore?: Date
  memberCount?: {
    min?: number
    max?: number
  }
}

// Search with filters
export async function searchTeams(query: string, filters?: TeamFilter) {
  const request = {
    query,
    filters: {
      status: filters?.status,
      region: filters?.region,
      dateRange: filters?.createdAfter ? {
        start: filters.createdAfter,
        end: filters.createdBefore
      } : undefined
    }
  }
  
  return grpcClient.searchTeams(request)
}
```

## 🔐 Authentication & Authorization

### API Authentication
```typescript
// Server action auth check
async function requireAuth() {
  const session = await getServerSession()
  if (!session?.user) {
    throw new Error('Unauthorized')
  }
  return session
}

// Team access check
async function requireTeamAccess(teamId: string, role?: TeamRole) {
  const session = await requireAuth()
  const membership = await getTeamMembership(teamId, session.user.id)
  
  if (!membership) {
    throw new Error('Not a team member')
  }
  
  if (role && !hasRole(membership.role, role)) {
    throw new Error('Insufficient permissions')
  }
  
  return { session, membership }
}
```

## 📝 API Documentation

### Proto Comments
```protobuf
// CreateTeam creates a new team for the authenticated user.
// The user becomes the owner of the created team.
//
// Requires authentication.
// Returns ALREADY_EXISTS if slug is taken.
rpc CreateTeam(CreateTeamRequest) returns (CreateTeamResponse);

message CreateTeamRequest {
    // Unique identifier for the team (3-63 chars, lowercase alphanumeric + hyphens)
    string slug = 1;
    
    // Display name for the team (1-100 chars)
    string display_name = 2;
    
    // Optional description (max 500 chars)
    optional string description = 3;
}
```

### TypeScript Documentation
```typescript
/**
 * Creates a new team for the authenticated user
 * 
 * @param input - Team creation parameters
 * @returns Created team data or error
 * 
 * @example
 * const result = await createTeam({
 *   slug: 'my-team',
 *   displayName: 'My Team',
 *   region: 'us-west-2'
 * })
 */
export async function createTeam(input: CreateTeamInput): Promise<ActionResponse<Team>> {
  // Implementation
}
```

## 🚀 Performance Guidelines

1. **Batch Operations** - Support batch APIs for bulk operations
2. **Field Selection** - Allow clients to specify needed fields
3. **Caching** - Use appropriate cache headers and tags
4. **Rate Limiting** - Implement rate limits for public APIs
5. **Compression** - Enable gzip for responses

## 📋 Checklist for New APIs

- [ ] Follow naming conventions
- [ ] Add input validation
- [ ] Implement proper error handling
- [ ] Add authentication/authorization
- [ ] Document with comments
- [ ] Include pagination for lists
- [ ] Add filtering/search where needed
- [ ] Test error scenarios
- [ ] Update TypeScript types
- [ ] Consider backwards compatibility