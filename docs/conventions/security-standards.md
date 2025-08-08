# Security Standards

This document outlines the security conventions and best practices for Kloudlite v2.

## 🔐 Authentication

### JWT Token Management
```go
// Token structure
type TokenClaims struct {
    UserID    string   `json:"userId"`
    Email     string   `json:"email"`
    Roles     []string `json:"roles"`
    SessionID string   `json:"sessionId"`
    jwt.StandardClaims
}

// Token generation
func GenerateTokens(user *User) (accessToken, refreshToken string, err error) {
    // Access token - short lived (15 minutes)
    accessClaims := TokenClaims{
        UserID: user.ID,
        Email:  user.Email,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
            IssuedAt:  time.Now().Unix(),
            Issuer:    "kloudlite.io",
        },
    }
    
    // Refresh token - long lived (7 days)
    refreshClaims := TokenClaims{
        UserID:    user.ID,
        SessionID: generateSessionID(),
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
        },
    }
    
    // Sign with different keys
    accessToken = jwt.SignWithHS256(accessClaims, accessSecret)
    refreshToken = jwt.SignWithHS256(refreshClaims, refreshSecret)
    
    return
}
```

### Session Management
```typescript
// Frontend session handling
export const authOptions: NextAuthOptions = {
  session: {
    strategy: 'jwt',
    maxAge: 30 * 24 * 60 * 60, // 30 days
  },
  jwt: {
    secret: process.env.NEXTAUTH_SECRET,
    encryption: true,
  },
  callbacks: {
    async jwt({ token, user, account }) {
      // Rotate refresh token
      if (shouldRotateRefreshToken(token)) {
        const newTokens = await rotateTokens(token.refreshToken)
        return { ...token, ...newTokens }
      }
      return token
    },
  },
}
```

### Password Requirements
```go
// Password validation
func ValidatePassword(password string) error {
    if len(password) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    
    var (
        hasUpper   = regexp.MustCompile(`[A-Z]`).MatchString
        hasLower   = regexp.MustCompile(`[a-z]`).MatchString
        hasNumber  = regexp.MustCompile(`[0-9]`).MatchString
        hasSpecial = regexp.MustCompile(`[!@#$%^&*]`).MatchString
    )
    
    if !hasUpper(password) || !hasLower(password) {
        return errors.New("password must contain uppercase and lowercase letters")
    }
    
    if !hasNumber(password) {
        return errors.New("password must contain at least one number")
    }
    
    if !hasSpecial(password) {
        return errors.New("password must contain at least one special character")
    }
    
    return nil
}

// Password hashing
func HashPassword(password string) (string, error) {
    // Use bcrypt with cost factor 12
    hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    return string(hash), err
}
```

## 🛡️ Authorization

### Role-Based Access Control (RBAC)
```go
// Permission checking
type Permission string

const (
    // Platform permissions
    PermManagePlatform Permission = "platform:manage"
    PermCreateTeams    Permission = "platform:teams:create"
    
    // Team permissions
    PermViewTeam       Permission = "team:view"
    PermManageTeam     Permission = "team:manage"
    PermManageMembers  Permission = "team:members:manage"
)

// Role definitions
var rolePermissions = map[Role][]Permission{
    RolePlatformSuperAdmin: {
        PermManagePlatform,
        PermCreateTeams,
        // All permissions
    },
    RoleTeamOwner: {
        PermViewTeam,
        PermManageTeam,
        PermManageMembers,
    },
    RoleTeamMember: {
        PermViewTeam,
    },
}

// Authorization middleware
func RequirePermission(perm Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := GetUser(c)
        if !HasPermission(user, perm) {
            c.AbortWithStatus(http.StatusForbidden)
            return
        }
        c.Next()
    }
}
```

### Resource-Level Security
```go
// Check team access
func (s *service) CanAccessTeam(ctx context.Context, userId, teamId string) bool {
    membership, err := s.GetTeamMembership(ctx, userId, teamId)
    if err != nil {
        return false
    }
    
    return membership != nil && membership.Status == "active"
}

// Secure query filters
func (s *service) ListTeams(ctx UserContext) ([]*Team, error) {
    // Only return teams user has access to
    filter := bson.M{
        "members.userId": ctx.GetUserId(),
        "members.status": "active",
    }
    
    return s.repo.Find(ctx, filter)
}
```

## 🔒 Data Security

### Encryption at Rest
```go
// Sensitive field encryption
type EncryptedField struct {
    Data string `json:"data" bson:"data"`
    IV   string `json:"iv" bson:"iv"`
}

func EncryptField(plaintext string, key []byte) (*EncryptedField, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    
    // Generate IV
    iv := make([]byte, aes.BlockSize)
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return nil, err
    }
    
    // Encrypt
    stream := cipher.NewCFBEncrypter(block, iv)
    ciphertext := make([]byte, len(plaintext))
    stream.XORKeyStream(ciphertext, []byte(plaintext))
    
    return &EncryptedField{
        Data: base64.StdEncoding.EncodeToString(ciphertext),
        IV:   base64.StdEncoding.EncodeToString(iv),
    }, nil
}
```

### Sensitive Data Handling
```go
// Never log sensitive data
type User struct {
    Email    string `json:"email"`
    Password string `json:"-"` // Never serialize
    APIKey   string `json:"-"` // Never expose
}

// Audit logging
func (s *service) LogAccess(ctx context.Context, resource, action string) {
    s.logger.Info("resource accessed",
        "userId", GetUserId(ctx),
        "resource", resource,
        "action", action,
        "ip", GetClientIP(ctx),
        // Never log sensitive data
    )
}
```

## 🌐 API Security

### Input Validation
```go
// Validate all inputs
func ValidateCreateTeamRequest(req *CreateTeamRequest) error {
    // Sanitize inputs
    req.Slug = strings.TrimSpace(strings.ToLower(req.Slug))
    req.DisplayName = strings.TrimSpace(req.DisplayName)
    
    // Validate slug
    if !slugRegex.MatchString(req.Slug) {
        return errors.New("invalid slug format")
    }
    
    // Prevent injection
    if containsSQLKeywords(req.DisplayName) {
        return errors.New("invalid characters in display name")
    }
    
    // Length limits
    if len(req.Description) > 1000 {
        return errors.New("description too long")
    }
    
    return nil
}

// XSS prevention
func SanitizeHTML(input string) string {
    p := bluemonday.StrictPolicy()
    return p.Sanitize(input)
}
```

### Rate Limiting
```go
// API rate limiting
func RateLimitMiddleware() gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Every(time.Second), 10) // 10 req/sec
    
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.AbortWithStatusJSON(429, gin.H{
                "error": "rate limit exceeded",
            })
            return
        }
        c.Next()
    }
}

// Per-user rate limiting
var userLimiters = &sync.Map{}

func PerUserRateLimit(userId string) bool {
    val, _ := userLimiters.LoadOrStore(userId, rate.NewLimiter(rate.Every(time.Minute), 60))
    limiter := val.(*rate.Limiter)
    return limiter.Allow()
}
```

### CORS Configuration
```typescript
// Frontend CORS setup
export async function middleware(request: NextRequest) {
  const response = NextResponse.next()
  
  // Set security headers
  response.headers.set('X-Frame-Options', 'DENY')
  response.headers.set('X-Content-Type-Options', 'nosniff')
  response.headers.set('X-XSS-Protection', '1; mode=block')
  response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin')
  
  // CORS for API routes
  if (request.nextUrl.pathname.startsWith('/api/')) {
    response.headers.set('Access-Control-Allow-Origin', process.env.ALLOWED_ORIGIN || '*')
    response.headers.set('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS')
    response.headers.set('Access-Control-Allow-Headers', 'Content-Type, Authorization')
  }
  
  return response
}
```

## 🔑 Secret Management

### Environment Variables
```bash
# .env.example (commit this)
DATABASE_URL=mongodb://localhost:27017/kloudlite
JWT_SECRET=your-secret-here
ENCRYPTION_KEY=your-key-here
OAUTH_CLIENT_SECRET=your-secret-here

# .env (never commit)
DATABASE_URL=mongodb://prod-server:27017/kloudlite
JWT_SECRET=actual-secret-value
ENCRYPTION_KEY=actual-key-value
OAUTH_CLIENT_SECRET=actual-secret-value
```

### Secret Rotation
```go
// Implement key rotation
type KeyRotation struct {
    CurrentKey  string
    PreviousKey string
    RotatedAt   time.Time
}

func (s *service) RotateEncryptionKey() error {
    newKey := generateSecureKey()
    
    // Store old key for decryption
    s.keyRotation = KeyRotation{
        CurrentKey:  newKey,
        PreviousKey: s.currentKey,
        RotatedAt:   time.Now(),
    }
    
    // Re-encrypt sensitive data in background
    go s.reencryptData()
    
    return nil
}
```

## 🚨 Security Headers

### Backend Security Headers
```go
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Header("Content-Security-Policy", "default-src 'self'")
        c.Next()
    }
}
```

### Frontend Security Headers
```typescript
// next.config.ts
export default {
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          {
            key: 'X-DNS-Prefetch-Control',
            value: 'on'
          },
          {
            key: 'X-Frame-Options',
            value: 'SAMEORIGIN'
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff'
          },
          {
            key: 'Referrer-Policy',
            value: 'origin-when-cross-origin'
          },
        ],
      },
    ]
  },
}
```

## 🔍 Security Monitoring

### Audit Logging
```go
type AuditLog struct {
    ID          string    `bson:"_id"`
    UserID      string    `bson:"userId"`
    Action      string    `bson:"action"`
    Resource    string    `bson:"resource"`
    ResourceID  string    `bson:"resourceId"`
    IP          string    `bson:"ip"`
    UserAgent   string    `bson:"userAgent"`
    Success     bool      `bson:"success"`
    Error       string    `bson:"error,omitempty"`
    Timestamp   time.Time `bson:"timestamp"`
}

func (s *service) LogSecurityEvent(ctx context.Context, event AuditLog) {
    // Store in separate audit collection
    s.auditRepo.Create(ctx, &event)
    
    // Alert on suspicious activity
    if s.isSuspicious(event) {
        s.alertSecurityTeam(event)
    }
}
```

### Intrusion Detection
```go
// Detect suspicious patterns
func (s *service) DetectAnomalies(userId string) bool {
    // Check failed login attempts
    failedAttempts := s.getFailedLoginCount(userId, time.Hour)
    if failedAttempts > 5 {
        s.lockAccount(userId)
        return true
    }
    
    // Check unusual access patterns
    locations := s.getAccessLocations(userId, 24*time.Hour)
    if s.hasGeographicAnomaly(locations) {
        s.requireAdditionalAuth(userId)
        return true
    }
    
    return false
}
```

## 📋 Security Checklist

### Development
- [ ] Validate all inputs
- [ ] Sanitize outputs
- [ ] Use parameterized queries
- [ ] Implement proper error handling
- [ ] Never log sensitive data
- [ ] Use secure random generators
- [ ] Implement rate limiting
- [ ] Add security headers

### Deployment
- [ ] Enable HTTPS everywhere
- [ ] Set secure cookie flags
- [ ] Configure firewall rules
- [ ] Enable audit logging
- [ ] Set up monitoring alerts
- [ ] Regular security updates
- [ ] Implement backup encryption
- [ ] Document security procedures

### Code Review
- [ ] Check for SQL/NoSQL injection
- [ ] Verify authentication checks
- [ ] Review authorization logic
- [ ] Check for XSS vulnerabilities
- [ ] Verify CSRF protection
- [ ] Review cryptographic usage
- [ ] Check secret management
- [ ] Verify input validation

## 🚀 Best Practices

1. **Defense in Depth** - Multiple security layers
2. **Principle of Least Privilege** - Minimal access rights
3. **Zero Trust** - Verify everything
4. **Fail Secure** - Deny by default
5. **Security by Design** - Built-in, not added on
6. **Regular Updates** - Keep dependencies current
7. **Security Training** - Educate the team
8. **Incident Response** - Have a plan ready