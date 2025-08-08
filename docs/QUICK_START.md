# Kloudlite v2 - Quick Start Guide

## 🚀 5-Minute Setup

### Prerequisites
```bash
# Required software
- MongoDB 4.4+ (running on localhost:27017)
- NATS 2.10+ (running on localhost:4222)  
- Node.js 18+ with pnpm
- Go 1.21+
- Task (taskfile.dev)
```

### Step 1: Clone & Configure

```bash
# Clone repository
git clone https://github.com/kloudlite/kloudlite-v2
cd kloudlite-v2

# Setup backend environment
cd api/apps/runner
cp .env.example .env

# Edit .env - Set these required values:
# AUTH__PLATFORM_OWNER_EMAIL=your-email@example.com
# AUTH__JWT_SECRET=your-super-secret-key-min-32-chars
# AUTH__MAILTRAP_API_TOKEN=your-mailtrap-token

# Setup frontend environment  
cd ../../../web
cp .env.example .env
# No changes needed for local development
```

### Step 2: Initialize & Run

```bash
# Terminal 1 - Backend
cd api/apps/runner
task init-nats      # One-time setup
task run-local      # Start all services

# Terminal 2 - Frontend
cd web
pnpm install
pnpm dev
```

### Step 3: Access Platform

1. Open http://localhost:3000
2. Check email for password reset link
3. Set password and login
4. You're now the platform super admin! 🎉

## 📁 Project Structure at a Glance

```
kloudlite-v2/
│
├── 🔧 api/                    # Backend (Go)
│   ├── apps/
│   │   ├── auth/             # Login, users, JWT
│   │   ├── accounts/         # Teams, platform
│   │   └── runner/           # Service orchestrator
│   │
│   └── grpc-proto/           # API definitions
│
├── 🎨 web/                    # Frontend (Next.js)
│   ├── app/                  # Pages
│   ├── components/           # React components
│   └── lib/                  # Utilities
│
└── 📚 docs/                   # You are here!
```

## 🔑 Key Concepts

### User Roles Hierarchy

```
┌─────────────────┐
│  Super Admin    │  • Platform owner
│  (You!)         │  • Full control
└────────┬────────┘
         │
┌────────▼────────┐
│     Admin       │  • Approve teams
│                 │  • Manage settings
└────────┬────────┘
         │
┌────────▼────────┐
│     User        │  • Request teams
│                 │  • Use platform
└─────────────────┘
```

### Team Creation Flow

```
Regular User ──► Requests Team ──► Admin Reviews ──► Team Created
                                          │
                                          └────────► Team Rejected

Platform Admin ──► Creates Team ──► Team Ready Immediately
```

## 🛠️ Common Tasks

### Add a Platform Admin

```javascript
// MongoDB shell
use kloudlite
db['platform-users'].insertOne({
  userId: "user-object-id",
  email: "admin@example.com",
  role: "admin",
  isActive: true,
  createdAt: new Date()
})
```

### Create Test User

```bash
# Sign up through UI at /auth/signup
# Or use MongoDB directly:
mongo kloudlite --eval '
db.users.insertOne({
  email: "test@example.com",
  name: "Test User",
  password: "hashed-password",
  verified: true,
  approved: true,
  joined: new Date()
})'
```

### View Logs

```bash
# Backend logs
tail -f api/apps/runner/backend.log

# Frontend logs
# Check terminal where pnpm dev is running

# MongoDB logs
tail -f /usr/local/var/log/mongodb/mongo.log

# NATS logs
nats-server -DV
```

## 🔍 Debugging Tips

### Service Health Checks

```bash
# Check if backend is running
curl http://localhost:50061/health

# Check MongoDB
mongo --eval "db.adminCommand('ping')"

# Check NATS
curl http://localhost:8222/varz

# List platform users
mongo kloudlite --eval "db['platform-users'].find().pretty()"

# View platform settings
mongo kloudlite --eval "db['platform-settings'].findOne()"
```

### Common Issues & Fixes

| Issue | Fix |
|-------|-----|
| "Connection refused :50061" | Run `task run-local` in api/apps/runner |
| "MONGO_URI not set" | Check .env file in api/apps/runner |
| Email not received | Check Mailtrap inbox or set up SMTP |
| "Platform not initialized" | Verify AUTH__PLATFORM_OWNER_EMAIL in .env |
| Frontend build errors | Run `pnpm install` in web directory |

## 🎯 Next Steps

1. **Configure Platform** 
   - Visit `/platform` to update settings
   - Enable OAuth providers
   - Set team approval rules

2. **Invite Team Members**
   - Create teams at `/teams/new`
   - Invite users via email
   - Manage permissions

3. **Explore Features**
   - Device flow at `/device`
   - API documentation
   - Kubernetes integration

## 📊 Development Workflow

### Making Changes

```bash
# Backend changes
1. Edit code in api/apps/
2. Backend auto-reloads
3. Check logs for errors

# Frontend changes  
1. Edit code in web/
2. Next.js hot-reloads
3. See changes instantly

# Proto changes
1. Edit .proto files
2. Regenerate code
3. Update both backend & frontend
```

### Testing Checklist

- [ ] User can sign up
- [ ] Email verification works
- [ ] Password reset works
- [ ] Regular user can request team
- [ ] Admin can approve/reject
- [ ] Team is created correctly
- [ ] User becomes team owner

## 🔗 Useful Links

- **Architecture**: [PLATFORM_ARCHITECTURE.md](./PLATFORM_ARCHITECTURE.md)
- **Workflows**: [TEAM_APPROVAL_WORKFLOW.md](./TEAM_APPROVAL_WORKFLOW.md)
- **API Reference**: See proto files in `/api/grpc-proto/`
- **Components**: Check `/web/components/ui/`

## 💡 Pro Tips

1. **Use tmux** for managing terminals:
   ```bash
   tmux new -s backend -d "cd api/apps/runner && task run-local"
   tmux new -s frontend -d "cd web && pnpm dev"
   ```

2. **MongoDB GUI**: Use MongoDB Compass for visual database exploration

3. **gRPC Testing**: Use grpcurl or Postman for API testing

4. **Hot Reload**: Both backend and frontend support hot reload

5. **Clean Start**:
   ```bash
   # Reset everything
   mongo kloudlite --eval "db.dropDatabase()"
   task init-nats
   task run-local
   ```

---

**Need Help?** 
- Check logs first
- Review error messages carefully  
- Ensure all services are running
- Verify environment configuration

**Happy Coding!** 🚀