# Kloudlite v2 Platform Architecture

## Table of Contents
1. [System Overview](#system-overview)
2. [Architecture Diagram](#architecture-diagram)
3. [Service Architecture](#service-architecture)
4. [User Roles & Permissions](#user-roles--permissions)
5. [Team Management Workflow](#team-management-workflow)
6. [API Reference](#api-reference)
7. [Database Schema](#database-schema)
8. [Setup Guide](#setup-guide)

## System Overview

Kloudlite v2 is a cloud-native platform engineering system that provides:
- **Multi-tenant team management** with approval workflows
- **Role-based access control** (RBAC) at platform and team levels
- **OAuth integration** for Google, GitHub, and Microsoft
- **Email verification** and password reset flows
- **gRPC-based microservices** architecture

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                           KLOUDLITE PLATFORM                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────┐                    ┌─────────────────┐       │
│  │                 │                    │                 │       │
│  │   WEB CLIENT    │◄──────────────────►│   API GATEWAY   │       │
│  │   (Next.js)     │    HTTP/gRPC       │                 │       │
│  │                 │                    │                 │       │
│  └─────────────────┘                    └────────┬────────┘       │
│                                                  │                 │
│                                                  │ gRPC            │
│  ┌───────────────────────────────────────────────┴──────────┐     │
│  │                                                          │     │
│  │  ┌──────────────┐    ┌──────────────┐    ┌──────────┐  │     │
│  │  │              │    │              │    │          │  │     │
│  │  │ AUTH SERVICE │◄──►│   ACCOUNTS   │◄──►│   IAM    │  │     │
│  │  │              │    │   SERVICE    │    │ SERVICE  │  │     │
│  │  │              │    │              │    │          │  │     │
│  │  └──────┬───────┘    └──────┬───────┘    └────┬─────┘  │     │
│  │         │                   │                  │         │     │
│  └─────────┼───────────────────┼──────────────────┼─────────┘     │
│            │                   │                  │                │
│  ┌─────────▼───────────────────▼──────────────────▼─────────┐     │
│  │                                                          │     │
│  │  ┌──────────────┐    ┌──────────────┐    ┌──────────┐  │     │
│  │  │              │    │              │    │          │  │     │
│  │  │   MongoDB    │    │     NATS     │    │  Redis   │  │     │
│  │  │              │    │   (KV Store) │    │ (Cache)  │  │     │
│  │  │              │    │              │    │          │  │     │
│  │  └──────────────┘    └──────────────┘    └──────────┘  │     │
│  │                                                          │     │
│  └──────────────────────────────────────────────────────────┘     │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
kloudlite-v2/
├── api/
│   ├── apps/
│   │   ├── auth/                 # Authentication service
│   │   │   ├── internal/
│   │   │   │   ├── domain/       # Business logic
│   │   │   │   ├── entities/     # Data models
│   │   │   │   └── app/
│   │   │       ├── grpc/         # gRPC handlers
│   │   │       └── email/        # Email templates
│   │   │
│   │   ├── accounts/             # Teams & platform management
│   │   │   ├── internal/
│   │   │   │   ├── domain/       # Business logic
│   │   │   │   ├── entities/     # Data models
│   │   │   │   └── app/
│   │   │       └── grpc/         # gRPC handlers
│   │   │
│   │   ├── iam/                  # Identity & access management
│   │   └── runner/               # Service orchestrator
│   │
│   ├── grpc-proto/               # Protocol buffer definitions
│   │   ├── auth.external.proto
│   │   ├── auth-internal.proto
│   │   ├── accounts.external.proto
│   │   └── accounts-internal.proto
│   │
│   └── pkg/                      # Shared packages
│       ├── repos/                # Database repositories
│       ├── errors/               # Error handling
│       └── kv/                   # Key-value storage
│
├── web/                          # Frontend application
│   ├── app/                      # Next.js app router
│   │   ├── auth/                 # Authentication pages
│   │   ├── teams/                # Team management
│   │   ├── platform/             # Platform admin
│   │   └── overview/             # Dashboard
│   │
│   ├── components/               # React components
│   ├── lib/                      # Utilities
│   │   ├── auth/                 # Auth helpers
│   │   └── grpc/                 # gRPC clients
│   │
│   └── grpc/                     # Generated TypeScript types
│
├── docs/                         # Documentation
└── helm-charts/                  # Kubernetes deployments
```

## Service Architecture

### 1. Auth Service
**Purpose**: Handles authentication, JWT tokens, and platform users

**Key Features**:
- JWT token generation and validation
- OAuth provider integration
- Email verification workflow
- Password reset functionality
- Platform user management (super_admin, admin, user roles)

**Database Collections**:
- `users` - User accounts
- `platform-users` - Platform-level user roles
- `device_flows` - CLI authentication flows

**NATS KV Buckets**:
- `verify-token` - Email verification tokens
- `reset-token` - Password reset tokens

### 2. Accounts Service
**Purpose**: Manages teams, platform settings, and approval workflows

**Key Features**:
- Team CRUD operations
- Team approval workflow for non-admin users
- Platform settings management
- Team membership management

**Database Collections**:
- `teams` - Team information
- `team-memberships` - User-team relationships
- `team-approval-requests` - Pending team requests
- `platform-settings` - Global platform configuration
- `invitations` - Team invitations
- `accounts` - Legacy account data

### 3. IAM Service
**Purpose**: Role-based access control and permissions

**Key Features**:
- Resource-based permissions
- Role definitions (owner, admin, member)
- Permission checking for operations

## User Roles & Permissions

### Platform Roles

| Role | Description | Permissions |
|------|-------------|-------------|
| **super_admin** | Platform owner | • Full platform control<br>• Manage all settings<br>• Manage platform users<br>• Create teams without approval<br>• Approve/reject team requests |
| **admin** | Platform administrator | • View platform settings<br>• Create teams without approval<br>• Approve/reject team requests<br>• Cannot manage platform users |
| **user** | Regular user | • Request team creation<br>• View own teams<br>• View own requests<br>• Limited platform access |

### Team Roles

| Role | Description | Permissions |
|------|-------------|-------------|
| **owner** | Team creator | • Full team control<br>• Delete team<br>• Manage members<br>• All team resources |
| **admin** | Team administrator | • Manage team settings<br>• Invite/remove members<br>• Manage resources |
| **member** | Team member | • View team resources<br>• Limited management |

## Team Management Workflow

### For Platform Admins

```
Platform Admin
     │
     ├─► Create Team ──► Team Created Immediately
     │
     └─► View Requests ──┬─► Approve ──► Team Created
                         │
                         └─► Reject ──► Request Denied
```

### For Regular Users

```
Regular User
     │
     ├─► Request Team Creation
     │         │
     │         ▼
     │    Request Pending
     │         │
     │         ├─► Admin Approves ──► Team Created
     │         │                      (User becomes owner)
     │         │
     │         └─► Admin Rejects ──► Request Denied
     │                               (With reason)
     │
     └─► View My Requests ──► Track Status
```

## API Reference

### Authentication Endpoints

#### Auth Service (External)
```proto
service Auth {
  // User authentication
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc Signup(SignupRequest) returns (SignupResponse);
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  
  // Password management
  rpc RequestPasswordReset(RequestPasswordResetRequest) returns (RequestPasswordResetResponse);
  rpc ResetPassword(ResetPasswordRequest) returns (ResetPasswordResponse);
  
  // Email verification
  rpc VerifyEmail(VerifyEmailRequest) returns (VerifyEmailResponse);
  rpc ResendVerificationEmail(ResendVerificationEmailRequest) returns (ResendVerificationEmailResponse);
  
  // Device flow (CLI)
  rpc InitiateDeviceFlow(InitiateDeviceFlowRequest) returns (InitiateDeviceFlowResponse);
  rpc CompleteDeviceFlow(CompleteDeviceFlowRequest) returns (CompleteDeviceFlowResponse);
}
```

#### Auth Service (Internal)
```proto
service AuthInternal {
  // Platform user management
  rpc GetPlatformUser(GetPlatformUserRequest) returns (GetPlatformUserResponse);
  rpc ListPlatformUsers(ListPlatformUsersRequest) returns (ListPlatformUsersResponse);
  rpc UpdatePlatformUserRole(UpdatePlatformUserRoleRequest) returns (UpdatePlatformUserRoleResponse);
  
  // Email service
  rpc SendEmail(SendEmailRequest) returns (SendEmailResponse);
}
```

### Team Management Endpoints

#### Accounts Service (External)
```proto
service Accounts {
  // Platform management
  rpc GetPlatformRole(GetPlatformRoleRequest) returns (GetPlatformRoleResponse);
  rpc GetPlatformSettings(GetPlatformSettingsRequest) returns (GetPlatformSettingsResponse);
  rpc UpdatePlatformSettings(UpdatePlatformSettingsRequest) returns (UpdatePlatformSettingsResponse);
  
  // Team operations
  rpc CreateTeam(CreateTeamRequest) returns (CreateTeamResponse);
  rpc ListTeams(ListTeamsRequest) returns (ListTeamsResponse);
  rpc GetTeamDetails(GetTeamDetailsRequest) returns (GetTeamDetailsResponse);
  
  // Team approval workflow
  rpc RequestTeamCreation(RequestTeamCreationRequest) returns (RequestTeamCreationResponse);
  rpc ListTeamRequests(ListTeamRequestsRequest) returns (ListTeamRequestsResponse);
  rpc GetTeamRequest(GetTeamRequestRequest) returns (GetTeamRequestResponse);
  rpc ApproveTeamRequest(ApproveTeamRequestRequest) returns (ApproveTeamRequestResponse);
  rpc RejectTeamRequest(RejectTeamRequestRequest) returns (RejectTeamRequestResponse);
  
  // Slug utilities
  rpc CheckTeamSlugAvailability(CheckTeamSlugAvailabilityRequest) returns (CheckTeamSlugAvailabilityResponse);
  rpc GenerateTeamSlugSuggestions(GenerateTeamSlugSuggestionsRequest) returns (GenerateTeamSlugSuggestionsResponse);
}
```

## Database Schema

### Core Entities

#### User (auth service)
```javascript
{
  _id: ObjectId,
  email: String,           // Unique email
  name: String,
  password: String,        // Hashed
  passwordSalt: String,
  verified: Boolean,       // Email verified
  approved: Boolean,       // Platform approved
  joined: Date,
  metadata: {
    lastLogin: Date,
    loginCount: Number
  }
}
```

#### PlatformUser (auth service)
```javascript
{
  _id: ObjectId,
  userId: ObjectId,        // Reference to User
  email: String,
  role: String,           // super_admin | admin | user
  isActive: Boolean,
  createdAt: Date,
  updatedAt: Date
}
```

#### Team (accounts service)
```javascript
{
  _id: ObjectId,
  slug: String,           // Unique identifier
  displayName: String,
  description: String,
  logo: String,
  region: String,
  ownerId: ObjectId,      // User who owns the team
  isActive: Boolean,
  contactEmail: String,
  createdAt: Date,
  updatedAt: Date
}
```

#### TeamApprovalRequest (accounts service)
```javascript
{
  _id: ObjectId,
  teamSlug: String,
  displayName: String,
  teamDescription: String,
  teamRegion: String,
  status: String,         // pending | approved | rejected
  requestedBy: ObjectId,
  requestedByEmail: String,
  requestedAt: Date,
  reviewedBy: ObjectId,   // Admin who reviewed
  reviewedByEmail: String,
  reviewedAt: Date,
  rejectionReason: String
}
```

#### PlatformSettings (accounts service)
```javascript
{
  _id: "platform-settings",  // Singleton
  platformOwnerEmail: String,
  supportEmail: String,
  allowSignup: Boolean,
  oauthProviders: {
    google: { enabled: Boolean, clientId: String, clientSecret: String },
    github: { enabled: Boolean, clientId: String, clientSecret: String },
    microsoft: { enabled: Boolean, clientId: String, clientSecret: String }
  },
  teamSettings: {
    requireApproval: Boolean,
    autoApproveFirstTeam: Boolean,
    maxTeamsPerUser: Number
  },
  features: {
    enableDeviceFlow: Boolean,
    enableCLI: Boolean,
    enableAPI: Boolean
  }
}
```

## Setup Guide

### Prerequisites

- MongoDB (localhost:27017)
- NATS Server (localhost:4222)
- Node.js 18+ and pnpm
- Go 1.21+

### Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/kloudlite/kloudlite-v2
   cd kloudlite-v2
   ```

2. **Configure environment**
   ```bash
   # Backend
   cd api/apps/runner
   cp .env.example .env
   # Edit .env with your settings
   
   # Frontend
   cd ../../../web
   cp .env.example .env
   # Edit .env with your settings
   ```

3. **Initialize NATS**
   ```bash
   cd api/apps/runner
   task init-nats
   ```

4. **Start backend**
   ```bash
   task run-local
   ```

5. **Start frontend**
   ```bash
   cd web
   pnpm install
   pnpm dev
   ```

6. **Access the platform**
   - Web UI: http://localhost:3000
   - Platform owner can login with configured email
   - Check email for password reset link

### First-Time Setup

1. Platform automatically initializes with owner email
2. Owner receives password reset email
3. Set password and login
4. Configure platform settings at `/platform`
5. Start inviting users or enabling signup

## Security Considerations

1. **Authentication**
   - JWT tokens with 15-minute expiry
   - Refresh tokens with 7-day expiry
   - Secure password hashing with salt

2. **Authorization**
   - Role-based access at platform level
   - Resource-based permissions at team level
   - All gRPC calls require authentication

3. **Data Protection**
   - Email verification required
   - OAuth provider secrets encrypted
   - Session tokens in secure cookies

4. **Network Security**
   - gRPC for internal communication
   - HTTPS for web traffic
   - IPv4/IPv6 support

## Monitoring & Debugging

### Health Checks
- Backend: `curl localhost:50061/health`
- MongoDB: `mongo --eval "db.adminCommand('ping')"`
- NATS: `curl localhost:8222/varz`

### Common Issues

| Issue | Solution |
|-------|----------|
| Connection refused | Check if services are running on correct ports |
| Platform not initialized | Verify PLATFORM_OWNER_EMAIL in .env |
| Email not sending | Check Mailtrap configuration |
| Team creation fails | Verify unique slug and user permissions |

### Debug Commands

```bash
# View backend logs
tail -f api/apps/runner/backend.log

# Check platform users
mongo kloudlite --eval "db['platform-users'].find().pretty()"

# List pending team requests
mongo kloudlite --eval "db['team-approval-requests'].find({status: 'pending'}).pretty()"

# Verify platform settings
mongo kloudlite --eval "db['platform-settings'].findOne()"
```