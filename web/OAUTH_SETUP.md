# OAuth Provider Setup Guide

## Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google+ API
4. Go to "Credentials" → "Create Credentials" → "OAuth client ID"
5. Choose "Web application"
6. Add authorized redirect URIs:
   - `http://localhost:3000/api/auth/callback/google`
   - `https://yourdomain.com/api/auth/callback/google` (for production)
7. Copy the Client ID and Client Secret

## GitHub OAuth Setup

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click "New OAuth App"
3. Fill in the application details:
   - Application name: Kloudlite
   - Homepage URL: `http://localhost:3000`
   - Authorization callback URL: `http://localhost:3000/api/auth/callback/github`
4. Click "Register application"
5. Copy the Client ID and generate a Client Secret

## Microsoft OAuth Setup

1. Go to [Azure Portal](https://portal.azure.com/)
2. Go to "Azure Active Directory" → "App registrations"
3. Click "New registration"
4. Fill in the application details:
   - Name: Kloudlite
   - Redirect URI: `http://localhost:3000/api/auth/callback/azure-ad`
5. After registration, go to "Certificates & secrets"
6. Create a new client secret
7. Copy the Application (client) ID, client secret, and Directory (tenant) ID

## Environment Variables

Create a `.env.local` file in the web directory with:

```bash
# NextAuth Configuration
NEXTAUTH_URL=http://localhost:3000
NEXTAUTH_SECRET=your-random-secret-key-here

# Google OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret

# GitHub OAuth
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret

# Microsoft OAuth
MICROSOFT_CLIENT_ID=your-microsoft-client-id
MICROSOFT_CLIENT_SECRET=your-microsoft-client-secret
MICROSOFT_TENANT_ID=your-microsoft-tenant-id

# gRPC Server
GRPC_SERVER_ADDRESS=localhost:8080
```

## Generate NextAuth Secret

Run this command to generate a secure secret:
```bash
openssl rand -base64 32
```

## Testing

After setting up the credentials:
1. Start your development server: `pnpm dev`
2. Navigate to `http://localhost:3000/auth/login`
3. Try the social login buttons
4. Check the browser console for any errors
5. Verify that the OAuth flow redirects properly

## Production Setup

For production, make sure to:
1. Update redirect URIs to use your production domain
2. Use environment variables or a secure secrets management system
3. Never commit secrets to version control
4. Consider using different OAuth apps for different environments