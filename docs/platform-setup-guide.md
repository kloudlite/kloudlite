# Kloudlite Platform Setup and Team Management Guide

## Overview

This guide covers the platform setup process and team management features in Kloudlite v2, including the team approval workflow for non-admin users.

## Platform Architecture

### User Roles

The platform supports three user roles:

1. **Super Admin** (`super_admin`)
   - Full platform control
   - Can manage platform settings
   - Can create teams without approval
   - Can approve/reject team requests
   - Can manage all platform users

2. **Admin** (`admin`)
   - Can create teams without approval
   - Can approve/reject team requests
   - Can view platform settings
   - Cannot manage other platform users

3. **User** (`user`)
   - Must request approval to create teams
   - Can view their own team requests
   - Limited platform access

### Service Architecture

- **Auth Service**: Manages platform users, authentication, and email verification
- **Accounts Service**: Manages platform settings, teams, and approval workflows

## Initial Platform Setup

### 1. Environment Configuration

Configure the following environment variables in `/api/apps/runner/.env`:

```env
# MongoDB Configuration
MONGO_URI=mongodb://localhost:27017
MONGO_DB_NAME=kloudlite

# NATS Configuration
NATS_URL=nats://localhost:4222

# Auth Service Configuration
AUTH__USER_EMAIL_VERIFICATION_ENABLED=true
AUTH__PLATFORM_OWNER_EMAIL=your-email@example.com
AUTH__JWT_SECRET=your-secure-jwt-secret
AUTH__MAILTRAP_API_TOKEN=your-mailtrap-token
```

### 2. Initialize NATS KV Buckets

```bash
cd api/apps/runner
task init-nats
```

This creates:
- `reset-token` bucket for password reset tokens
- `verify-token` bucket for email verification tokens

### 3. Start Backend Services

```bash
cd api/apps/runner
task run-local
```

### 4. Platform Initialization

On first run, the platform automatically:
1. Creates platform settings with the owner email
2. Creates a user account for the platform owner
3. Assigns `super_admin` role to the owner
4. Sends a password reset email to set initial password

## Team Management

### For Admins (Creating Teams)

Admins can create teams directly without approval:

1. Navigate to `/teams/new`
2. Fill in team details:
   - Display Name
   - Slug (auto-generated or custom)
   - Description
   - Region
3. Click "Create Team"
4. Team is created immediately

### For Regular Users (Requesting Teams)

Users must request approval to create teams:

1. Navigate to `/teams/new`
2. Fill in team details
3. Click "Request Team"
4. Wait for admin approval

### Platform Admin Workflow

#### Viewing Team Requests

1. Navigate to `/platform`
2. Click on "Team Requests" tab
3. View pending requests with:
   - Team name and slug
   - Requester information
   - Request date
   - Description

#### Approving Team Requests

1. Review the request details
2. Click "Approve" button
3. The system will:
   - Create the team
   - Assign the requester as team owner
   - Update request status to approved
   - Notify the requester (when implemented)

#### Rejecting Team Requests

1. Review the request details
2. Click "Reject" button
3. Provide a rejection reason
4. The requester will be notified with the reason

## Platform Settings Management

Platform admins can manage settings at `/platform`:

### General Settings
- **Platform Owner Email**: Primary administrator email
- **Support Email**: Contact for platform support
- **Allow Signup**: Enable/disable new user registrations

### OAuth Providers
Configure OAuth providers with:
- Google OAuth
- GitHub OAuth
- Microsoft Azure AD

### Team Settings
- **Require Approval**: Force approval workflow for all users
- **Auto Approve First Team**: Allow users to create their first team without approval
- **Max Teams Per User**: Limit team creation per user

### Platform Features
- **Enable Device Flow**: Allow CLI authentication
- **Enable CLI**: Enable command-line interface access
- **Enable API**: Enable programmatic API access

## API Endpoints

### Accounts Service (gRPC)

#### Platform Management
- `GetPlatformRole`: Check user's platform role
- `GetPlatformSettings`: Retrieve platform configuration
- `UpdatePlatformSettings`: Update platform configuration (admin only)

#### Team Management
- `CreateTeam`: Direct team creation (admin) or request (user)
- `RequestTeamCreation`: Submit team approval request
- `ListTeamRequests`: View team requests (filtered by role)
- `ApproveTeamRequest`: Approve a team request (admin only)
- `RejectTeamRequest`: Reject a team request (admin only)

### Auth Service (gRPC)

#### Platform User Management
- `GetPlatformUser`: Get platform user details
- `ListPlatformUsers`: List all platform users (internal)
- `UpdatePlatformUserRole`: Change user role (internal)

## Security Considerations

1. **JWT Authentication**: All API calls require valid JWT tokens
2. **Role-Based Access**: Operations restricted based on user roles
3. **Email Verification**: Users must verify email before accessing protected routes
4. **Secure Communication**: All services communicate via gRPC with authentication

## Troubleshooting

### Common Issues

1. **Connection Refused Errors**
   - Ensure backend services are running
   - Check if using correct ports (50061 for gRPC)
   - Verify IPv4 vs IPv6 connectivity

2. **Platform Not Initialized**
   - Check `AUTH__PLATFORM_OWNER_EMAIL` is set
   - Verify MongoDB connection
   - Check logs for initialization errors

3. **Team Creation Fails**
   - Verify user role and permissions
   - Check if team slug is unique
   - Ensure platform settings allow team creation

### Debug Commands

```bash
# Check backend logs
tail -f api/apps/runner/backend.log

# Verify MongoDB collections
mongo kloudlite --eval "db.getCollectionNames()"

# Check platform settings
mongo kloudlite --eval "db['platform-settings'].find().pretty()"

# List platform users
mongo kloudlite --eval "db['platform-users'].find().pretty()"
```

## Future Enhancements

1. **Email Notifications**: Send emails for team approval/rejection
2. **Team Management UI**: Interface for managing team members and settings
3. **Audit Logging**: Track all platform administrative actions
4. **Bulk Operations**: Approve/reject multiple team requests
5. **Team Templates**: Pre-configured team settings for common use cases

## Support

For issues or questions:
- Check backend logs for detailed error messages
- Verify all services are running correctly
- Ensure proper environment configuration
- Contact platform support email configured in settings