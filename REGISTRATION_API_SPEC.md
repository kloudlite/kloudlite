# Registration API Specification

Base URL: `https://onboard.kloudlite.io`

## API Overview

This document describes the complete API specification for Kloudlite deployment registration, IP configuration, and certificate management.

### Available APIs

1. **Verify Installation Key** - `POST /api/register/verify-key`
   - Verify installation key and generate secret key on first call
   - Health check polling endpoint

2. **Uninstall Deployment** - `POST /api/register/uninstall`
   - Complete cleanup of DNS, certificates, and configuration

3. **Configure IP Address** - `POST /api/register/configure-ips`
   - Register installation and workmachine IP addresses
   - Automatic DNS record creation

4. **Generate TLS Certificates** - `POST /api/register/generate-certificates`
   - Generate Cloudflare Origin CA certificates
   - Multi-level support (installation, workmachine, workspace)

5. **Download TLS Certificates** - `GET /api/register/download-certificates`
   - Download certificates in JSON, PEM, or bundle format

## Authentication

All installation/deployment APIs require Bearer token authentication:
```
Authorization: Bearer <secret-key>
```

The secret key is generated when the deployment first verifies itself.

---

## 1. Verify Installation Key

**Endpoint:** `POST /api/register/verify-key`

**Description:** Verify installation key and get user information. Used by deployment during initial setup and for health check polling (every 10 minutes). Generates secret key on first call.

**Authentication:** Not required (this generates the secret key)

**Request Body:**
```json
{
  "installationKey": "uuid-string",
  "markComplete": true  // Optional, legacy support
}
```

**Request Parameters:**
- `installationKey` (required): UUID provided during registration
- `markComplete` (optional): Mark installation as complete (legacy)

**Response (Success - 200):**
```json
{
  "success": true,
  "user": {
    "userId": "github-123456",
    "email": "user@example.com",
    "name": "John Doe",
    "providers": ["github"],
    "registeredAt": "2025-10-23T10:00:00Z",
    "hasCompletedInstallation": true,
    "subdomain": "user-subdomain",
    "domainConfigured": true,
    "url": "https://user-subdomain.kloudlite.io",
    "deploymentReady": false,
    "ipRecords": [],
    "secretKey": "generated-uuid-secret-key",
    "lastHealthCheck": "2025-10-23T13:00:00Z"
  }
}
```

**Key Behaviors:**

1. **First Call (Secret Key Generation):**
   - If `secretKey` doesn't exist, generates a new UUID
   - Marks installation as complete
   - Returns the secret key (store this securely!)

2. **Subsequent Calls (Health Check):**
   - Updates `lastHealthCheck` timestamp
   - Returns current user configuration
   - Includes current `ipRecords` and `deploymentReady` status

3. **Polling:**
   - Deployment should poll this endpoint every 10 minutes
   - Used to monitor configuration changes
   - Updates health check timestamp each time

**Error Responses:**

`400 Bad Request`:
```json
{
  "error": "Installation key is required"
}
```

`404 Not Found`:
```json
{
  "error": "Invalid installation key"
}
```

`500 Internal Server Error`:
```json
{
  "error": "Internal server error"
}
```

---

## 2. Uninstall Deployment

**Endpoint:** `POST /api/register/uninstall`

**Description:** Completely remove installation and cleanup all DNS records, certificates, and configuration. Called by deployment or user to uninstall.

**Authentication:** Required - Bearer token with secret key

**Request Body:**
```json
{
  "installationKey": "uuid-string"
}
```

**Request Parameters:**
- `installationKey` (required): UUID provided during registration

**Cleanup Operations:**

1. **Delete IP Records** - Removes all IP record entries from database
2. **Delete DNS Records** - Removes all DNS records from Cloudflare (subdomain and wildcard)
3. **Revoke Certificates** - Revokes TLS certificates from Cloudflare Origin CA
4. **Delete Domain Reservation** - Frees up the subdomain for reuse
5. **Reset Installation** - Clears installation completion status and deployment ready flag

**Response (Success - 200):**
```json
{
  "success": true,
  "message": "Installation uninstalled successfully",
  "email": "user@example.com",
  "subdomain": "user-subdomain",
  "dnsRecordsDeleted": 4,
  "ipRecordsDeleted": 4,
  "certificatesRevoked": 1
}
```

**Response Fields:**
- `success`: Always true on successful uninstall
- `message`: Confirmation message
- `email`: User email
- `subdomain`: Subdomain that was freed
- `dnsRecordsDeleted`: Count of DNS records removed from Cloudflare
- `ipRecordsDeleted`: Count of IP records removed from database
- `certificatesRevoked`: Count of TLS certificates revoked

**Error Responses:**

`401 Unauthorized`:
```json
{
  "error": "Missing or invalid authorization header"
}
```

`403 Forbidden`:
```json
{
  "error": "Invalid secret key"
}
```

`404 Not Found`:
```json
{
  "error": "Invalid installation key"
}
```

`400 Bad Request`:
```json
{
  "error": "Installation key is required"
}
```

`500 Internal Server Error`:
```json
{
  "error": "Internal server error"
}
```

**Notes:**
- Uninstall is idempotent - safe to call multiple times
- Partial failures in DNS/certificate deletion are logged but don't fail the request
- After uninstall, user can register again with the same account
- Secret key becomes invalid after uninstall

---

## 3. Configure IP Address

**Endpoint:** `POST /api/register/configure-ips`

**Description:** Called by installed deployment to register IP addresses and create DNS records.

**Authentication:** Required - Bearer token with secret key

**Request Body:**
```json
{
  "installationKey": "uuid-string",
  "type": "installation" | "workmachine",
  "ip": "1.2.3.4",
  "workMachineName": "user1"  // Required only for type="workmachine"
}
```

**Request Parameters:**
- `installationKey` (required): UUID provided during registration
- `type` (required): Either "installation" or "workmachine"
- `ip` (required): IPv4 address of the deployment
- `workMachineName` (optional): Name of work machine (required when type="workmachine")

**Response (Success - 200):**
```json
{
  "success": true,
  "type": "installation",
  "ip": "1.2.3.4",
  "workMachineName": "user1",  // Only if type="workmachine"
  "totalRecords": 2,
  "subdomain": "user-subdomain",
  "dnsRecordsCreated": 2,
  "dnsSuccess": true
}
```

**DNS Records Created:**

For `type="installation"`:
- `subdomain.khost.dev` → IP
- `*.subdomain.khost.dev` → IP

For `type="workmachine"`:
- `workmachinename.subdomain.khost.dev` → IP
- `*.workmachinename.subdomain.khost.dev` → IP

**Error Responses:**

`401 Unauthorized`:
```json
{
  "error": "Missing or invalid authorization header"
}
```

`403 Forbidden`:
```json
{
  "error": "Invalid secret key"
}
```

`404 Not Found`:
```json
{
  "error": "Invalid installation key"
}
```

`400 Bad Request`:
```json
{
  "error": "Installation key is required"
}
// OR
{
  "error": "Type must be \"installation\" or \"workmachine\""
}
// OR
{
  "error": "IP address is required"
}
// OR
{
  "error": "workMachineName is required for workmachine type"
}
// OR
{
  "error": "User must have a subdomain assigned before configuring IPs"
}
```

---

## 4. Generate TLS Certificates

**Endpoint:** `POST /api/register/generate-certificates`

**Description:** Generate TLS certificates using Cloudflare Origin CA for HTTPS support.

**Authentication:** Required - Bearer token with secret key

**Request Body:**
```json
{
  "installationKey": "uuid-string",
  "scope": "installation" | "workmachine" | "workspace",  // Optional, defaults to "installation"
  "scopeIdentifier": "dev1",  // Required for workmachine/workspace scopes
  "parentScopeIdentifier": "dev1"  // Required for workspace scope
}
```

**Request Parameters:**
- `installationKey` (required): UUID provided during registration
- `scope` (optional): Certificate scope - "installation", "workmachine", or "workspace" (default: "installation")
- `scopeIdentifier` (conditional):
  - For workmachine: wm-user name
  - For workspace: workspace name
  - Required when scope is not "installation"
- `parentScopeIdentifier` (conditional): wm-user name, required when scope="workspace"

**Certificate Hostnames by Scope:**

**Installation scope:**
- `subdomain.khost.dev`
- `*.subdomain.khost.dev`

**Workmachine scope (scopeIdentifier="dev1"):**
- `dev1.subdomain.khost.dev`
- `*.dev1.subdomain.khost.dev`

**Workspace scope (scopeIdentifier="workspace1", parentScopeIdentifier="dev1"):**
- `workspace1.dev1.subdomain.khost.dev`
- `*.workspace1.dev1.subdomain.khost.dev`

**Response (Success - 200):**
```json
{
  "success": true,
  "certificateId": "cloudflare-cert-id",
  "hostnames": [
    "subdomain.khost.dev",
    "*.subdomain.khost.dev"
  ],
  "scope": "installation",
  "scopeIdentifier": null,
  "parentScopeIdentifier": null,
  "validFrom": "2025-10-23T00:00:00Z",
  "validUntil": "2040-10-23T00:00:00Z",
  "message": "Certificate generated successfully for installation scope"
}
```

**Error Responses:**

`401 Unauthorized`:
```json
{
  "error": "Missing or invalid authorization header"
}
```

`403 Forbidden`:
```json
{
  "error": "Invalid secret key"
}
```

`404 Not Found`:
```json
{
  "error": "Invalid installation key"
}
```

`400 Bad Request`:
```json
{
  "error": "Installation key is required"
}
// OR
{
  "error": "scopeIdentifier (wm-user) is required for workmachine scope"
}
// OR
{
  "error": "scopeIdentifier (workspace) and parentScopeIdentifier (wm-user) are required for workspace scope"
}
// OR
{
  "error": "User must have a subdomain assigned before generating certificates"
}
```

`500 Internal Server Error`:
```json
{
  "error": "Failed to generate certificate"
}
```

---

## 5. Download TLS Certificates

**Endpoint:** `GET /api/register/download-certificates`

**Description:** Download previously generated TLS certificates (certificate + private key).

**Authentication:** Required - Bearer token with secret key

**Query Parameters:**
- `installationKey` (required): UUID provided during registration
- `format` (optional): Response format - "json", "pem", or "bundle" (default: "json")
- `scope` (optional): Filter by certificate scope - "installation", "workmachine", or "workspace"
- `scopeIdentifier` (conditional): Required with scope="workmachine" or scope="workspace"
- `parentScopeIdentifier` (conditional): Required with scope="workspace"

**Examples:**

Get installation certificate:
```
GET /api/register/download-certificates?installationKey=abc-123&format=json
Authorization: Bearer <secret-key>
```

Get workmachine certificate:
```
GET /api/register/download-certificates?installationKey=abc-123&scope=workmachine&scopeIdentifier=dev1
Authorization: Bearer <secret-key>
```

Get workspace certificate:
```
GET /api/register/download-certificates?installationKey=abc-123&scope=workspace&scopeIdentifier=workspace1&parentScopeIdentifier=dev1
Authorization: Bearer <secret-key>
```

**Response Formats:**

**format="json" (default):**
```json
{
  "success": true,
  "certificate": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
  "privateKey": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
  "hostnames": [
    "subdomain.khost.dev",
    "*.subdomain.khost.dev"
  ],
  "scope": "installation",
  "scopeIdentifier": null,
  "parentScopeIdentifier": null,
  "validFrom": "2025-10-23T00:00:00Z",
  "validUntil": "2040-10-23T00:00:00Z",
  "cloudflareCertId": "cloudflare-cert-id"
}
```

**format="pem":**
```json
{
  "certificate": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
  "privateKey": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
  "hostnames": ["subdomain.khost.dev", "*.subdomain.khost.dev"],
  "scope": "installation",
  "scopeIdentifier": null,
  "parentScopeIdentifier": null,
  "validFrom": "2025-10-23T00:00:00Z",
  "validUntil": "2040-10-23T00:00:00Z"
}
```

**format="bundle":**
Returns raw text file with certificate and key concatenated:
```
Content-Type: text/plain
Content-Disposition: attachment; filename="subdomain-tls-bundle.pem"

-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
-----BEGIN PRIVATE KEY-----
...
-----END PRIVATE KEY-----
```

**Error Responses:**

`401 Unauthorized`:
```json
{
  "error": "Missing or invalid authorization header"
}
```

`403 Forbidden`:
```json
{
  "error": "Invalid secret key"
}
```

`404 Not Found`:
```json
{
  "error": "Invalid installation key"
}
// OR
{
  "error": "No certificate found for this installation"
}
// OR (with scope)
{
  "error": "No certificate found for this installation for workmachine scope"
}
```

`400 Bad Request`:
```json
{
  "error": "Installation key is required"
}
// OR
{
  "error": "scopeIdentifier (wm-user) is required for workmachine scope"
}
// OR
{
  "error": "scopeIdentifier (workspace) and parentScopeIdentifier (wm-user) are required for workspace scope"
}
```

---

## Installation Flow

### Step 1: User Registration
User registers via OAuth (GitHub/Google/Microsoft Entra) at `https://onboard.kloudlite.io/register`.

After successful OAuth:
- User receives `installationKey` (UUID)
- No `secretKey` yet (generated on first verification)

### Step 2: Deployment Installation
User deploys Kloudlite using the `installationKey`.

### Step 3: Verify Installation Key (Get Secret Key)
Deployment calls `/api/register/verify-key` on first startup:

```bash
curl -X POST https://onboard.kloudlite.io/api/register/verify-key \
  -H "Content-Type: application/json" \
  -d '{
    "installationKey": "abc-123"
  }'
```

**Response includes:**
- `secretKey`: Newly generated UUID (store this securely!)
- `subdomain`: Assigned subdomain (if available)
- User information and configuration

**Important:** Save the `secretKey` - it's required for all subsequent API calls.

### Step 4: Configure IP Address
Deployment calls `/api/register/configure-ips` for the first time:

```bash
curl -X POST https://onboard.kloudlite.io/api/register/configure-ips \
  -H "Authorization: Bearer <secret-key-if-available>" \
  -H "Content-Type: application/json" \
  -d '{
    "installationKey": "abc-123",
    "type": "installation",
    "ip": "203.0.113.10"
  }'
```

**DNS records created:**
- `subdomain.khost.dev` → IP
- `*.subdomain.khost.dev` → IP

**Deployment marked as ready.**

### Step 5: Certificate Generation
Deployment generates TLS certificates:

```bash
curl -X POST https://onboard.kloudlite.io/api/register/generate-certificates \
  -H "Authorization: Bearer <secret-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "installationKey": "abc-123"
  }'
```

Response includes certificate ID and validity period.

### Step 6: Certificate Download
Deployment downloads certificates:

```bash
curl -X GET 'https://onboard.kloudlite.io/api/register/download-certificates?installationKey=abc-123&format=bundle' \
  -H "Authorization: Bearer <secret-key>" \
  -o tls-bundle.pem
```

### Step 7: Configure Additional IPs (Optional)
For workmachines:

```bash
curl -X POST https://onboard.kloudlite.io/api/register/configure-ips \
  -H "Authorization: Bearer <secret-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "installationKey": "abc-123",
    "type": "workmachine",
    "ip": "203.0.113.20",
    "workMachineName": "dev1"
  }'
```

Creates DNS: `dev1.subdomain.khost.dev` → IP

---

## Security Notes

1. **Secret Key Generation:** Secret key is automatically generated on first IP configuration call
2. **Secret Key Storage:** Store secret key securely - it's required for all subsequent API calls
3. **Bearer Token:** Always use `Authorization: Bearer <secret-key>` header
4. **HTTPS Only:** All API calls must use HTTPS
5. **Rate Limiting:** No explicit rate limits currently, but DNS operations may take time
6. **Certificate Validity:** Cloudflare Origin CA certificates are valid for 15 years

---

## Common Error Scenarios

### Missing Authorization Header
```bash
# Request without Authorization header
curl -X POST https://onboard.kloudlite.io/api/register/configure-ips \
  -H "Content-Type: application/json" \
  -d '{"installationKey": "abc-123", "type": "installation", "ip": "1.2.3.4"}'

# Response: 401 Unauthorized
{
  "error": "Missing or invalid authorization header"
}
```

### Invalid Secret Key
```bash
curl -X POST https://onboard.kloudlite.io/api/register/configure-ips \
  -H "Authorization: Bearer wrong-secret" \
  -H "Content-Type: application/json" \
  -d '{"installationKey": "abc-123", "type": "installation", "ip": "1.2.3.4"}'

# Response: 403 Forbidden
{
  "error": "Invalid secret key"
}
```

### Invalid Installation Key
```bash
curl -X POST https://onboard.kloudlite.io/api/register/configure-ips \
  -H "Authorization: Bearer <secret-key>" \
  -H "Content-Type: application/json" \
  -d '{"installationKey": "invalid-key", "type": "installation", "ip": "1.2.3.4"}'

# Response: 404 Not Found
{
  "error": "Invalid installation key"
}
```

---

## Testing Examples

### Complete Installation Flow

```bash
# 1. User registers via web UI and gets installationKey
INSTALLATION_KEY="your-installation-key-here"

# 2. Verify installation key and get secret key
VERIFY_RESPONSE=$(curl -X POST https://onboard.kloudlite.io/api/register/verify-key \
  -H "Content-Type: application/json" \
  -d "{
    \"installationKey\": \"$INSTALLATION_KEY\"
  }")

# Extract secret key and subdomain
SECRET_KEY=$(echo $VERIFY_RESPONSE | jq -r '.user.secretKey')
SUBDOMAIN=$(echo $VERIFY_RESPONSE | jq -r '.user.subdomain')

echo "Secret Key: $SECRET_KEY"
echo "Subdomain: $SUBDOMAIN"

# 3. Configure installation IP
curl -X POST https://onboard.kloudlite.io/api/register/configure-ips \
  -H "Authorization: Bearer $SECRET_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"installationKey\": \"$INSTALLATION_KEY\",
    \"type\": \"installation\",
    \"ip\": \"203.0.113.10\"
  }"

# 4. Generate certificate
curl -X POST https://onboard.kloudlite.io/api/register/generate-certificates \
  -H "Authorization: Bearer $SECRET_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"installationKey\": \"$INSTALLATION_KEY\"}"

# 5. Download certificate bundle
curl -X GET "https://onboard.kloudlite.io/api/register/download-certificates?installationKey=$INSTALLATION_KEY&format=bundle" \
  -H "Authorization: Bearer $SECRET_KEY" \
  -o "$SUBDOMAIN-tls-bundle.pem"

# 6. Configure workmachine IP
curl -X POST https://onboard.kloudlite.io/api/register/configure-ips \
  -H "Authorization: Bearer $SECRET_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"installationKey\": \"$INSTALLATION_KEY\",
    \"type\": \"workmachine\",
    \"ip\": \"203.0.113.20\",
    \"workMachineName\": \"dev1\"
  }"

# 7. Generate workmachine certificate
curl -X POST https://onboard.kloudlite.io/api/register/generate-certificates \
  -H "Authorization: Bearer $SECRET_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"installationKey\": \"$INSTALLATION_KEY\",
    \"scope\": \"workmachine\",
    \"scopeIdentifier\": \"dev1\"
  }"

echo "Installation complete!"
echo "Your domains:"
echo "  - https://$SUBDOMAIN.khost.dev"
echo "  - https://dev1.$SUBDOMAIN.khost.dev"
```
