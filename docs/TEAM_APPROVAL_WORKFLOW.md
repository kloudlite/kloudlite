# Team Approval Workflow

## Overview

The team approval workflow ensures controlled team creation based on user roles. Platform admins can create teams instantly, while regular users must submit requests for approval.

## Visual Workflow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         TEAM CREATION WORKFLOW                               │
└─────────────────────────────────────────────────────────────────────────────┘

    [User Login]
         │
         ▼
    ┌─────────┐
    │ Check   │
    │Platform │────────────────┐
    │ Role    │                │
    └─────────┘                │
         │                     │
         ▼                     ▼
    ┌─────────┐          ┌──────────┐
    │Regular  │          │Platform  │
    │  User   │          │  Admin   │
    └────┬────┘          └────┬─────┘
         │                    │
         ▼                    ▼
   ┌──────────┐         ┌──────────┐
   │ Request  │         │ Create   │
   │  Team    │         │  Team    │
   │Creation  │         │Directly  │
   └────┬─────┘         └────┬─────┘
        │                    │
        ▼                    ▼
   ┌──────────┐         ┌──────────┐
   │ Request  │         │  Team    │
   │ Pending  │         │ Created  │
   └────┬─────┘         └──────────┘
        │
        ├─────────────┐
        ▼             ▼
   [Admin Review]  [User Views Status]
        │
        ├──────────┬──────────┐
        ▼          ▼          ▼
   ┌────────┐ ┌────────┐ ┌────────┐
   │Approve │ │ Reject │ │Pending │
   └────┬───┘ └────┬───┘ └────────┘
        │          │
        ▼          ▼
   ┌────────┐ ┌─────────┐
   │ Team   │ │Request  │
   │Created │ │ Denied  │
   └────────┘ └─────────┘
```

## User Journeys

### 1. Platform Admin Journey

```
Platform Admin
│
├─► Navigate to /teams/new
│   │
│   ├─► Fill team details
│   │   • Display Name: "Engineering Team"
│   │   • Slug: "engineering" (auto-generated)
│   │   • Description: "Core engineering team"
│   │   • Region: "us-west-2"
│   │
│   └─► Click "Create Team"
│       │
│       └─► Team created immediately
│           │
│           └─► Redirected to team dashboard
```

### 2. Regular User Journey

```
Regular User
│
├─► Navigate to /teams/new
│   │
│   ├─► Fill team details
│   │   • Display Name: "Marketing Team"
│   │   • Slug: "marketing"
│   │   • Description: "Digital marketing initiatives"
│   │   • Region: "us-east-1"
│   │
│   └─► Click "Request Team"
│       │
│       ├─► Request submitted
│       │   │
│       │   └─► Email sent to platform admins
│       │
│       └─► Redirected to /teams/requests
│           │
│           └─► View request status (Pending)
```

### 3. Admin Review Journey

```
Platform Admin (Reviewing Requests)
│
├─► Navigate to /platform
│   │
│   └─► Click "Team Requests" tab
│       │
│       ├─► View pending requests
│       │   │
│       │   ├─► Request #1: Marketing Team
│       │   │   • Requested by: user@example.com
│       │   │   • Date: 2024-01-15
│       │   │   • Description: Digital marketing...
│       │   │
│       │   └─► Actions:
│       │       ├─► [Approve] → Team created
│       │       │              → User notified
│       │       │              → User is team owner
│       │       │
│       │       └─► [Reject]  → Provide reason
│       │                      → User notified
│       │                      → Request closed
│       │
│       └─► View approved/rejected history
```

## Implementation Components

### Frontend Pages

```
/teams/new                    # Team creation/request form
├── Form Fields:
│   ├── Display Name         # Required, user-friendly name
│   ├── Slug                 # Auto-generated, URL-safe
│   ├── Description          # Optional team description  
│   └── Region              # Required, deployment region
│
├── Actions (based on role):
│   ├── "Create Team"        # For admins
│   └── "Request Team"       # For users
│
└── Validations:
    ├── Slug availability
    ├── Name uniqueness
    └── Region selection

/teams/requests              # User's team requests
├── Request List:
│   ├── Team Name
│   ├── Status (Pending/Approved/Rejected)
│   ├── Requested Date
│   └── Review Details (if processed)
│
└── Filters:
    ├── Status filter
    └── Date range

/platform                    # Platform management
└── Team Requests Tab:
    ├── Pending Requests:
    │   ├── Team details
    │   ├── Requester info
    │   ├── Request date
    │   └── Action buttons
    │
    ├── Processed Requests:
    │   ├── History log
    │   ├── Reviewer info
    │   └── Decision reason
    │
    └── Statistics:
        ├── Total requests
        ├── Approval rate
        └── Average response time
```

### Backend Flow

```
1. Check User Role
   │
   ├─► Platform Admin?
   │   └─► Direct team creation
   │       └─► Return team ID
   │
   └─► Regular User?
       └─► Create approval request
           └─► Return request ID

2. Admin Reviews Request
   │
   ├─► Validate admin permissions
   │
   ├─► Approve?
   │   ├─► Create team
   │   ├─► Add requester as owner
   │   ├─► Update request status
   │   └─► Send notification
   │
   └─► Reject?
       ├─► Update request status
       ├─► Store rejection reason
       └─► Send notification
```

## Database Operations

### Creating a Team Request

```javascript
// Team Approval Request Document
{
  _id: new ObjectId(),
  teamSlug: "marketing",
  displayName: "Marketing Team",
  teamDescription: "Digital marketing initiatives",
  teamRegion: "us-east-1",
  status: "pending",
  requestedBy: ObjectId("user-id"),
  requestedByEmail: "user@example.com",
  requestedAt: new Date(),
  // Fields populated after review:
  reviewedBy: null,
  reviewedByEmail: null,
  reviewedAt: null,
  rejectionReason: null
}
```

### Approving a Request

```javascript
// 1. Update request status
db.teamApprovalRequests.updateOne(
  { _id: requestId },
  {
    $set: {
      status: "approved",
      reviewedBy: adminUserId,
      reviewedByEmail: "admin@example.com",
      reviewedAt: new Date()
    }
  }
)

// 2. Create team
db.teams.insertOne({
  _id: new ObjectId(),
  slug: "marketing",
  displayName: "Marketing Team",
  description: "Digital marketing initiatives",
  region: "us-east-1",
  ownerId: requestedByUserId,
  isActive: true,
  createdAt: new Date()
})

// 3. Create team membership
db.teamMemberships.insertOne({
  _id: new ObjectId(),
  teamId: teamId,
  userId: requestedByUserId,
  role: "account_owner",
  createdAt: new Date()
})
```

## API Examples

### Request Team Creation (User)

```typescript
// Frontend call
const request = await requestTeamCreation({
  displayName: "Marketing Team",
  slug: "marketing",
  description: "Digital marketing initiatives",
  region: "us-east-1"
})

// Response
{
  requestId: "req-123",
  status: "pending"
}
```

### List Pending Requests (Admin)

```typescript
// Frontend call
const requests = await listTeamRequests({
  status: "pending"
})

// Response
{
  requests: [
    {
      requestId: "req-123",
      slug: "marketing",
      displayName: "Marketing Team",
      description: "Digital marketing initiatives",
      region: "us-east-1",
      status: "pending",
      requestedBy: "user-id",
      requestedByEmail: "user@example.com",
      requestedAt: "2024-01-15T10:00:00Z"
    }
  ]
}
```

### Approve Request (Admin)

```typescript
// Frontend call
const result = await approveTeamRequest({
  requestId: "req-123"
})

// Response
{
  teamId: "team-456",
  success: true
}
```

### Reject Request (Admin)

```typescript
// Frontend call
const result = await rejectTeamRequest({
  requestId: "req-123",
  reason: "Team name conflicts with existing project"
})

// Response
{
  success: true
}
```

## Configuration Options

### Platform Settings

```javascript
{
  teamSettings: {
    requireApproval: true,        // Force approval for all users
    autoApproveFirstTeam: false,  // Auto-approve user's first team
    maxTeamsPerUser: 5           // Limit teams per user
  }
}
```

### Role Permissions

| Action | Super Admin | Admin | User |
|--------|-------------|-------|------|
| Create team directly | ✅ | ✅ | ❌ |
| Request team creation | ✅ | ✅ | ✅ |
| View all requests | ✅ | ✅ | ❌ |
| View own requests | ✅ | ✅ | ✅ |
| Approve requests | ✅ | ✅ | ❌ |
| Reject requests | ✅ | ✅ | ❌ |

## Error Handling

### Common Errors

```typescript
// Slug already exists
{
  error: "SLUG_EXISTS",
  message: "Team slug 'marketing' is already taken",
  suggestions: ["marketing-team", "marketing-dept", "mkt-team"]
}

// Unauthorized approval attempt
{
  error: "UNAUTHORIZED",
  message: "Only platform admins can approve team requests"
}

// Request not found
{
  error: "NOT_FOUND",
  message: "Team request 'req-123' not found"
}

// Max teams reached
{
  error: "LIMIT_EXCEEDED",
  message: "User has reached maximum team limit (5)"
}
```

## Future Enhancements

1. **Automated Approvals**
   - Auto-approve based on user history
   - Domain-based auto-approval
   - First team auto-approval

2. **Enhanced Notifications**
   - Email notifications for new requests
   - Slack/Discord integration
   - In-app notifications

3. **Request Templates**
   - Pre-defined team templates
   - Department-based defaults
   - Quick approval for templates

4. **Analytics Dashboard**
   - Request volume trends
   - Approval/rejection rates
   - Average processing time
   - User activity metrics