# Database Conventions

This document outlines the conventions for database design, MongoDB usage, and data access patterns in Kloudlite v2.

## 📊 MongoDB Conventions

### Collection Naming
- **Format**: Lowercase with hyphens
- **Plural**: Use plural forms for collections
- **Examples**:
  - `users`
  - `teams`
  - `team-memberships`
  - `team-approval-requests`
  - `platform-users`
  - `notifications`

### Document Structure

#### Base Entity Pattern
```go
type BaseEntity struct {
    ID        repos.ID  `json:"id" bson:"_id"`
    CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
    CreatedBy string    `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
    UpdatedBy string    `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"`
}
```

#### Entity Example
```go
type Team struct {
    repos.BaseEntity `json:",inline" bson:",inline"`
    
    // Core fields
    Slug        string     `json:"slug" bson:"slug"`
    DisplayName string     `json:"displayName" bson:"displayName"`
    Description string     `json:"description,omitempty" bson:"description,omitempty"`
    
    // Status and metadata
    Status      TeamStatus `json:"status" bson:"status"`
    Region      string     `json:"region" bson:"region"`
    
    // Relationships
    OwnerUserId repos.ID   `json:"ownerUserId" bson:"ownerUserId"`
    
    // Additional data
    Metadata    map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
}
```

### Field Naming
- **Format**: camelCase in MongoDB
- **JSON tags**: Match field names
- **BSON tags**: Match field names
- **Omitempty**: Use for optional fields

```go
type User struct {
    Email         string    `json:"email" bson:"email"`
    DisplayName   string    `json:"displayName" bson:"displayName"`
    Bio           string    `json:"bio,omitempty" bson:"bio,omitempty"`
    EmailVerified bool      `json:"emailVerified" bson:"emailVerified"`
    LastLoginAt   time.Time `json:"lastLoginAt,omitempty" bson:"lastLoginAt,omitempty"`
}
```

## 🔍 Indexing Strategy

### Index Types

#### Single Field Index
```go
{
    Name: "slug_unique",
    Key: repos.IndexKey{Key: "slug", Value: repos.IndexAsc},
    Unique: true,
}
```

#### Compound Index
```go
{
    Name: "team_status_created",
    Key: []repos.IndexKey{
        {Key: "teamId", Value: repos.IndexAsc},
        {Key: "status", Value: repos.IndexAsc},
        {Key: "createdAt", Value: repos.IndexDesc},
    },
}
```

#### Text Index
```go
{
    Name: "search_idx",
    Key: []repos.IndexKey{
        {Key: "displayName", Value: repos.IndexText},
        {Key: "description", Value: repos.IndexText},
    },
}
```

### Index Naming Convention
- **Format**: `{field1}_{field2}_{type}`
- **Types**: `unique`, `idx`, `text`
- **Examples**:
  - `slug_unique`
  - `email_unique`
  - `teamId_status_idx`
  - `displayName_text`

### Index Guidelines
1. **Create indexes for**:
   - Unique constraints
   - Foreign key lookups
   - Common query filters
   - Sort fields

2. **Compound indexes**:
   - Order matters (most selective first)
   - Can serve queries on prefixes
   - Consider query patterns

3. **Avoid over-indexing**:
   - Each index has storage cost
   - Impacts write performance
   - Monitor index usage

## 🔄 Repository Pattern

### Interface Definition
```go
type TeamRepository interface {
    // Basic CRUD
    Create(ctx context.Context, team *entities.Team) (*entities.Team, error)
    FindById(ctx context.Context, id repos.ID) (*entities.Team, error)
    FindBySlug(ctx context.Context, slug string) (*entities.Team, error)
    Update(ctx context.Context, id repos.ID, updates bson.M) (*entities.Team, error)
    Delete(ctx context.Context, id repos.ID) error
    
    // Queries
    List(ctx context.Context, filter Filter, opts ...ListOption) ([]*entities.Team, error)
    Count(ctx context.Context, filter Filter) (int64, error)
    
    // Batch operations
    BulkUpdate(ctx context.Context, filter Filter, updates bson.M) (int64, error)
}
```

### Repository Implementation
```go
type mongoTeamRepository struct {
    collection *mongo.Collection
    logger     *slog.Logger
}

func (r *mongoTeamRepository) FindBySlug(ctx context.Context, slug string) (*entities.Team, error) {
    var team entities.Team
    err := r.collection.FindOne(ctx, bson.M{"slug": slug}).Decode(&team)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("failed to find team: %w", err)
    }
    return &team, nil
}
```

## 📝 Query Patterns

### Filter Building
```go
// Simple filter
filter := bson.M{
    "status": "active",
    "region": "us-west-2",
}

// Complex filter with operators
filter := bson.M{
    "createdAt": bson.M{
        "$gte": startDate,
        "$lt":  endDate,
    },
    "memberCount": bson.M{"$gt": 0},
    "$or": []bson.M{
        {"status": "active"},
        {"status": "pending"},
    },
}
```

### Pagination
```go
type ListOptions struct {
    Limit  int
    Offset int
    Sort   []SortOption
}

func (r *repository) List(ctx context.Context, filter bson.M, opts ListOptions) ([]*Entity, error) {
    cursor, err := r.collection.Find(ctx, filter,
        options.Find().
            SetLimit(int64(opts.Limit)).
            SetSkip(int64(opts.Offset)).
            SetSort(buildSort(opts.Sort)),
    )
    // ... handle cursor
}
```

### Projection
```go
// Select specific fields
opts := options.FindOne().SetProjection(bson.M{
    "slug": 1,
    "displayName": 1,
    "status": 1,
    "_id": 1,
})

// Exclude fields
opts := options.FindOne().SetProjection(bson.M{
    "metadata": 0,
    "internalNotes": 0,
})
```

## 🔐 Data Validation

### Schema Validation
```go
type Team struct {
    Slug string `json:"slug" bson:"slug" validate:"required,slug,min=3,max=63"`
    DisplayName string `json:"displayName" bson:"displayName" validate:"required,min=1,max=100"`
    Email string `json:"email" bson:"email" validate:"required,email"`
}

func ValidateTeam(team *Team) error {
    validate := validator.New()
    return validate.Struct(team)
}
```

### Custom Validators
```go
// Register custom validation
validate.RegisterValidation("slug", ValidateSlug)

func ValidateSlug(fl validator.FieldLevel) bool {
    slug := fl.Field().String()
    match, _ := regexp.MatchString("^[a-z0-9-]+$", slug)
    return match && !strings.HasPrefix(slug, "-") && !strings.HasSuffix(slug, "-")
}
```

## 🔄 Migration Strategy

### Migration Files
```
/api/migrations/
├── 001_initial_schema.go
├── 002_add_team_regions.go
├── 003_add_notifications.go
└── migrate.go
```

### Migration Pattern
```go
type Migration struct {
    Version     int
    Description string
    Up          func(db *mongo.Database) error
    Down        func(db *mongo.Database) error
}

var migrations = []Migration{
    {
        Version:     1,
        Description: "Create initial indexes",
        Up: func(db *mongo.Database) error {
            return createIndexes(db)
        },
        Down: func(db *mongo.Database) error {
            return dropIndexes(db)
        },
    },
}
```

## 🚀 Performance Best Practices

### Query Optimization
1. **Use indexes** for all query patterns
2. **Limit results** with pagination
3. **Project only needed fields**
4. **Use aggregation pipeline** for complex queries
5. **Monitor slow queries** with profiler

### Write Optimization
```go
// Bulk writes
models := []mongo.WriteModel{
    mongo.NewInsertOneModel().SetDocument(doc1),
    mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update),
}
result, err := collection.BulkWrite(ctx, models)

// Upsert pattern
opts := options.Update().SetUpsert(true)
result, err := collection.UpdateOne(ctx, filter, update, opts)
```

### Connection Pooling
```go
clientOptions := options.Client().
    ApplyURI(uri).
    SetMaxPoolSize(100).
    SetMinPoolSize(10).
    SetMaxConnIdleTime(60 * time.Second)
```

## 🔍 Aggregation Patterns

### Common Pipelines
```go
// Count by status
pipeline := []bson.M{
    {"$match": bson.M{"teamId": teamId}},
    {"$group": bson.M{
        "_id": "$status",
        "count": bson.M{"$sum": 1},
    }},
}

// Join with lookup
pipeline := []bson.M{
    {"$match": bson.M{"status": "active"}},
    {"$lookup": bson.M{
        "from": "users",
        "localField": "ownerUserId",
        "foreignField": "_id",
        "as": "owner",
    }},
    {"$unwind": "$owner"},
}
```

## 🛡️ Data Security

### Field Level Security
```go
// Sensitive fields
type User struct {
    Email          string `json:"email" bson:"email"`
    HashedPassword string `json:"-" bson:"hashedPassword"` // Never expose
    APIKey         string `json:"-" bson:"apiKey,omitempty"`
}
```

### Audit Fields
```go
type AuditableEntity struct {
    CreatedBy string    `json:"createdBy" bson:"createdBy"`
    CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
    UpdatedBy string    `json:"updatedBy" bson:"updatedBy"`
    UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
    DeletedAt *time.Time `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
}
```

## 📋 Naming Conventions Summary

| Type | Convention | Example |
|------|-----------|---------|
| Collection | lowercase-hyphen (plural) | `team-memberships` |
| Field | camelCase | `displayName` |
| Index | field_type | `email_unique` |
| Repository method | VerbNoun | `FindBySlug` |
| Filter variable | descriptiveName | `activeTeamsFilter` |

## 🚨 Common Pitfalls to Avoid

1. **Missing indexes** on foreign keys
2. **N+1 queries** - use aggregation
3. **Unbounded queries** - always paginate
4. **Large documents** - consider splitting
5. **Inconsistent naming** - follow conventions
6. **No validation** - validate at boundaries
7. **Exposing internal fields** - use projections